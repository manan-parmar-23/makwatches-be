package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/the-devesta/pehnaw-be/internal/models"
	"github.com/the-devesta/pehnaw-be/pkg/utils"
)

// CreateProduct adds a new product to the database (admin only)
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse product data
	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid product data",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if product.Name == "" || product.Description == "" || product.Price <= 0 || product.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Missing required product fields",
		})
	}

	// Get S3 client
	s3Client, err := utils.NewS3Client(
		h.Config.AWSS3AccessKey,
		h.Config.AWSS3SecretKey,
		h.Config.AWSS3Region,
		h.Config.AWSS3BucketName,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to initialize S3 client",
			"error":   err.Error(),
		})
	}

	// Upload images (if any)
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]
		if len(files) > 0 {
			product.Images = []string{} // Initialize the Images array
			
			// Upload each image
			for _, file := range files {
				// Upload to S3
				imageURL, err := s3Client.UploadFile(ctx, file, "products")
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"message": "Failed to upload image",
						"error":   err.Error(),
					})
				}
				
				// Add URL to the product
				product.Images = append(product.Images, imageURL)
				
				// If no main image is set yet, use this one
				if product.ImageURL == "" {
					product.ImageURL = imageURL
				}
			}
		}
	}
	
	// Set timestamps
	product.CreatedAt = time.Now()
	product.UpdatedAt = product.CreatedAt

	// Insert product into database
	collection := h.DB.Collections().Products
	result, err := collection.InsertOne(ctx, product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create product",
			"error":   err.Error(),
		})
	}

	// Get the inserted ID
	product.ID = result.InsertedID.(primitive.ObjectID)

	// Invalidate relevant caches
	cacheKey := "products:" + product.Category
	h.DB.CacheDel(ctx, cacheKey)
	
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product created successfully",
		"data":    product,
	})
}

// UpdateProduct updates an existing product (admin only)
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get product ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product ID is required",
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

	// First, get the existing product to check if it exists
	collection := h.DB.Collections().Products
	var existingProduct models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingProduct)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
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

	// Fiber handles multipart form parsing automatically

	// Parse product data
	var updatedProduct models.Product
	if err := c.BodyParser(&updatedProduct); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid product data",
			"error":   err.Error(),
		})
	}

	// Keep existing fields if not provided
	if updatedProduct.Name == "" {
		updatedProduct.Name = existingProduct.Name
	}
	if updatedProduct.Description == "" {
		updatedProduct.Description = existingProduct.Description
	}
	if updatedProduct.Price <= 0 {
		updatedProduct.Price = existingProduct.Price
	}
	if updatedProduct.Category == "" {
		updatedProduct.Category = existingProduct.Category
	}
	if updatedProduct.Stock < 0 {
		updatedProduct.Stock = existingProduct.Stock
	}

	// Handle keepExistingImages flag
	keepExistingImages := true
	if keepImagesStr := c.FormValue("keepExistingImages"); keepImagesStr != "" {
		keepExistingImages, _ = strconv.ParseBool(keepImagesStr)
	}

	// Initialize with existing images if keeping them
	if keepExistingImages {
		updatedProduct.Images = existingProduct.Images
		updatedProduct.ImageURL = existingProduct.ImageURL
	} else {
		updatedProduct.Images = []string{}
		updatedProduct.ImageURL = ""
	}

	// Get S3 client for image uploads
	s3Client, err := utils.NewS3Client(
		h.Config.AWSS3AccessKey,
		h.Config.AWSS3SecretKey,
		h.Config.AWSS3Region,
		h.Config.AWSS3BucketName,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to initialize S3 client",
			"error":   err.Error(),
		})
	}

	// Upload new images (if any)
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]
		if len(files) > 0 {
			// Upload each image
			for _, file := range files {
				// Upload to S3
				imageURL, err := s3Client.UploadFile(ctx, file, "products")
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"message": "Failed to upload image",
						"error":   err.Error(),
					})
				}
				
				// Add URL to the product
				updatedProduct.Images = append(updatedProduct.Images, imageURL)
				
				// If no main image is set yet, use this one
				if updatedProduct.ImageURL == "" {
					updatedProduct.ImageURL = imageURL
				}
			}
		}
	}

	// Ensure at least one image
	if len(updatedProduct.Images) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product must have at least one image",
		})
	}

	// Keep original ID and created timestamp
	updatedProduct.ID = objectID
	updatedProduct.CreatedAt = existingProduct.CreatedAt
	updatedProduct.UpdatedAt = time.Now()

	// Update the product in database
	update := bson.M{
		"$set": bson.M{
			"name":        updatedProduct.Name,
			"description": updatedProduct.Description,
			"price":       updatedProduct.Price,
			"category":    updatedProduct.Category,
			"image_url":   updatedProduct.ImageURL,
			"images":      updatedProduct.Images,
			"stock":       updatedProduct.Stock,
			"updated_at":  updatedProduct.UpdatedAt,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update product",
			"error":   err.Error(),
		})
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("product:%s", id)
	h.DB.CacheDel(ctx, cacheKey)
	categoryCacheKey := "products:" + updatedProduct.Category
	h.DB.CacheDel(ctx, categoryCacheKey)
	oldCategoryCacheKey := "products:" + existingProduct.Category
	h.DB.CacheDel(ctx, oldCategoryCacheKey)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product updated successfully",
		"data":    updatedProduct,
	})
}

// DeleteProduct removes a product from the database (admin only)
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get product ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Product ID is required",
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

	// First, get the product to delete (to get image URLs)
	collection := h.DB.Collections().Products
	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
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

	// Delete the product
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete product",
			"error":   err.Error(),
		})
	}

	// Optional: Delete images from S3
	// Note: You might want to skip this if images could be shared between products
	if len(product.Images) > 0 && c.Query("deleteImages") == "true" {
		s3Client, err := utils.NewS3Client(
			h.Config.AWSS3AccessKey,
			h.Config.AWSS3SecretKey,
			h.Config.AWSS3Region,
			h.Config.AWSS3BucketName,
		)
		if err != nil {
			// Log error but don't fail the request
			// Images will remain in S3 but product is deleted from DB
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"success": true,
				"message": "Product deleted successfully but images remain in storage",
			})
		}

		// Delete all images
		for _, imageURL := range product.Images {
			_ = s3Client.DeleteFile(ctx, imageURL)
			// Continue even if some images fail to delete
		}
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("product:%s", id)
	h.DB.CacheDel(ctx, cacheKey)
	categoryCacheKey := "products:" + product.Category
	h.DB.CacheDel(ctx, categoryCacheKey)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product deleted successfully",
	})
}
