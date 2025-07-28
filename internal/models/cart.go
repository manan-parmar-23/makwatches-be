package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CartItem represents an item in a user's cart
type CartItem struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userId" bson:"user_id"`
	ProductID primitive.ObjectID `json:"productId" bson:"product_id"`
	Product   *Product           `json:"product,omitempty" bson:"product,omitempty"`
	Quantity  int                `json:"quantity" bson:"quantity"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CartItemRequest represents the data required for adding a product to cart
type CartItemRequest struct {
	ProductID string `json:"productId" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

// CartResponse represents the response for cart operations
type CartResponse struct {
	Items []CartItem `json:"items"`
	Total float64    `json:"total"`
}
