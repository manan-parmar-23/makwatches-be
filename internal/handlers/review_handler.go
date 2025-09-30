package handlers

import (
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

// ReviewHandler handles product review operations
type ReviewHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewReviewHandler creates a new instance of ReviewHandler
func NewReviewHandler(db *database.DBClient, cfg *config.Config) *ReviewHandler {
	return &ReviewHandler{
		DB:     db,
		Config: cfg,
	}
}

// GetProductReviews returns reviews for a specific product
func (h *ReviewHandler) GetProductReviews(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get product ID from parameters
	productID, err := primitive.ObjectIDFromHex(c.Params("productId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid product ID",
		})
	}

	// Parse query parameters for pagination
	page := 1
	limit := 10

	if c.Query("page") != "" {
		_, err := fmt.Sscanf(c.Query("page"), "%d", &page)
		if err != nil || page < 1 {
			page = 1
		}
	}

	if c.Query("limit") != "" {
		_, err := fmt.Sscanf(c.Query("limit"), "%d", &limit)
		if err != nil || limit < 1 || limit > 100 {
			limit = 10
		}
	}

	// Set up options for pagination and sorting
	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Newest first

	// Find reviews for the product
	reviewCollection := h.DB.Collections().Reviews
	cursor, err := reviewCollection.Find(
		ctx,
		bson.M{"product_id": productID},
		findOptions,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve reviews",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Decode the results
	var reviews []models.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode reviews",
			"error":   err.Error(),
		})
	}

	// Get user details for the reviews
	userIDs := make([]primitive.ObjectID, 0, len(reviews))
	for _, review := range reviews {
		userIDs = append(userIDs, review.UserID)
	}

	userCollection := h.DB.Collections().Users
	userCursor, err := userCollection.Find(
		ctx,
		bson.M{"_id": bson.M{"$in": userIDs}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve user details",
			"error":   err.Error(),
		})
	}
	defer userCursor.Close(ctx)

	// Map users by ID for quick lookup
	users := make(map[primitive.ObjectID]models.User)
	var userList []models.User
	if err := userCursor.All(ctx, &userList); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode users",
			"error":   err.Error(),
		})
	}

	for _, user := range userList {
		users[user.ID] = user
	}

	// Build response with user details
	response := make([]fiber.Map, 0, len(reviews))
	for _, review := range reviews {
		user, exists := users[review.UserID]

		// Set default user name if user not found
		userName := "Anonymous"
		if exists {
			userName = user.Name
		}

		response = append(response, fiber.Map{
			"id":        review.ID,
			"productId": review.ProductID,
			"userId":    review.UserID,
			"userName":  userName,
			"rating":    review.Rating,
			"title":     review.Title,
			"comment":   review.Comment,
			"photoUrls": review.PhotoURLs,
			"helpful":   review.Helpful,
			"verified":  review.Verified,
			"createdAt": review.CreatedAt,
		})
	}

	// Get total count for pagination info
	totalCount, err := reviewCollection.CountDocuments(ctx, bson.M{"product_id": productID})
	if err != nil {
		totalCount = int64(len(reviews))
	}

	// Calculate pagination info
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Reviews retrieved successfully",
		"data":    response,
		"pagination": fiber.Map{
			"page":       page,
			"limit":      limit,
			"totalItems": totalCount,
			"totalPages": totalPages,
		},
	})
}

