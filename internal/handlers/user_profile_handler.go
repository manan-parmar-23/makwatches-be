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

// UserProfileHandler handles user profile operations
type UserProfileHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewUserProfileHandler creates a new instance of UserProfileHandler
func NewUserProfileHandler(db *database.DBClient, cfg *config.Config) *UserProfileHandler {
	return &UserProfileHandler{
		DB:     db,
		Config: cfg,
	}
}

// GetProfile returns the user's profile information
func (h *UserProfileHandler) GetProfile(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get base user data
	var userData models.User
	userCollection := h.DB.Collections().Users
	err := userCollection.FindOne(ctx, bson.M{"_id": user.UserID}).Decode(&userData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve user data",
			"error":   err.Error(),
		})
	}

	// Get extended profile data
	var profile models.UserProfile
	profileCollection := h.DB.Collections().UserProfiles
	err = profileCollection.FindOne(ctx, bson.M{"user_id": user.UserID}).Decode(&profile)
	if err != nil && err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve profile data",
			"error":   err.Error(),
		})
	}

	// Combine user and profile data
	response := fiber.Map{
		"id":        userData.ID,
		"name":      userData.Name,
		"email":     userData.Email,
		"role":      userData.Role,
		"createdAt": userData.CreatedAt,
	}

	// Add profile data if exists
	if err != mongo.ErrNoDocuments {
		response["dateOfBirth"] = profile.DateOfBirth
		response["gender"] = profile.Gender
		response["phone"] = profile.Phone
		response["avatarUrl"] = profile.AvatarURL
		response["bio"] = profile.Bio
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User profile retrieved successfully",
		"data":    response,
	})
}

// UpdateProfile updates the user's profile information
func (h *UserProfileHandler) UpdateProfile(c *fiber.Ctx) error {
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
	var req models.ProfileUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Check if profile exists
	profileCollection := h.DB.Collections().UserProfiles
	var existingProfile models.UserProfile
	err := profileCollection.FindOne(ctx, bson.M{"user_id": user.UserID}).Decode(&existingProfile)

	now := time.Now()

	if err == mongo.ErrNoDocuments {
		// Create new profile
		newProfile := models.UserProfile{
			UserID:      user.UserID,
			DateOfBirth: req.DateOfBirth,
			Gender:      req.Gender,
			Phone:       req.Phone,
			AvatarURL:   req.AvatarURL,
			Bio:         req.Bio,
			UpdatedAt:   now,
		}

		_, err = profileCollection.InsertOne(ctx, newProfile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create profile",
				"error":   err.Error(),
			})
		}
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check existing profile",
			"error":   err.Error(),
		})
	} else {
		// Update existing profile
		update := bson.M{"updated_at": now}

		if req.DateOfBirth != nil {
			update["date_of_birth"] = req.DateOfBirth
		}
		if req.Gender != "" {
			update["gender"] = req.Gender
		}
		if req.Phone != "" {
			update["phone"] = req.Phone
		}
		if req.AvatarURL != "" {
			update["avatar_url"] = req.AvatarURL
		}
		if req.Bio != "" {
			update["bio"] = req.Bio
		}

		_, err = profileCollection.UpdateOne(
			ctx,
			bson.M{"user_id": user.UserID},
			bson.M{"$set": update},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update profile",
				"error":   err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
	})
}

// UpdatePreferences updates the user's preferences
func (h *UserProfileHandler) UpdatePreferences(c *fiber.Ctx) error {
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
	var req models.PreferencesUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Check if preferences exist
	prefsCollection := h.DB.Collections().UserPreferences
	var existingPrefs models.UserPreferences
	err := prefsCollection.FindOne(ctx, bson.M{"user_id": user.UserID}).Decode(&existingPrefs)

	now := time.Now()

	if err == mongo.ErrNoDocuments {
		// Create new preferences
		newPrefs := models.UserPreferences{
			ID:                 primitive.NewObjectID(),
			UserID:             user.UserID,
			FavoriteCategories: req.FavoriteCategories,
			FavoriteBrands:     req.FavoriteBrands,
			SizePreferences:    req.SizePreferences,
			ColorPreferences:   req.ColorPreferences,
			PriceRange:         req.PriceRange,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		_, err = prefsCollection.InsertOne(ctx, newPrefs)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create preferences",
				"error":   err.Error(),
			})
		}
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to check existing preferences",
			"error":   err.Error(),
		})
	} else {
		// Update existing preferences
		update := bson.M{"updated_at": now}

		if req.FavoriteCategories != nil {
			update["favorite_categories"] = req.FavoriteCategories
		}
		if req.FavoriteBrands != nil {
			update["favorite_brands"] = req.FavoriteBrands
		}
		if req.SizePreferences != nil {
			update["size_preferences"] = req.SizePreferences
		}
		if req.ColorPreferences != nil {
			update["color_preferences"] = req.ColorPreferences
		}
		if req.PriceRange != nil {
			update["price_range"] = req.PriceRange
		}

		_, err = prefsCollection.UpdateOne(
			ctx,
			bson.M{"user_id": user.UserID},
			bson.M{"$set": update},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update preferences",
				"error":   err.Error(),
			})
		}
	}

	// Invalidate recommendations cache
	cacheKey := fmt.Sprintf("recommendations:%s", user.UserID.Hex())
	h.DB.CacheDel(ctx, cacheKey)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Preferences updated successfully",
	})
}
