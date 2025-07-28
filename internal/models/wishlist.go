package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Wishlist represents a wishlist item
type Wishlist struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userId" bson:"user_id"`
	ProductID primitive.ObjectID `json:"productId" bson:"product_id"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updated_at,omitempty"`
}

// WishlistResponse represents a wishlist item with product details
type WishlistResponse struct {
	ID          primitive.ObjectID `json:"id"`
	ProductID   primitive.ObjectID `json:"productId"`
	Name        string             `json:"name"`
	Price       float64            `json:"price"`
	ImageURL    string             `json:"imageUrl"`
	Description string             `json:"description"`
	InStock     bool               `json:"inStock"`
	AddedAt     time.Time          `json:"addedAt"`
}
