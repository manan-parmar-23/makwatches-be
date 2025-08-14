package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Category represents a top-level category (Men or Women)
type Category struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name"`
	Subcategories []Subcategory      `json:"subcategories" bson:"subcategories"`
	CreatedAt     time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updated_at"`
}

// Subcategory represents a nested category under a main category
type Subcategory struct {
	ID   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
}

// CreateCategoryRequest request body for creating a category
// Example:
//
//	{
//	  "name": "Men",
//	  "subcategories": ["Shirts", "Jeans"]
//	}
type CreateCategoryRequest struct {
	Name          string   `json:"name"`
	Subcategories []string `json:"subcategories"`
}

// AddSubcategoryRequest for adding a new subcategory
// Example:
// { "name": "Shoes" }
type AddSubcategoryRequest struct {
	Name string `json:"name"`
}

// UpdateNameRequest used for updating category or subcategory names
// Example:
// { "name": "Women" }
type UpdateNameRequest struct {
	Name string `json:"name"`
}
