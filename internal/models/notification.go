package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a user notification
type Notification struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"user_id"`
	Type        string             `json:"type" bson:"type"` // "order", "promotion", "product", "system"
	Title       string             `json:"title" bson:"title"`
	Message     string             `json:"message" bson:"message"`
	IsRead      bool               `json:"isRead" bson:"is_read"`
	ReferenceID primitive.ObjectID `json:"referenceId,omitempty" bson:"reference_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
}

// NotificationRequest is used for creating a notification
type NotificationRequest struct {
	UserID      string `json:"userId" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=order promotion product system"`
	Title       string `json:"title" validate:"required"`
	Message     string `json:"message" validate:"required"`
	ReferenceID string `json:"referenceId,omitempty"`
}
