package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/shivam-mishra-20/mak-watches-be/internal/config"
	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"github.com/shivam-mishra-20/mak-watches-be/internal/middleware"
	"github.com/shivam-mishra-20/mak-watches-be/internal/models"
)

// OrderHandler handles order related requests
type OrderHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(db *database.DBClient, cfg *config.Config) *OrderHandler {
	return &OrderHandler{
		DB:     db,
		Config: cfg,
	}
}

// Checkout processes the checkout and creates an order
func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Parse request body
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate request
	if req.ShippingAddress.Street == "" || req.ShippingAddress.City == "" ||
		req.ShippingAddress.State == "" || req.ShippingAddress.ZipCode == "" ||
		req.ShippingAddress.Country == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Complete shipping address is required",
		})
	}

	if req.PaymentInfo.Method == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment method is required",
		})
	}

	// Get the user's cart
	cartCollection := h.DB.Collections().CartItems
	cursor, err := cartCollection.Find(ctx, bson.M{"user_id": user.UserID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve cart",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Parse cart items
	var cartItems []models.CartItem
	if err := cursor.All(ctx, &cartItems); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode cart items",
			"error":   err.Error(),
		})
	}

	// Check if cart is empty
	if len(cartItems) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Cart is empty",
		})
	}

	// Create order items and calculate total (authoritative server-side)
	var orderItems []models.OrderItem
	var total float64
	productsCollection := h.DB.Collections().Products

	for _, item := range cartItems {
		// Get product details
		var product models.Product
		err := productsCollection.FindOne(ctx, bson.M{"_id": item.ProductID}).Decode(&product)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to retrieve product details",
				"error":   err.Error(),
			})
		}

		// Check if there's enough stock
		if product.Stock < item.Quantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("Not enough stock for product %s", product.Name),
			})
		}

		// Create order item
		orderItem := models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Size:        item.Size,
			Quantity:    item.Quantity,
			Subtotal:    product.Price * float64(item.Quantity),
		}

		orderItems = append(orderItems, orderItem)
		total += orderItem.Subtotal

		// Update product stock
		_, err = productsCollection.UpdateOne(
			ctx,
			bson.M{"_id": product.ID},
			bson.M{"$inc": bson.M{"stock": -item.Quantity}},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update product stock",
				"error":   err.Error(),
			})
		}

		// Invalidate product cache
		productCacheKey := fmt.Sprintf("product:%s", product.ID.Hex())
		h.DB.CacheDel(ctx, productCacheKey)
	}

	// Verify Razorpay signature if method is razorpay
	if req.PaymentInfo.Method == "razorpay" {
		if req.PaymentInfo.RazorpayOrderID == "" || req.PaymentInfo.RazorpayPaymentID == "" || req.PaymentInfo.RazorpaySignature == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Missing Razorpay payment details"})
		}
		mac := hmac.New(sha256.New, []byte(h.Config.RazorpaySecret))
		mac.Write([]byte(req.PaymentInfo.RazorpayOrderID + "|" + req.PaymentInfo.RazorpayPaymentID))
		expected := hex.EncodeToString(mac.Sum(nil))
		if expected != req.PaymentInfo.RazorpaySignature {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid payment signature"})
		}
	}

	// Defensive: If client supplied a clientTotal ensure it matches authoritative total
	if req.ClientTotal != nil {
		clientTotal := *req.ClientTotal
		// Allow small rounding difference (₹1)
		if clientTotal < total-1 || clientTotal > total+1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("Total mismatch. Client: %.2f Server: %.2f", clientTotal, total),
			})
		}
	}

	// Determine order and payment statuses
	orderStatus := "pending"  // pending -> processing -> shipped -> delivered/cancelled/returned
	paymentStatus := "unpaid" // unpaid | paid | refunded | failed
	switch req.PaymentInfo.Method {
	case "razorpay":
		// Signature already verified above, consider payment successful
		paymentStatus = "paid"
		orderStatus = "processing"
	case "cod":
		paymentStatus = "unpaid"
		orderStatus = "processing"
	}

	// Create the order
	now := time.Now()
	order := models.Order{
		ID:              primitive.NewObjectID(),
		UserID:          user.UserID,
		Items:           orderItems,
		Total:           total,
		Status:          orderStatus,
		PaymentStatus:   paymentStatus,
		ShippingAddress: req.ShippingAddress,
		PaymentInfo:     req.PaymentInfo,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Insert the order into the database
	orderCollection := h.DB.Collections().Orders
	_, err = orderCollection.InsertOne(ctx, order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create order",
			"error":   err.Error(),
		})
	}

	// Clear the user's cart
	_, err = cartCollection.DeleteMany(ctx, bson.M{"user_id": user.UserID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to clear cart after order",
			"error":   err.Error(),
		})
	}

	// Invalidate cart cache
	cartCacheKey := fmt.Sprintf("cart:%s", user.UserID.Hex())
	h.DB.CacheDel(ctx, cartCacheKey)

	// Invalidate order cache
	ordersCacheKey := fmt.Sprintf("orders:%s", user.UserID.Hex())
	h.DB.CacheDel(ctx, ordersCacheKey)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Order placed successfully",
		"data":    order,
	})
}

