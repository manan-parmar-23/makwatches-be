package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Build response with product details (check all sources)
	productCollection := h.DB.Collections().Products
	heroCollection := h.DB.MongoDB.Collection("hero_slides")
	collectionCollection := h.DB.MongoDB.Collection("home_collection_features")

	response := make([]fiber.Map, 0, len(wishlistItems))
	for _, item := range wishlistItems {
		// Try to find product in regular products
		var product models.Product
		err := productCollection.FindOne(ctx, bson.M{"_id": item.ProductID}).Decode(&product)
		if err == nil {
			// Found in regular products
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
			continue
		}

		// Try to find in hero slides by productId
		var heroSlide models.HeroSlide
		err = heroCollection.FindOne(ctx, bson.M{"productId": item.ProductID}).Decode(&heroSlide)
		if err == nil {
			// Found in hero slides
			priceFloat := 0.0
			fmt.Sscanf(heroSlide.Price, "₹%f", &priceFloat)
			if priceFloat == 0 {
				fmt.Sscanf(heroSlide.Price, "%f", &priceFloat)
			}

			response = append(response, fiber.Map{
				"wishlistId":  item.ID,
				"productId":   item.ProductID,
				"name":        heroSlide.Title,
				"price":       priceFloat,
				"image":       heroSlide.Image,
				"description": heroSlide.Description,
				"inStock":     true, // Home content always in stock
				"addedAt":     item.CreatedAt,
				"source":      "hero_slide",
			})
			continue
		}

		// Try to find in collection features by productId
		var collectionFeature models.HomeCollectionFeature
		err = collectionCollection.FindOne(ctx, bson.M{"productId": item.ProductID}).Decode(&collectionFeature)
		if err == nil {
			// Found in collection features
			priceFloat := 0.0
			if collectionFeature.Price != "" {
				fmt.Sscanf(collectionFeature.Price, "₹%f", &priceFloat)
				if priceFloat == 0 {
					fmt.Sscanf(collectionFeature.Price, "%f", &priceFloat)
				}
			}

			response = append(response, fiber.Map{
				"wishlistId":  item.ID,
				"productId":   item.ProductID,
				"name":        collectionFeature.Title,
				"price":       priceFloat,
				"image":       collectionFeature.Image,
				"description": collectionFeature.Description,
				"inStock":     true, // Home content always in stock
				"addedAt":     item.CreatedAt,
				"source":      "collection_feature",
			})
			continue
		}

		// Product not found in any collection - skip
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

	// Check if product exists (try home content first, then regular products)
	heroCollection := h.DB.MongoDB.Collection("hero_slides")
	collectionCollection := h.DB.MongoDB.Collection("home_collection_features")
	productCollection := h.DB.Collections().Products

	productExists := false

	// Try hero slides
	var heroSlide models.HeroSlide
	err = heroCollection.FindOne(ctx, bson.M{"productId": productID}).Decode(&heroSlide)
	if err == nil {
		productExists = true
	} else {
		// Try collection features
		var collectionFeature models.HomeCollectionFeature
		err = collectionCollection.FindOne(ctx, bson.M{"productId": productID}).Decode(&collectionFeature)
		if err == nil {
			productExists = true
		} else {
			// Try regular products
			var product models.Product
			err = productCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
			if err == nil {
				productExists = true
			}
		}
	}

	if !productExists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Product not found",
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

	// Return success message - GetWishlist will fetch full details
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product added to wishlist",
		"data": fiber.Map{
			"wishlistId": wishlistItem.ID,
			"productId":  productID,
			"addedAt":    wishlistItem.CreatedAt,
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
