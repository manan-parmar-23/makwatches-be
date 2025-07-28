package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Review represents a product review
type Review struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"user_id"`
	ProductID   primitive.ObjectID `json:"productId" bson:"product_id"`
	Rating      float64            `json:"rating" bson:"rating"`
	Title       string             `json:"title" bson:"title"`
	Comment     string             `json:"comment" bson:"comment"`
	PhotoURLs   []string           `json:"photoUrls,omitempty" bson:"photo_urls,omitempty"`
	Helpful     int                `json:"helpful" bson:"helpful"`
	Verified    bool               `json:"verified" bson:"verified"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// ReviewRequest is used for creating or updating a review
type ReviewRequest struct {
	ProductID  string   `json:"productId" validate:"required"`
	Rating     float64  `json:"rating" validate:"required,min=1,max=5"`
	Title      string   `json:"title" validate:"required"`
	Comment    string   `json:"comment" validate:"required"`
	PhotoURLs  []string `json:"photoUrls,omitempty"`
}

// ReviewResponse represents a review with user information
type ReviewResponse struct {
	ID         primitive.ObjectID `json:"id"`
	UserID     primitive.ObjectID `json:"userId"`
	UserName   string             `json:"userName"`
	ProductID  primitive.ObjectID `json:"productId"`
	Rating     float64            `json:"rating"`
	Title      string             `json:"title"`
	Comment    string             `json:"comment"`
	PhotoURLs  []string           `json:"photoUrls,omitempty"`
	Helpful    int                `json:"helpful"`
	Verified   bool               `json:"verified"`
	CreatedAt  time.Time          `json:"createdAt"`
}

// ReviewSummary represents the summary of reviews for a product
type ReviewSummary struct {
	ProductID    primitive.ObjectID `json:"productId"`
	AverageRating float64           `json:"averageRating"`
	TotalReviews  int               `json:"totalReviews"`
	RatingCounts  map[int]int       `json:"ratingCounts"` // map of rating to count
}