// GetOrders retrieves order history for a user
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
	ctx := c.Context()

	// Determine the target user ID from route params or the authenticated token
	tokenUser, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	userIDParam := c.Params("userID")
	var userID primitive.ObjectID
	var err error
	if userIDParam == "" {
		// If no param provided (e.g., /account/orders), default to the authenticated user's ID
		userID = tokenUser.UserID
	} else {
		// Convert user ID from string to ObjectID
		userID, err = primitive.ObjectIDFromHex(userIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format",
				"error":   err.Error(),
			})
		}
	}

	// Authorization: user can view own orders; admin can view any user's orders
	if tokenUser.UserID != userID && tokenUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Not authorized to view these orders",
		})
	}

	// Check if the orders are in Redis cache
	cacheKey := fmt.Sprintf("orders:%s", userID.Hex())
	var orders []models.Order
	err = h.DB.CacheGet(ctx, cacheKey, &orders)
	if err == nil {
		// Cache hit
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Orders retrieved from cache",
			"data":    orders,
		})
	}

	// Find all orders for the user, sorted by creation date descending
	orderCollection := h.DB.Collections().Orders
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := orderCollection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve orders",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Parse the results
	if err := cursor.All(ctx, &orders); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode orders",
			"error":   err.Error(),
		})
	}

	// Map orders to convert ObjectID to hex string for frontend
	type OrderResponse struct {
		ID              string             `json:"id"`
		UserID          string             `json:"userId"`
		Items           []models.OrderItem `json:"items"`
		Total           float64            `json:"total"`
		Status          string             `json:"status"`
		PaymentStatus   string             `json:"paymentStatus"`
		ShippingAddress models.Address     `json:"shippingAddress"`
		PaymentInfo     models.PaymentInfo `json:"paymentInfo"`
		CreatedAt       time.Time          `json:"createdAt"`
		UpdatedAt       time.Time          `json:"updatedAt"`
	}
	var respOrders []OrderResponse
	for _, o := range orders {
		payStatus := o.PaymentStatus
		if payStatus == "" {
			if o.Status == "paid" || o.PaymentInfo.RazorpayPaymentID != "" {
				payStatus = "paid"
			} else if o.Status == "cancelled" {
				payStatus = "refunded"
			} else {
				payStatus = "unpaid"
			}
		}
		respOrders = append(respOrders, OrderResponse{
			ID:              o.ID.Hex(),
			UserID:          o.UserID.Hex(),
			Items:           o.Items,
			Total:           o.Total,
			Status:          o.Status,
			PaymentStatus:   payStatus,
			ShippingAddress: o.ShippingAddress,
			PaymentInfo:     o.PaymentInfo,
			CreatedAt:       o.CreatedAt,
			UpdatedAt:       o.UpdatedAt,
		})
	}

	// Cache the orders (expire after 15 minutes)
	h.DB.CacheSet(ctx, cacheKey, respOrders, 15*time.Minute)

	// Return the orders
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Orders retrieved successfully",
		"data":    respOrders,
	})
}

