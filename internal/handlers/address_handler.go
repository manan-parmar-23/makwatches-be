package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/the-devesta/pehnaw-be/internal/config"
	"github.com/the-devesta/pehnaw-be/internal/database"
	"github.com/the-devesta/pehnaw-be/internal/middleware"
	"github.com/the-devesta/pehnaw-be/internal/models"
)

// AddressBookHandler handles address operations
type AddressBookHandler struct {
	DB     *database.DBClient
	Config *config.Config
}

// NewAddressBookHandler creates a new instance of AddressBookHandler
func NewAddressBookHandler(db *database.DBClient, cfg *config.Config) *AddressBookHandler {
	return &AddressBookHandler{
		DB:     db,
		Config: cfg,
	}
}

// GetAddresses returns all addresses for the current user
func (h *AddressBookHandler) GetAddresses(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Find all addresses
	addressCollection := h.DB.Collections().UserAddresses
	cursor, err := addressCollection.Find(ctx, bson.M{"user_id": user.UserID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve addresses",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Decode the results
	addresses := []models.UserAddress{}
	if err := cursor.All(ctx, &addresses); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to decode addresses",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Addresses retrieved successfully",
		"data":    addresses,
	})
}

// GetAddress returns a single address by ID
func (h *AddressBookHandler) GetAddress(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get address ID from parameters
	addressID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid address ID",
		})
	}

	// Find the address
	var address models.UserAddress
	addressCollection := h.DB.Collections().UserAddresses
	err = addressCollection.FindOne(ctx, bson.M{
		"_id":     addressID,
		"user_id": user.UserID,
	}).Decode(&address)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Address not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve address",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Address retrieved successfully",
		"data":    address,
	})
}

// CreateAddress adds a new address to the user's address book
func (h *AddressBookHandler) CreateAddress(c *fiber.Ctx) error {
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
	var req models.UserAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Create the new address
	now := time.Now()
	newAddress := models.UserAddress{
		ID:        primitive.NewObjectID(),
		UserID:    user.UserID,
		Name:      req.Name,
		Street:    req.Street,
		City:      req.City,
		State:     req.State,
		ZipCode:   req.ZipCode,
		Country:   req.Country,
		Phone:     req.Phone,
		IsDefault: req.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Check if this is the default address
	addressCollection := h.DB.Collections().UserAddresses
	if req.IsDefault {
		// Update existing default addresses
		_, err := addressCollection.UpdateMany(
			ctx,
			bson.M{"user_id": user.UserID, "is_default": true},
			bson.M{"$set": bson.M{"is_default": false, "updated_at": now}},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update existing default address",
				"error":   err.Error(),
			})
		}
	} else {
		// Check if this is the first address, if so make it default
		count, err := addressCollection.CountDocuments(ctx, bson.M{"user_id": user.UserID})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count addresses",
				"error":   err.Error(),
			})
		}

		if count == 0 {
			newAddress.IsDefault = true
		}
	}

	// Insert the address
	_, err := addressCollection.InsertOne(ctx, newAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create address",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Address created successfully",
		"data":    newAddress,
	})
}

// UpdateAddress updates an existing address
func (h *AddressBookHandler) UpdateAddress(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get address ID from parameters
	addressID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid address ID",
		})
	}

	// Parse request body
	var req models.UserAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Prepare the update
	now := time.Now()
	update := bson.M{
		"name":       req.Name,
		"street":     req.Street,
		"city":       req.City,
		"state":      req.State,
		"zip_code":   req.ZipCode,
		"country":    req.Country,
		"phone":      req.Phone,
		"updated_at": now,
	}

	addressCollection := h.DB.Collections().UserAddresses

	// Check if this is the default address
	if req.IsDefault {
		// Update existing default addresses
		_, err := addressCollection.UpdateMany(
			ctx,
			bson.M{
				"user_id":    user.UserID,
				"is_default": true,
				"_id":        bson.M{"$ne": addressID},
			},
			bson.M{"$set": bson.M{"is_default": false, "updated_at": now}},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update existing default address",
				"error":   err.Error(),
			})
		}
		update["is_default"] = true
	}

	// Update the address
	result, err := addressCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":     addressID,
			"user_id": user.UserID,
		},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update address",
			"error":   err.Error(),
		})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Address not found or does not belong to you",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Address updated successfully",
	})
}

// DeleteAddress removes an address from the user's address book
func (h *AddressBookHandler) DeleteAddress(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get address ID from parameters
	addressID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid address ID",
		})
	}

	// Find the address to check if it's default
	var address models.UserAddress
	addressCollection := h.DB.Collections().UserAddresses
	err = addressCollection.FindOne(ctx, bson.M{
		"_id":     addressID,
		"user_id": user.UserID,
	}).Decode(&address)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Address not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve address",
			"error":   err.Error(),
		})
	}

	// Delete the address
	result, err := addressCollection.DeleteOne(
		ctx,
		bson.M{
			"_id":     addressID,
			"user_id": user.UserID,
		},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete address",
			"error":   err.Error(),
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Address not found or does not belong to you",
		})
	}

	// If deleted address was default, set another address as default
	if address.IsDefault {
		limit := int64(1)
		cursor, err := addressCollection.Find(
			ctx,
			bson.M{"user_id": user.UserID},
			options.Find().
				SetLimit(limit).
				SetSort(bson.D{{Key: "created_at", Value: 1}}),
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to find replacement default address",
				"error":   err.Error(),
			})
		}
		defer cursor.Close(ctx)

		var addresses []models.UserAddress
		if err := cursor.All(ctx, &addresses); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to decode addresses",
				"error":   err.Error(),
			})
		}

		if len(addresses) > 0 {
			// Set the first address as default
			_, err = addressCollection.UpdateOne(
				ctx,
				bson.M{"_id": addresses[0].ID},
				bson.M{"$set": bson.M{"is_default": true, "updated_at": time.Now()}},
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Failed to update new default address",
					"error":   err.Error(),
				})
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Address deleted successfully",
	})
}

// SetDefaultAddress sets an address as the default
func (h *AddressBookHandler) SetDefaultAddress(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get user info from token
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}

	// Get address ID from parameters
	addressID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid address ID",
		})
	}

	now := time.Now()
	addressCollection := h.DB.Collections().UserAddresses

	// First, verify the address exists and belongs to the user
	count, err := addressCollection.CountDocuments(
		ctx,
		bson.M{
			"_id":     addressID,
			"user_id": user.UserID,
		},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to verify address",
			"error":   err.Error(),
		})
	}

	if count == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Address not found or does not belong to you",
		})
	}

	// Update existing default addresses
	_, err = addressCollection.UpdateMany(
		ctx,
		bson.M{
			"user_id":    user.UserID,
			"is_default": true,
		},
		bson.M{"$set": bson.M{"is_default": false, "updated_at": now}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update existing default address",
			"error":   err.Error(),
		})
	}

	// Set the new default address
	_, err = addressCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":     addressID,
			"user_id": user.UserID,
		},
		bson.M{"$set": bson.M{"is_default": true, "updated_at": now}},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to set default address",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Address set as default successfully",
	})
}
