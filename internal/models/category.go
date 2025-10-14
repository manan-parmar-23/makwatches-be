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
	// Category-level discount fields (optional)
	DiscountPercentage *float64   `json:"discountPercentage,omitempty" bson:"discount_percentage,omitempty"`
	DiscountAmount     *float64   `json:"discountAmount,omitempty" bson:"discount_amount,omitempty"`
	DiscountStartDate  *time.Time `json:"discountStartDate,omitempty" bson:"discount_start_date,omitempty"`
	DiscountEndDate    *time.Time `json:"discountEndDate,omitempty" bson:"discount_end_date,omitempty"`
	CreatedAt          time.Time  `json:"createdAt" bson:"created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" bson:"updated_at"`
}

// Subcategory represents a nested category under a main category
type Subcategory struct {
	ID   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
	// ImageURL is an optional image associated with the subcategory
	ImageURL string `json:"imageUrl,omitempty" bson:"image_url,omitempty"`
	// Subcategory-level discount fields (optional)
	DiscountPercentage *float64   `json:"discountPercentage,omitempty" bson:"discount_percentage,omitempty"`
	DiscountAmount     *float64   `json:"discountAmount,omitempty" bson:"discount_amount,omitempty"`
	DiscountStartDate  *time.Time `json:"discountStartDate,omitempty" bson:"discount_start_date,omitempty"`
	DiscountEndDate    *time.Time `json:"discountEndDate,omitempty" bson:"discount_end_date,omitempty"`
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
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

// UpdateNameRequest used for updating category or subcategory names
// Example:
// { "name": "Women" }
type UpdateNameRequest struct {
	Name string `json:"name"`
}

// UpdateSubcategoryRequest allows updating subcategory fields optionally
// Example:
// { "name": "Sneakers", "imageUrl": "https://..." }
type UpdateSubcategoryRequest struct {
	Name     *string `json:"name"`
	ImageURL *string `json:"imageUrl"`
}

// SubcategoryInput represents input for creating subcategories with optional image
type SubcategoryInput struct {
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

// CategoryDiscountRequest for updating category-level discounts
type CategoryDiscountRequest struct {
	DiscountPercentage *float64   `json:"discountPercentage,omitempty"`
	DiscountAmount     *float64   `json:"discountAmount,omitempty"`
	DiscountStartDate  *time.Time `json:"discountStartDate,omitempty"`
	DiscountEndDate    *time.Time `json:"discountEndDate,omitempty"`
}

// SubcategoryDiscountRequest for updating subcategory-level discounts
type SubcategoryDiscountRequest struct {
	DiscountPercentage *float64   `json:"discountPercentage,omitempty"`
	DiscountAmount     *float64   `json:"discountAmount,omitempty"`
	DiscountStartDate  *time.Time `json:"discountStartDate,omitempty"`
	DiscountEndDate    *time.Time `json:"discountEndDate,omitempty"`
}