// GetOrder retrieves a specific order by ID
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get order ID from URL parameter
	orderIDParam := c.Params("orderID")
	if orderIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Order ID is required",
		})
	}

	// Convert order ID from string to ObjectID
	orderID, err := primitive.ObjectIDFromHex(orderIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid order ID format",
			"error":   err.Error(),
		})
	}

	// Check if the order is in Redis cache
	cacheKey := fmt.Sprintf("order:%s", orderID.Hex())
	var order models.Order
	err = h.DB.CacheGet(ctx, cacheKey, &order)
	if err == nil {
		// Cache hit
		// Check if the user is authorized to view this order
		tokenUser, ok := c.Locals("user").(*middleware.TokenMetadata)
		if !ok || (order.UserID != tokenUser.UserID && tokenUser.Role != "admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Not authorized to view this order",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Order retrieved from cache",
			"data":    order,
		})
	}

	// Find the order in the database
	orderCollection := h.DB.Collections().Orders
	err = orderCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve order",
			"error":   err.Error(),
		})
	}

	// Check if the user is authorized to view this order
	tokenUser, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok || (order.UserID != tokenUser.UserID && tokenUser.Role != "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Not authorized to view this order",
		})
	}

	// Cache the order (expire after 15 minutes)
	h.DB.CacheSet(ctx, cacheKey, order, 15*time.Minute)

	// Return the order
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Order retrieved successfully",
		"data":    order,
	})
}

// UpdateOrderStatus updates the status of an order (admin only)
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	ctx := c.Context()

	// Only admin can update order status
	tokenUser, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok || tokenUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only admins can update order status",
		})
	}

	// Get order ID from URL parameter
	orderIDParam := c.Params("orderID")
	if orderIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Order ID is required",
		})
	}

	// Convert order ID from string to ObjectID
	orderID, err := primitive.ObjectIDFromHex(orderIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid order ID format",
			"error":   err.Error(),
		})
	}

	// Parse request body
	type StatusUpdate struct {
		Status        string `json:"status"`
		PaymentStatus string `json:"paymentStatus,omitempty"`
	}
	var req StatusUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate statuses
	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"shipped":    true,
		"delivered":  true,
		"cancelled":  true,
		"returned":   true,
	}

	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid order status. Must be one of: pending, processing, shipped, delivered, cancelled, returned",
		})
	}

	validPaymentStatuses := map[string]bool{
		"unpaid":   true,
		"paid":     true,
		"failed":   true,
		"refunded": true,
	}
	if req.PaymentStatus != "" && !validPaymentStatuses[req.PaymentStatus] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid payment status. Must be one of: unpaid, paid, failed, refunded",
		})
	}

	// Update the order status
	now := time.Now()
	orderCollection := h.DB.Collections().Orders
	setFields := bson.M{
		"status":     req.Status,
		"updated_at": now,
	}
	if req.PaymentStatus != "" {
		setFields["payment_status"] = req.PaymentStatus
	}
	result, err := orderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{"$set": setFields},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update order status",
			"error":   err.Error(),
		})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Order not found",
		})
	}

	// Get the updated order
	var updatedOrder models.Order
	err = orderCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&updatedOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve updated order",
			"error":   err.Error(),
		})
	}

	// Invalidate order caches
	orderCacheKey := fmt.Sprintf("order:%s", orderID.Hex())
	userOrdersCacheKey := fmt.Sprintf("orders:%s", updatedOrder.UserID.Hex())
	h.DB.CacheDel(ctx, orderCacheKey)
	h.DB.CacheDel(ctx, userOrdersCacheKey)

	// Return the updated order
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Order status updated successfully",
		"data":    updatedOrder,
	})
}

