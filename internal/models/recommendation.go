package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecommendationSource represents the source of a recommendation
type RecommendationSource string

const (
	SourceBrowsingHistory RecommendationSource = "browsing_history"
	SourcePurchaseHistory RecommendationSource = "purchase_history"
	SourceSimilarUsers    RecommendationSource = "similar_users"
	SourceTrending        RecommendationSource = "trending"
	SourceFavorites       RecommendationSource = "favorites"
)

// RecommendationFeedback represents user feedback on recommendations
type RecommendationFeedback struct {
	ID              primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	UserID          primitive.ObjectID   `json:"userId" bson:"user_id"`
	ProductID       primitive.ObjectID   `json:"productId" bson:"product_id"`
	RecommendationID primitive.ObjectID   `json:"recommendationId" bson:"recommendation_id"`
	Action          string               `json:"action" bson:"action"` // "click", "add_to_cart", "purchase", "dismiss"
	CreatedAt       time.Time            `json:"createdAt" bson:"created_at"`
}

// RecommendationItem represents a single recommendation
type RecommendationItem struct {
	ID         primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	UserID     primitive.ObjectID   `json:"userId" bson:"user_id"`
	ProductID  primitive.ObjectID   `json:"productId" bson:"product_id"`
	Score      float64              `json:"score" bson:"score"`
	Source     RecommendationSource `json:"source" bson:"source"`
	Reason     string               `json:"reason" bson:"reason"`
	CreatedAt  time.Time            `json:"createdAt" bson:"created_at"`
}

// RecommendationResponse represents a recommendation with product details
type RecommendationResponse struct {
	ID        primitive.ObjectID   `json:"id"`
	ProductID primitive.ObjectID   `json:"productId"`
	Product   Product              `json:"product"`
	Score     float64              `json:"score"`
	Source    RecommendationSource `json:"source"`
	Reason    string               `json:"reason"`
}

// RecommendationFeedbackRequest is used for submitting feedback
type RecommendationFeedbackRequest struct {
	ProductID        string `json:"productId" validate:"required"`
	RecommendationID string `json:"recommendationId" validate:"required"`
	Action           string `json:"action" validate:"required,oneof=click add_to_cart purchase dismiss"`
}
