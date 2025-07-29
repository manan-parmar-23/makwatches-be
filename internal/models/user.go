package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name"`
	Email         string             `json:"email" bson:"email"`
	Password      string             `json:"-" bson:"password"` // Password is not included in JSON responses
	Role          string             `json:"role" bson:"role"`
	GoogleID      string             `json:"googleId,omitempty" bson:"google_id,omitempty"`
	Picture       string             `json:"picture,omitempty" bson:"picture,omitempty"`
	AuthProvider  string             `json:"authProvider" bson:"auth_provider"` // "local", "google", etc.
	CreatedAt     time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updated_at"`
}

// UserResponse is the response returned after user actions (omits sensitive info)
type UserResponse struct {
	ID           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Email        string             `json:"email"`
	Role         string             `json:"role"`
	Picture      string             `json:"picture,omitempty"`
	AuthProvider string             `json:"authProvider,omitempty"`
}

// RegisterRequest represents the data required for user registration
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents the data required for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// GoogleUser represents the data received from Google OAuth
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}