// CancelOrder cancels an order if it's still in "pending" or "processing" status
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get order ID from URL parameter
	orderIDParam := c.Params("orderID")
	if orderIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Order ID is required",
		})
	}

	// Convert order ID from string to ObjectID
	orderID, err := primitive.ObjectIDFromHex(orderIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid order ID format",
			"error":   err.Error(),
		})
	}

	// Get the order
	orderCollection := h.DB.Collections().Orders
	var order models.Order
	err = orderCollection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve order",
			"error":   err.Error(),
		})
	}

	// Check if the user is authorized to cancel this order
	tokenUser, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok || (order.UserID != tokenUser.UserID && tokenUser.Role != "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Not authorized to cancel this order",
		})
	}

	// Check if the order can be cancelled
	if order.Status != "pending" && order.Status != "processing" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Only pending or processing orders can be cancelled",
		})
	}

	// Update the order status to "cancelled" and set paymentStatus if prepaid
	now := time.Now()
	setCancel := bson.M{
		"status":     "cancelled",
		"updated_at": now,
	}
	if order.PaymentStatus == "paid" {
		// Business rule: mark as refunded; real refund should be processed via gateway
		setCancel["payment_status"] = "refunded"
	}
	_, err = orderCollection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{"$set": setCancel},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to cancel order",
			"error":   err.Error(),
		})
	}

	// Return inventory to stock
	productsCollection := h.DB.Collections().Products
	for _, item := range order.Items {
		_, err = productsCollection.UpdateOne(
			ctx,
			bson.M{"_id": item.ProductID},
			bson.M{"$inc": bson.M{"stock": item.Quantity}},
		)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Error restoring inventory for product %s: %v\n", item.ProductID.Hex(), err)
		}

		// Invalidate product cache
		productCacheKey := fmt.Sprintf("product:%s", item.ProductID.Hex())
		h.DB.CacheDel(ctx, productCacheKey)
	}

	// Invalidate order caches
	orderCacheKey := fmt.Sprintf("order:%s", orderID.Hex())
	userOrdersCacheKey := fmt.Sprintf("orders:%s", order.UserID.Hex())
	h.DB.CacheDel(ctx, orderCacheKey)
	h.DB.CacheDel(ctx, userOrdersCacheKey)

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Order cancelled successfully",
	})
}

// GetAllOrders returns all orders (admin only)
func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	ctx := c.Context()
	// Only admin can access
	tokenUser, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok || tokenUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Not authorized",
		})
	}
	orderCollection := h.DB.Collections().Orders
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := orderCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve orders",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)
	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode orders",
			"error":   err.Error(),
		})
	}
	// Map orders to frontend format if needed
	type OrderResponse struct {
		ID              string             `json:"id"`
		UserID          string             `json:"userId"`
		CustomerName    string             `json:"customerName"`
		Items           []models.OrderItem `json:"items"`
		Total           float64            `json:"total"`
		Status          string             `json:"status"`
		PaymentStatus   string             `json:"paymentStatus"`
		ShippingAddress models.Address     `json:"shippingAddress"`
		PaymentInfo     models.PaymentInfo `json:"paymentInfo"`
		CreatedAt       time.Time          `json:"createdAt"`
		UpdatedAt       time.Time          `json:"updatedAt"`
	}
	userCollection := h.DB.Collections().Users
	// Cache userId to name to avoid duplicate DB calls
	userNameCache := make(map[string]string)
	var respOrders []OrderResponse
	for _, o := range orders {
		payStatus := o.PaymentStatus
		if payStatus == "" {
			if o.Status == "paid" || o.PaymentInfo.RazorpayPaymentID != "" {
				payStatus = "paid"
			} else if o.Status == "cancelled" {
				payStatus = "refunded"
			} else {
				payStatus = "unpaid"
			}
		}
		userIdStr := o.UserID.Hex()
		customerName := ""
		if cached, ok := userNameCache[userIdStr]; ok {
			customerName = cached
		} else {
			var user models.User
			err := userCollection.FindOne(ctx, bson.M{"_id": o.UserID}).Decode(&user)
			if err == nil {
				customerName = user.Name
			}
			userNameCache[userIdStr] = customerName
		}
		respOrders = append(respOrders, OrderResponse{
			ID:              o.ID.Hex(),
			UserID:          userIdStr,
			CustomerName:    customerName,
			Items:           o.Items,
			Total:           o.Total,
			Status:          o.Status,
			PaymentStatus:   payStatus,
			ShippingAddress: o.ShippingAddress,
			PaymentInfo:     o.PaymentInfo,
			CreatedAt:       o.CreatedAt,
			UpdatedAt:       o.UpdatedAt,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "All orders retrieved",
		"data":    respOrders,
	})
}
