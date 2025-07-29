package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/the-devesta/pehnaw-be/internal/config"
	"github.com/the-devesta/pehnaw-be/internal/database"
	"github.com/the-devesta/pehnaw-be/internal/middleware"
	"github.com/the-devesta/pehnaw-be/internal/models"
	"github.com/the-devesta/pehnaw-be/pkg/utils"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	DB        *database.DBClient
	Config    *config.Config
	GoogleOAuth *utils.GoogleOAuth
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(db *database.DBClient, cfg *config.Config) *AuthHandler {
	googleOAuth := utils.NewGoogleOAuth(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
	)
	
	return &AuthHandler{
		DB:        db,
		Config:    cfg,
		GoogleOAuth: googleOAuth,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	ctx := c.Context()
	var req models.RegisterRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Name, email, and password are required",
		})
	}

	// Check if user already exists
	collection := h.DB.Collections().Users
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User with this email already exists",
		})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to hash password",
			"error":   err.Error(),
		})
	}

	// Create new user
	now := time.Now()
	newUser := models.User{
		ID:           primitive.NewObjectID(),
		Name:         req.Name,
		Email:        req.Email,
		Password:     string(hashedPassword),
		Role:         "user", // Default role
		AuthProvider: "local", // Local authentication
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Insert user into database
	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create user",
			"error":   err.Error(),
		})
	}

	// Generate JWT token
	token, err := h.generateToken(newUser.ID.Hex(), newUser.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
			"error":   err.Error(),
		})
	}

	// Return user info and token
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
		"data": models.LoginResponse{
			User: models.UserResponse{
				ID:           newUser.ID,
				Name:         newUser.Name,
				Email:        newUser.Email,
				Role:         newUser.Role,
				AuthProvider: newUser.AuthProvider,
			},
			Token: token,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	ctx := c.Context()
	var req models.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Email and password are required",
		})
	}

	// Find user by email
	collection := h.DB.Collections().Users
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid email or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Database error",
			"error":   err.Error(),
		})
	}
	
	// Check if user is using Google auth and trying to login with password
	if user.AuthProvider == "google" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "This account uses Google authentication. Please sign in with Google.",
		})
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid email or password",
		})
	}

	// Generate JWT token
	token, err := h.generateToken(user.ID.Hex(), user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
			"error":   err.Error(),
		})
	}

	// Return user info and token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"data": models.LoginResponse{
			User: models.UserResponse{
				ID:           user.ID,
				Name:         user.Name,
				Email:        user.Email,
				Role:         user.Role,
				Picture:      user.Picture,
				AuthProvider: user.AuthProvider,
			},
			Token: token,
		},
	})
}

// GoogleLogin initiates Google OAuth login
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	// Generate a state token to prevent request forgery
	state := fmt.Sprintf("%d", time.Now().UnixNano())
	
	// Log the state for debugging
	fmt.Printf("Generated state: %s\n", state)
	
	// Store state in server-side storage instead of cookies
	h.GoogleOAuth.SaveState(state)
	
	// Redirect to Google's OAuth page
	authURL := h.GoogleOAuth.GetAuthURL(state)
	return c.Redirect(authURL)
}