// GetUserReviews returns reviews by the current user
func (h *ReviewHandler) GetUserReviews(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Parse query parameters for pagination
	page := 1
	limit := 10

	if c.Query("page") != "" {
		_, err := fmt.Sscanf(c.Query("page"), "%d", &page)
		if err != nil || page < 1 {
			page = 1
		}
	}

	if c.Query("limit") != "" {
		_, err := fmt.Sscanf(c.Query("limit"), "%d", &limit)
		if err != nil || limit < 1 || limit > 100 {
			limit = 10
		}
	}

	// Set up options for pagination and sorting
	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Newest first

	// Find reviews by the user
	reviewCollection := h.DB.Collections().Reviews
	cursor, err := reviewCollection.Find(
		ctx,
		bson.M{"user_id": user.UserID},
		findOptions,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve reviews",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Decode the results
	var reviews []models.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode reviews",
			"error":   err.Error(),
		})
	}

	// If no reviews found, return empty array
	if len(reviews) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "No reviews found",
			"data":    []models.Review{},
			"pagination": fiber.Map{
				"page":       page,
				"limit":      limit,
				"totalItems": 0,
				"totalPages": 0,
			},
		})
	}

	// Get product details for the reviews
	productIDs := make([]primitive.ObjectID, 0, len(reviews))
	for _, review := range reviews {
		productIDs = append(productIDs, review.ProductID)
	}

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
	response := make([]fiber.Map, 0, len(reviews))
	for _, review := range reviews {
		product, exists := products[review.ProductID]

		// Set default product name if product not found
		productName := "Unknown Product"
		productImage := ""
		if exists {
			productName = product.Name
			productImage = product.ImageURL
		}

		response = append(response, fiber.Map{
			"id":           review.ID,
			"productId":    review.ProductID,
			"productName":  productName,
			"productImage": productImage,
			"rating":       review.Rating,
			"title":        review.Title,
			"comment":      review.Comment,
			"photoUrls":    review.PhotoURLs,
			"helpful":      review.Helpful,
			"verified":     review.Verified,
			"createdAt":    review.CreatedAt,
		})
	}

	// Get total count for pagination info
	totalCount, err := reviewCollection.CountDocuments(ctx, bson.M{"user_id": user.UserID})
	if err != nil {
		totalCount = int64(len(reviews))
	}

	// Calculate pagination info
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Reviews retrieved successfully",
		"data":    response,
		"pagination": fiber.Map{
			"page":       page,
			"limit":      limit,
			"totalItems": totalCount,
			"totalPages": totalPages,
		},
	})
}

// CreateReview adds a new review for a product
func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
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
		ProductID string   `json:"productId" validate:"required"`
		Rating    float64  `json:"rating" validate:"required,min=1,max=5"`
		Title     string   `json:"title" validate:"required"`
		Comment   string   `json:"comment" validate:"required,min=5"`
		PhotoURLs []string `json:"photoUrls,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate the request
	if req.Rating < 1 || req.Rating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Rating must be between 1 and 5",
		})
	}

	if len(req.Comment) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Comment must be at least 5 characters long",
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

	// Check if user has already reviewed this product
	reviewCollection := h.DB.Collections().Reviews
	count, err := reviewCollection.CountDocuments(
		ctx,
		bson.M{
			"user_id":    user.UserID,
			"product_id": productID,
		},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check existing reviews",
			"error":   err.Error(),
		})
	}

	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "You have already reviewed this product",
		})
	}

	// Create the new review
	now := time.Now()
	review := models.Review{
		ID:        primitive.NewObjectID(),
		UserID:    user.UserID,
		ProductID: productID,
		Rating:    req.Rating,
		Title:     req.Title,
		Comment:   req.Comment,
		PhotoURLs: req.PhotoURLs,
		Helpful:   0,
		Verified:  true, // Set as verified if ordered by user (could check order history)
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert the review
	_, err = reviewCollection.InsertOne(ctx, review)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create review",
			"error":   err.Error(),
		})
	}

	// Update product rating
	// Get all ratings for the product
	cursor, err := reviewCollection.Find(
		ctx,
		bson.M{"product_id": productID},
	)
	if err == nil {
		defer cursor.Close(ctx)

		var reviews []models.Review
		if err := cursor.All(ctx, &reviews); err == nil {
			// Calculate average rating
			var totalRating float64
			for _, r := range reviews {
				totalRating += r.Rating
			}

			avgRating := totalRating / float64(len(reviews))

			// Update product rating
			_, err = productCollection.UpdateOne(
				ctx,
				bson.M{"_id": productID},
				bson.M{"$set": bson.M{
					"avg_rating":    avgRating,
					"ratings_count": len(reviews),
				}},
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Failed to update product rating",
					"error":   err.Error(),
				})
			}
		}
	}

	// Get user name
	userCollection := h.DB.Collections().Users
	var userData models.User
	err = userCollection.FindOne(ctx, bson.M{"_id": user.UserID}).Decode(&userData)
	userName := "Anonymous"
	if err == nil {
		userName = userData.Name
	}

	// Return the created review
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Review created successfully",
		"data": fiber.Map{
			"id":        review.ID,
			"productId": review.ProductID,
			"userId":    review.UserID,
			"userName":  userName,
			"rating":    review.Rating,
			"title":     review.Title,
			"comment":   review.Comment,
			"photoUrls": review.PhotoURLs,
			"helpful":   review.Helpful,
			"verified":  review.Verified,
			"createdAt": review.CreatedAt,
		},
	})
}

// UpdateReview modifies an existing review
func (h *ReviewHandler) UpdateReview(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get review ID from parameters
	reviewID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid review ID",
		})
	}

	// Parse request body
	var req struct {
		Rating    float64  `json:"rating" validate:"required,min=1,max=5"`
		Title     string   `json:"title" validate:"required"`
		Comment   string   `json:"comment" validate:"required,min=5"`
		PhotoURLs []string `json:"photoUrls,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate the request
	if req.Rating < 1 || req.Rating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Rating must be between 1 and 5",
		})
	}

	if len(req.Comment) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Comment must be at least 5 characters long",
		})
	}

	// Check if the review exists and belongs to the user
	reviewCollection := h.DB.Collections().Reviews
	var existingReview models.Review
	err = reviewCollection.FindOne(
		ctx,
		bson.M{
			"_id":     reviewID,
			"user_id": user.UserID,
		},
	).Decode(&existingReview)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Review not found or does not belong to you",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check review",
			"error":   err.Error(),
		})
	}

	// Update the review
	update := bson.M{
		"rating":     req.Rating,
		"title":      req.Title,
		"comment":    req.Comment,
		"updated_at": time.Now(),
	}

	if len(req.PhotoURLs) > 0 {
		update["photo_urls"] = req.PhotoURLs
	}

	_, err = reviewCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":     reviewID,
			"user_id": user.UserID,
		},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update review",
			"error":   err.Error(),
		})
	}

	// Update product rating
	// Get all ratings for the product
	productID := existingReview.ProductID
	cursor, err := reviewCollection.Find(
		ctx,
		bson.M{"product_id": productID},
	)
	if err == nil {
		defer cursor.Close(ctx)

		var reviews []models.Review
		if err := cursor.All(ctx, &reviews); err == nil {
			// Calculate average rating
			var totalRating float64
			for _, r := range reviews {
				totalRating += r.Rating
			}

			avgRating := totalRating / float64(len(reviews))

			// Update product rating
			productCollection := h.DB.Collections().Products
			_, err = productCollection.UpdateOne(
				ctx,
				bson.M{"_id": productID},
				bson.M{"$set": bson.M{
					"avg_rating":    avgRating,
					"ratings_count": len(reviews),
				}},
			)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Review updated successfully",
	})
}

