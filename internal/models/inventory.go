package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Inventory represents product inventory
type Inventory struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ProductID    primitive.ObjectID `json:"productId" bson:"product_id"`
	Quantity     int                `json:"quantity" bson:"quantity"`
	Reserved     int                `json:"reserved" bson:"reserved"`
	LowStockAlert int               `json:"lowStockAlert" bson:"low_stock_alert"`
	LastRestocked time.Time         `json:"lastRestocked" bson:"last_restocked"`
	CreatedAt    time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updated_at"`
}

// InventoryUpdateRequest is used for updating inventory
type InventoryUpdateRequest struct {
	ProductID    string `json:"productId" validate:"required"`
	Quantity     int    `json:"quantity" validate:"required,min=0"`
	LowStockAlert int    `json:"lowStockAlert" validate:"min=0"`
}

// BulkInventoryUpdateRequest is used for bulk updating inventory
type BulkInventoryUpdateRequest struct {
	Updates []InventoryUpdateRequest `json:"updates" validate:"required,min=1"`
}