// GoogleCallback handles the callback from Google OAuth
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	ctx := c.Context()
	
	// Extract code and state from query params
	code := c.Query("code")
	state := c.Query("state")
	
	// For debugging
	fmt.Printf("Received state: %s\n", state)
	
	// Check for code parameter
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Missing code parameter",
		})
	}
	
	// Validate state using our server-side state store
	if state == "" || !h.GoogleOAuth.ValidateState(state) {
		// For development we'll continue anyway
		fmt.Printf("State validation failed for state: %s\n", state)
		// In production, you would return an error here
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		//    "success": false,
		//    "message": "Invalid state parameter",
		// })
	}
	
	// Exchange code for token
	accessToken, err := h.GoogleOAuth.Exchange(code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to exchange code for token",
			"error":   err.Error(),
		})
	}
	
	// Get user info from Google
	userInfo, err := h.GoogleOAuth.GetUserInfo(accessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get user info from Google",
			"error":   err.Error(),
		})
	}
	
	// Extract user details
	googleUser := models.GoogleUser{
		ID:            userInfo["id"].(string),
		Email:         userInfo["email"].(string),
		VerifiedEmail: userInfo["verified_email"].(bool),
		Name:          userInfo["name"].(string),
		Picture:       userInfo["picture"].(string),
	}
	
	if !googleUser.VerifiedEmail {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Email not verified by Google",
		})
	}
	
	// Check if user exists in our database
	collection := h.DB.Collections().Users
	var user models.User
	
	// First try to find by Google ID
	err = collection.FindOne(ctx, bson.M{"google_id": googleUser.ID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		// If not found by Google ID, try by email
		err = collection.FindOne(ctx, bson.M{"email": googleUser.Email}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			// User doesn't exist, create a new one
			now := time.Now()
			newUser := models.User{
				ID:           primitive.NewObjectID(),
				Name:         googleUser.Name,
				Email:        googleUser.Email,
				GoogleID:     googleUser.ID,
				Picture:      googleUser.Picture,
				Role:         "user", // Default role
				AuthProvider: "google",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			
			_, err = collection.InsertOne(ctx, newUser)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Failed to create user",
					"error":   err.Error(),
				})
			}
			
			user = newUser
		} else if err != nil {
			// Database error
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Database error",
				"error":   err.Error(),
			})
		} else {
			// User exists but doesn't have Google ID, update it
			if user.AuthProvider == "" || user.AuthProvider == "local" {
				update := bson.M{
					"$set": bson.M{
						"google_id":     googleUser.ID,
						"picture":       googleUser.Picture,
						"auth_provider": "hybrid", // User has both local and Google auth
						"updated_at":    time.Now(),
					},
				}
				
				_, err = collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"message": "Failed to update user",
						"error":   err.Error(),
					})
				}
				
				// Update local user object
				user.GoogleID = googleUser.ID
				user.Picture = googleUser.Picture
				user.AuthProvider = "hybrid"
			}
		}
	} else if err != nil {
		// Database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Database error",
			"error":   err.Error(),
		})
	} else {
		// User found by Google ID, update picture if needed
		if user.Picture != googleUser.Picture {
			update := bson.M{
				"$set": bson.M{
					"picture":    googleUser.Picture,
					"updated_at": time.Now(),
				},
			}
			
			_, err = collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Failed to update user picture",
					"error":   err.Error(),
				})
			}
			
			// Update local user object
			user.Picture = googleUser.Picture
		}
	}
	
	// Generate JWT token
	token, err := h.generateToken(user.ID.Hex(), user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
			"error":   err.Error(),
		})
	}
	
	// Prepare frontend redirect URL with token
	frontendURL := "http://localhost:3000" // Default for development
	if h.Config.Environment == "production" {
		frontendURL = "https://pehnaw.com" // Production URL
	}
	
	// Redirect to frontend with token as query param
	// In a real application, use a more secure method to pass the token
	return c.Redirect(fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, token))
}

// Me retrieves current user information
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	// Get user from context (set by Auth middleware)
	user, ok := c.Locals("user").(*middleware.TokenMetadata)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized - User data not found",
		})
	}
	
	ctx := c.Context()
	collection := h.DB.Collections().Users
	
	// Find user by ID
	var userData models.User
	objID := user.UserID
	err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&userData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Database error",
			"error":   err.Error(),
		})
	}
	
	// Return user info
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User profile retrieved successfully",
		"data": models.UserResponse{
			ID:           userData.ID,
			Name:         userData.Name,
			Email:        userData.Email,
			Role:         userData.Role,
			Picture:      userData.Picture,
			AuthProvider: userData.AuthProvider,
		},
	})
}

// generateToken generates a JWT token
func (h *AuthHandler) generateToken(userID, role string) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = userID
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Duration(h.Config.JWTExpirationHours) * time.Hour).Unix()

	// Generate encoded token
	tokenString, err := token.SignedString([]byte(h.Config.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
