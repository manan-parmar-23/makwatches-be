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
	// Discount fields (optional)
	DiscountPercentage *float64   `json:"discountPercentage,omitempty" bson:"discount_percentage,omitempty"` // Percentage discount (0-100)
	DiscountAmount     *float64   `json:"discountAmount,omitempty" bson:"discount_amount,omitempty"`         // Fixed amount discount
	DiscountStartDate  *time.Time `json:"discountStartDate,omitempty" bson:"discount_start_date,omitempty"`  // When discount starts
	DiscountEndDate    *time.Time `json:"discountEndDate,omitempty" bson:"discount_end_date,omitempty"`      // When discount ends
	CreatedAt          time.Time  `json:"createdAt" bson:"created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" bson:"updated_at"`
}

// IsDiscountActive checks if the product has an active discount
func (p *Product) IsDiscountActive() bool {
	now := time.Now()

	// Check if discount exists
	if p.DiscountPercentage == nil && p.DiscountAmount == nil {
		return false
	}

	// Check date range if specified
	if p.DiscountStartDate != nil && now.Before(*p.DiscountStartDate) {
		return false
	}
	if p.DiscountEndDate != nil && now.After(*p.DiscountEndDate) {
		return false
	}

	return true
}

// GetFinalPrice returns the price after applying active discount
func (p *Product) GetFinalPrice() float64 {
	if !p.IsDiscountActive() {
		return p.Price
	}

	// Apply percentage discount first if exists
	if p.DiscountPercentage != nil && *p.DiscountPercentage > 0 {
		discount := p.Price * (*p.DiscountPercentage / 100.0)
		return p.Price - discount
	}

	// Apply fixed amount discount
	if p.DiscountAmount != nil && *p.DiscountAmount > 0 {
		finalPrice := p.Price - *p.DiscountAmount
		if finalPrice < 0 {
			return 0
		}
		return finalPrice
	}

	return p.Price
}

// GetDiscountAmount returns the discount amount applied
func (p *Product) GetDiscountAmount() float64 {
	if !p.IsDiscountActive() {
		return 0
	}
	return p.Price - p.GetFinalPrice()
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
