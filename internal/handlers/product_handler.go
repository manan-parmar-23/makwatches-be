package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/the-devesta/pehnaw-be/internal/config"
	"github.com/the-devesta/pehnaw-be/internal/database"
	"github.com/the-devesta/pehnaw-be/internal/models"
)

// ProductHandler handles product related requests
type ProductHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewProductHandler creates a new instance of ProductHandler
func NewProductHandler(db *database.DBClient, cfg *config.Config) *ProductHandler {
	return &ProductHandler{
		DB:     db,
		Config: cfg,
	}
}

// GetProducts returns all products with optional filters
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	ctx := c.Context()
	
	// Parse query parameters for filtering
	category := c.Query("category")
	minPriceStr := c.Query("minPrice")
	maxPriceStr := c.Query("maxPrice")
	sortBy := c.Query("sortBy", "createdAt") // Default sort by createdAt
	order := c.Query("order", "desc")        // Default order desc
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	
	// Convert string parameters to appropriate types
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	
	// Build the filter
	filter := bson.M{}
	
	// Add category filter if provided
	if category != "" {
		filter["category"] = category
	}
	
	// Add price range filters if provided
	if minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err == nil && minPrice >= 0 {
			filter["price"] = bson.M{"$gte": minPrice}
		}
	}
	
	if maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err == nil && maxPrice > 0 {
			if _, ok := filter["price"]; ok {
				filter["price"].(bson.M)["$lte"] = maxPrice
			} else {
				filter["price"] = bson.M{"$lte": maxPrice}
			}
		}
	}
	
	// Determine sort direction
	sortDirection := 1 // ascending
	if order == "desc" {
		sortDirection = -1 // descending
	}
	
	// Configure options for pagination and sorting
	findOptions := options.Find()
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: sortBy, Value: sortDirection}})
	
	// First check if we have this query cached in Redis
	cacheKey := fmt.Sprintf("products:%s:%s:%s:%s:%d:%d", 
		category, minPriceStr, maxPriceStr, sortBy, page, limit)
	
	var products []models.Product
	err = h.DB.CacheGet(ctx, cacheKey, &products)
	if err == nil {
		// Cache hit
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Products retrieved from cache",
			"data": products,
			"meta": fiber.Map{
				"page":  page,
				"limit": limit,
			},
		})
	}
	
	// Cache miss, get from database
	collection := h.DB.Collections().Products
	
	// Count total matching documents for pagination info
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count products",
			"error":   err.Error(),
		})
	}
	
	// Execute the query
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve products",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)
	
	// Decode the results
	if err := cursor.All(ctx, &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode products",
			"error":   err.Error(),
		})
	}
	
	// Cache the results for future requests (expire after 10 minutes)
	h.DB.CacheSet(ctx, cacheKey, products, 10*time.Minute)
	
	// Return the products
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Products retrieved successfully",
		"data":    products,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit), // ceiling division
		},
	})
}

// GetProductByID returns a single product by ID
func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	ctx := c.Context()
	
	// Get product ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product ID is required",
		})
	}
	
	// Check if the product is in Redis cache
	cacheKey := fmt.Sprintf("product:%s", id)
	
	var product models.Product
	err := h.DB.CacheGet(ctx, cacheKey, &product)
	if err == nil {
		// Cache hit
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Product retrieved from cache",
			"data":    product,
		})
	}
	
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid product ID format",
			"error":   err.Error(),
		})
	}
	
	// Find product in database
	collection := h.DB.Collections().Products
	if err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&product); err != nil {
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
	
	// Cache the product for future requests (expire after 30 minutes)
	h.DB.CacheSet(ctx, cacheKey, product, 30*time.Minute)
	
	// Return the product
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product retrieved successfully",
		"data":    product,
	})
}
