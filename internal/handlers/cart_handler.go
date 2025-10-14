package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/shivam-mishra-20/mak-watches-be/internal/config"
	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"github.com/shivam-mishra-20/mak-watches-be/internal/middleware"
	"github.com/shivam-mishra-20/mak-watches-be/internal/models"
)

// CartHandler handles cart related requests
type CartHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewCartHandler creates a new instance of CartHandler
func NewCartHandler(db *database.DBClient, cfg *config.Config) *CartHandler {
	return &CartHandler{
		DB:     db,
		Config: cfg,
	}
}

// AddToCart adds a product to the user's cart
func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from the token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Parse request body
	var req models.CartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.ProductID == "" || req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product ID and quantity > 0 are required",
		})
	}

	// Convert product ID from string to ObjectID
	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid product ID format",
			"error":   err.Error(),
		})
	}

	// Check if the product exists
	var product models.Product
	collection := h.DB.Collections().Products
	err = collection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Product not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve product",
			"error":   err.Error(),
		})
	}

	// Check if the product is in stock
	if product.Stock < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Not enough stock available",
		})
	}

	// Check if the product (same size) is already in the cart. Size empty matches only empty.
	cartCollection := h.DB.Collections().CartItems
	var existingCartItem models.CartItem
	query := bson.M{"user_id": user.UserID, "product_id": productID}
	if req.Size != "" {
		query["size"] = req.Size
	} else {
		query["size"] = bson.M{"$in": bson.A{"", nil}}
	}
	err = cartCollection.FindOne(ctx, query).Decode(&existingCartItem)

	now := time.Now()

	switch err {
	case nil:
		// Update existing cart item
		_, err = cartCollection.UpdateOne(
			ctx,
			bson.M{"_id": existingCartItem.ID},
			bson.M{
				"$set": bson.M{
					"quantity":   existingCartItem.Quantity + req.Quantity,
					"updated_at": now,
				},
			},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update cart item",
				"error":   err.Error(),
			})
		}
	case mongo.ErrNoDocuments:
		// Add new cart item
		cartItem := models.CartItem{
			ID:        primitive.NewObjectID(),
			UserID:    user.UserID,
			ProductID: productID,
			Size:      req.Size,
			Quantity:  req.Quantity,
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err = cartCollection.InsertOne(ctx, cartItem)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to add product to cart",
				"error":   err.Error(),
			})
		}
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	// Invalidate cart cache
	cacheKey := fmt.Sprintf("cart:%s", user.UserID.Hex())
	h.DB.CacheDel(ctx, cacheKey)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product added to cart successfully",
	})
}

// GetCart retrieves a user's cart
func (h *CartHandler) GetCart(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user ID from URL parameter or from token
	userIDParam := c.Params("userID")
	var userID primitive.ObjectID
	var err error

	if userIDParam != "" {
		userID, err = primitive.ObjectIDFromHex(userIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format",
				"error":   err.Error(),
			})
		}
	} else {
		// Get user info from token
		user, ok := c.Locals("user").(*middleware.TokenMetadata)
		if !ok || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Unauthorized - User data not found",
			})
		}
		userID = user.UserID
	}

	// Check if the cart is in Redis cache
	cacheKey := fmt.Sprintf("cart:%s", userID.Hex())
	var cartResponse models.CartResponse
	err = h.DB.CacheGet(ctx, cacheKey, &cartResponse)
	if err == nil {
		// Cache hit
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Cart retrieved from cache",
			"data":    cartResponse,
		})
	}

	// Find all cart items for the user
	cartCollection := h.DB.Collections().CartItems
	cursor, err := cartCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve cart items",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Parse the results
	var cartItems []models.CartItem
	if err := cursor.All(ctx, &cartItems); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode cart items",
			"error":   err.Error(),
		})
	}

	// If cart is empty
	if len(cartItems) == 0 {
		emptyCart := models.CartResponse{
			Items: []models.CartItem{},
			Total: 0,
		}

		// Cache empty cart (expire after 30 minutes)
		h.DB.CacheSet(ctx, cacheKey, emptyCart, 30*time.Minute)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Cart is empty",
			"data":    emptyCart,
		})
	}

	// Fetch product details for each cart item
	productCollection := h.DB.Collections().Products
	var total float64

	for i, item := range cartItems {
		var product models.Product
		err := productCollection.FindOne(ctx, bson.M{"_id": item.ProductID}).Decode(&product)
		if err == nil {
			cartItems[i].Product = &product
			total += product.Price * float64(item.Quantity)
		}
	}

	// Create cart response
	cartResponse = models.CartResponse{
		Items: cartItems,
		Total: total,
	}

	// Cache the cart (expire after 30 minutes)
	h.DB.CacheSet(ctx, cacheKey, cartResponse, 30*time.Minute)

	// Return the cart
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Cart retrieved successfully",
		"data":    cartResponse,
	})
}

// RemoveFromCart removes an item from the cart
func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user ID and product ID from URL parameters
	userIDParam := c.Params("userID")
	productIDParam := c.Params("productID")

	if userIDParam == "" || productIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID and product ID are required",
		})
	}

	// Convert IDs from string to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
			"error":   err.Error(),
		})
	}

	productID, err := primitive.ObjectIDFromHex(productIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid product ID format",
			"error":   err.Error(),
		})
	}

	// Check if the user is authorized to remove this item
	tokenUser, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok || tokenUser == nil || (tokenUser.UserID != userID && tokenUser.Role != "admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Not authorized to modify this cart",
		})
	}

	// Remove the item from the cart
	cartCollection := h.DB.Collections().CartItems
	result, err := cartCollection.DeleteOne(ctx, bson.M{
		"user_id":    userID,
		"product_id": productID,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove item from cart",
			"error":   err.Error(),
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Item not found in cart",
		})
	}

	// Invalidate cart cache
	cacheKey := fmt.Sprintf("cart:%s", userID.Hex())
	h.DB.CacheDel(ctx, cacheKey)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Item removed from cart successfully",
	})
}