// DeleteReview removes a review
func (h *ReviewHandler) DeleteReview(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get review ID from parameters
	reviewID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid review ID",
		})
	}

	// Check if the review exists and belongs to the user
	reviewCollection := h.DB.Collections().Reviews
	var existingReview models.Review
	err = reviewCollection.FindOne(
		ctx,
		bson.M{
			"_id":     reviewID,
			"user_id": user.UserID,
		},
	).Decode(&existingReview)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Review not found or does not belong to you",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check review",
			"error":   err.Error(),
		})
	}

	// Store product ID for rating update
	productID := existingReview.ProductID

	// Delete the review
	_, err = reviewCollection.DeleteOne(
		ctx,
		bson.M{
			"_id":     reviewID,
			"user_id": user.UserID,
		},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete review",
			"error":   err.Error(),
		})
	}

	// Update product rating
	cursor, err := reviewCollection.Find(
		ctx,
		bson.M{"product_id": productID},
	)
	if err == nil {
		defer cursor.Close(ctx)

		var reviews []models.Review
		if err := cursor.All(ctx, &reviews); err == nil {
			productCollection := h.DB.Collections().Products

			if len(reviews) > 0 {
				// Calculate average rating
				var totalRating float64
				for _, r := range reviews {
					totalRating += r.Rating
				}

				avgRating := totalRating / float64(len(reviews))

				// Update product rating
				_, err = productCollection.UpdateOne(
					ctx,
					bson.M{"_id": productID},
					bson.M{"$set": bson.M{
						"avg_rating":    avgRating,
						"ratings_count": len(reviews),
					}},
				)
			} else {
				// No reviews left, reset rating
				_, err = productCollection.UpdateOne(
					ctx,
					bson.M{"_id": productID},
					bson.M{"$set": bson.M{
						"avg_rating":    0,
						"ratings_count": 0,
					}},
				)
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Review deleted successfully",
	})
}

// MarkReviewHelpful increases the helpful count for a review
func (h *ReviewHandler) MarkReviewHelpful(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get review ID from parameters
	reviewID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid review ID",
		})
	}

	// Check if review exists
	reviewCollection := h.DB.Collections().Reviews
	var existingReview models.Review
	err = reviewCollection.FindOne(
		ctx,
		bson.M{"_id": reviewID},
	).Decode(&existingReview)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Review not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check review",
			"error":   err.Error(),
		})
	}

	// Increment helpful count
	_, err = reviewCollection.UpdateOne(
		ctx,
		bson.M{"_id": reviewID},
		bson.M{"$inc": bson.M{"helpful": 1}},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to mark review as helpful",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Review marked as helpful",
	})
}
