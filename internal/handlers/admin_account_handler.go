// internal/handlers/admin_account_handler.go
package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Account is a light representation for admin listing
type Account struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name  string             `json:"name" bson:"name"`
	Email string             `json:"email" bson:"email"`
	Role  string             `json:"role" bson:"role"`
}

type AdminAccountHandler struct {
	DB *database.DBClient
}

func (h *AdminAccountHandler) GetAllAccounts(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.DB.MongoDB.Collection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch accounts",
		})
	}
	defer cursor.Close(ctx)

	var accounts []Account
	if err := cursor.All(ctx, &accounts); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse accounts",
		})
	}

	return c.JSON(accounts)
}

// DeleteAccount removes a user and (best-effort) all associated data across collections.
// NOTE: If there are regulatory/audit requirements for retaining orders, you may
// want to anonymize instead of deleting those. For now we fully delete.
func (h *AdminAccountHandler) DeleteAccount(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rawID := c.Params("id")
	if rawID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID is required",
		})
	}
	userID, err := primitive.ObjectIDFromHex(rawID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
			"error":   err.Error(),
		})
	}

	// First ensure user exists
	var existing Account
	if err := h.DB.MongoDB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&existing); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to lookup user",
			"error":   err.Error(),
		})
	}

	// Build deletion tasks (collection pointer, filter description)
	// Each uses {user_id: userID} except users collection which uses _id.
	deletions := []struct {
		name       string
		collection string
		filter     bson.M
	}{
		{"user", "users", bson.M{"_id": userID}},
		{"profile", "user_profiles", bson.M{"user_id": userID}},
		{"preferences", "user_preferences", bson.M{"user_id": userID}},
		{"addresses", "user_addresses", bson.M{"user_id": userID}},
		{"cart items", "cart_items", bson.M{"user_id": userID}},
		{"orders", "orders", bson.M{"user_id": userID}},
		{"inventories", "inventories", bson.M{"user_id": userID}},
		{"reviews", "reviews", bson.M{"user_id": userID}},
		{"wishlists", "wishlists", bson.M{"user_id": userID}},
		{"chat conversations", "chat_conversations", bson.M{"user_id": userID}},
		{"chat messages", "chat_messages", bson.M{"user_id": userID}},
		{"notifications", "notifications", bson.M{"user_id": userID}},
		{"recommendations", "recommendations", bson.M{"user_id": userID}},
		{"recommendation feedbacks", "recommendation_feedbacks", bson.M{"user_id": userID}},
	}

	summary := fiber.Map{}
	for _, d := range deletions {
		coll := h.DB.MongoDB.Collection(d.collection)
		res, derr := coll.DeleteMany(ctx, d.filter)
		if derr != nil {
			summary[d.name] = fmt.Sprintf("error: %v", derr)
			continue
		}
		summary[d.name] = res.DeletedCount
	}

	// Best-effort cache invalidation of known per-user keys.
	// (If adding new user-scoped caches, append here.)
	_ = h.DB.CacheDel(ctx,
		fmt.Sprintf("recommendations:%s", userID.Hex()),
		fmt.Sprintf("wishlist:%s", userID.Hex()),
		fmt.Sprintf("profile:%s", userID.Hex()),
	)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User and related data deleted",
		"data": fiber.Map{
			"userId":  userID.Hex(),
			"summary": summary,
		},
	})
}
func GetAllAccounts(db *mongo.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement logic to fetch accounts from db
		return c.SendString("All accounts endpoint")
	}
}
