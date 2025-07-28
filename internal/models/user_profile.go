package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserProfile extends the base user model with additional profile information
type UserProfile struct {
	UserID      primitive.ObjectID `json:"userId" bson:"user_id"`
	DateOfBirth *time.Time         `json:"dateOfBirth,omitempty" bson:"date_of_birth,omitempty"`
	Gender      string             `json:"gender,omitempty" bson:"gender,omitempty"`
	Phone       string             `json:"phone,omitempty" bson:"phone,omitempty"`
	AvatarURL   string             `json:"avatarUrl,omitempty" bson:"avatar_url,omitempty"`
	Bio         string             `json:"bio,omitempty" bson:"bio,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// UserPreferences represents user preferences for recommendations
type UserPreferences struct {
	ID               primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	UserID           primitive.ObjectID   `json:"userId" bson:"user_id"`
	FavoriteCategories []string           `json:"favoriteCategories" bson:"favorite_categories"`
	FavoriteBrands    []string           `json:"favoriteBrands" bson:"favorite_brands"`
	SizePreferences   map[string]string  `json:"sizePreferences" bson:"size_preferences"`
	ColorPreferences  []string           `json:"colorPreferences" bson:"color_preferences"`
	PriceRange        []float64          `json:"priceRange" bson:"price_range"` // [min, max]
	CreatedAt         time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updated_at"`
}

// ProfileUpdateRequest is used for updating user profile
type ProfileUpdateRequest struct {
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
	Gender      string     `json:"gender,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	AvatarURL   string     `json:"avatarUrl,omitempty"`
	Bio         string     `json:"bio,omitempty"`
}

// PreferencesUpdateRequest is used for updating user preferences
type PreferencesUpdateRequest struct {
	FavoriteCategories []string          `json:"favoriteCategories,omitempty"`
	FavoriteBrands     []string          `json:"favoriteBrands,omitempty"`
	SizePreferences    map[string]string `json:"sizePreferences,omitempty"`
	ColorPreferences   []string          `json:"colorPreferences,omitempty"`
	PriceRange         []float64         `json:"priceRange,omitempty"`
}
