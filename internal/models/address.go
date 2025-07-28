package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Address represents a user's shipping address
type Address struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"user_id"`
	Name        string             `json:"name" bson:"name"`
	Street      string             `json:"street" bson:"street"`
	City        string             `json:"city" bson:"city"`
	State       string             `json:"state" bson:"state"`
	ZipCode     string             `json:"zipCode" bson:"zip_code"`
	Country     string             `json:"country" bson:"country"`
	Phone       string             `json:"phone" bson:"phone"`
	IsDefault   bool               `json:"isDefault" bson:"is_default"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// AddressRequest is used for creating or updating an address
type AddressRequest struct {
	Name      string `json:"name" validate:"required"`
	Street    string `json:"street" validate:"required"`
	City      string `json:"city" validate:"required"`
	State     string `json:"state" validate:"required"`
	ZipCode   string `json:"zipCode" validate:"required"`
	Country   string `json:"country" validate:"required"`
	Phone     string `json:"phone" validate:"required"`
	IsDefault bool   `json:"isDefault"`
}
