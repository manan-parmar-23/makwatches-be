package handlers

import (
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

// WishlistHandler handles wishlist operations
type WishlistHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewWishlistHandler creates a new instance of WishlistHandler
func NewWishlistHandler(db *database.DBClient, cfg *config.Config) *WishlistHandler {
	return &WishlistHandler{
		DB:     db,
		Config: cfg,
	}
}

// GetWishlist returns all items in the user's wishlist
func (h *WishlistHandler) GetWishlist(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get wishlist items
	wishlistCollection := h.DB.Collections().Wishlists
	cursor, err := wishlistCollection.Find(
		ctx,
		bson.M{"user_id": user.UserID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve wishlist",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Decode wishlist items
	var wishlistItems []models.Wishlist
	if err := cursor.All(ctx, &wishlistItems); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode wishlist items",
			"error":   err.Error(),
		})
	}

	// If no items found, return empty array
	if len(wishlistItems) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "No items in wishlist",
			"data":    []models.Wishlist{},
		})
	}

	// Collect product IDs to retrieve product details
	productIDs := make([]primitive.ObjectID, 0, len(wishlistItems))
	for _, item := range wishlistItems {
		productIDs = append(productIDs, item.ProductID)
	}

	// Get product details
	productCollection := h.DB.Collections().Products
	productCursor, err := productCollection.Find(
		ctx,
		bson.M{"_id": bson.M{"$in": productIDs}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve product details",
			"error":   err.Error(),
		})
	}
	defer productCursor.Close(ctx)

	// Map products by ID for quick lookup
	products := make(map[primitive.ObjectID]models.Product)
	var productList []models.Product
	if err := productCursor.All(ctx, &productList); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode products",
			"error":   err.Error(),
		})
	}

	for _, product := range productList {
		products[product.ID] = product
	}

	// Build response with product details
	response := make([]fiber.Map, 0, len(wishlistItems))
	for _, item := range wishlistItems {
		product, exists := products[item.ProductID]
		if !exists {
			continue
		}

		response = append(response, fiber.Map{
			"wishlistId":  item.ID,
			"productId":   product.ID,
			"name":        product.Name,
			"price":       product.Price,
			"image":       product.ImageURL,
			"description": product.Description,
			"inStock":     product.Stock > 0,
			"addedAt":     item.CreatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Wishlist retrieved successfully",
		"data":    response,
	})
}

// AddToWishlist adds a product to the user's wishlist
func (h *WishlistHandler) AddToWishlist(c *fiber.Ctx) error {
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
	var req struct {
		ProductID string `json:"productId" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Convert string ID to ObjectID
	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid product ID",
		})
	}

	// Check if product exists
	productCollection := h.DB.Collections().Products
	var product models.Product
	err = productCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Product not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check product",
			"error":   err.Error(),
		})
	}

	// Check if product is already in wishlist
	wishlistCollection := h.DB.Collections().Wishlists
	count, err := wishlistCollection.CountDocuments(
		ctx,
		bson.M{
			"user_id":    user.UserID,
			"product_id": productID,
		},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check wishlist",
			"error":   err.Error(),
		})
	}

	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Product already in wishlist",
		})
	}

	// Add product to wishlist
	now := time.Now()
	wishlistItem := models.Wishlist{
		ID:        primitive.NewObjectID(),
		UserID:    user.UserID,
		ProductID: productID,
		CreatedAt: now,
	}

	_, err = wishlistCollection.InsertOne(ctx, wishlistItem)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to add product to wishlist",
			"error":   err.Error(),
		})
	}

	// Return product details with wishlist info
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product added to wishlist",
		"data": fiber.Map{
			"wishlistId":  wishlistItem.ID,
			"productId":   product.ID,
			"name":        product.Name,
			"price":       product.Price,
			"image":       product.ImageURL,
			"description": product.Description,
			"inStock":     product.Stock > 0,
			"addedAt":     wishlistItem.CreatedAt,
		},
	})
}

// RemoveFromWishlist removes a product from the user's wishlist
func (h *WishlistHandler) RemoveFromWishlist(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get wishlist item ID from parameters
	itemID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid wishlist item ID",
		})
	}

	// Delete wishlist item
	wishlistCollection := h.DB.Collections().Wishlists
	result, err := wishlistCollection.DeleteOne(
		ctx,
		bson.M{
			"_id":     itemID,
			"user_id": user.UserID,
		},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove product from wishlist",
			"error":   err.Error(),
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Wishlist item not found or does not belong to you",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product removed from wishlist",
	})
}

// ClearWishlist removes all products from the user's wishlist
func (h *WishlistHandler) ClearWishlist(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Delete all wishlist items
	wishlistCollection := h.DB.Collections().Wishlists
	result, err := wishlistCollection.DeleteMany(
		ctx,
		bson.M{"user_id": user.UserID},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to clear wishlist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Wishlist cleared successfully",
		"count":   result.DeletedCount,
	})
}
