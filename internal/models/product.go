package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a product in the system
type Product struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Brand        string             `json:"brand,omitempty" bson:"brand,omitempty"`
	Description  string             `json:"description" bson:"description"`
	Price        float64            `json:"price" bson:"price"`
	Category     string             `json:"category" bson:"category"`
	MainCategory string             `json:"mainCategory,omitempty" bson:"main_category,omitempty"`
	Subcategory  string             `json:"subcategory,omitempty" bson:"subcategory,omitempty"`
	ImageURL     string             `json:"imageUrl" bson:"image_url"` // Main image (legacy support)
	Images       []string           `json:"images" bson:"images"`      // Multiple S3 image URLs
	Stock        int                `json:"stock" bson:"stock"`
	CreatedAt    time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updated_at"`
}

// ProductFilters represents filters for product queries
type ProductFilters struct {
	Category string   `query:"category"`
	MinPrice *float64 `query:"minPrice"`
	MaxPrice *float64 `query:"maxPrice"`
	SortBy   string   `query:"sortBy"`
	Order    string   `query:"order"`
	Page     int      `query:"page"`
	Limit    int      `query:"limit"`
}
