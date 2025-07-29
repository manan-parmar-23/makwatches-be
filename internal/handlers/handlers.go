package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/the-devesta/pehnaw-be/internal/config"
	"github.com/the-devesta/pehnaw-be/internal/database"
	"github.com/the-devesta/pehnaw-be/internal/middleware"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, db *database.DBClient, cfg *config.Config) {
	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,https://pehnaw.com",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	
	// Health check endpoint
	app.Get("/health", HealthHandler)

	// Welcome endpoint
	app.Get("/welcome", WelcomeHandler)
	
	// Initialize handlers
	authHandler := NewAuthHandler(db, cfg)
	productHandler := NewProductHandler(db, cfg)
	cartHandler := NewCartHandler(db, cfg)
	orderHandler := NewOrderHandler(db, cfg)
	recHandler := NewRecommendationHandler(db, cfg)
	userProfileHandler := NewUserProfileHandler(db, cfg)
	wishlistHandler := NewWishlistHandler(db, cfg)
	addressBookHandler := NewAddressBookHandler(db, cfg)
	
	// Auth routes
	auth := app.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/google", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)
	
	// Product routes
	products := app.Group("/products")
	products.Get("/", productHandler.GetProducts)
	products.Get("/:id", productHandler.GetProductByID)
	
	// Admin product routes
	adminProducts := products.Group("/", middleware.Role("admin"))
	adminProducts.Post("/", productHandler.CreateProduct)
	adminProducts.Put("/:id", productHandler.UpdateProduct)
	adminProducts.Delete("/:id", productHandler.DeleteProduct)
	
	// Protected routes
	api := app.Group("/", middleware.Auth(cfg.JWTSecret))
	
	// User "me" endpoint
	api.Get("/me", authHandler.Me)
	
	// Cart routes
	cart := api.Group("/cart")
	cart.Post("/", cartHandler.AddToCart)
	cart.Get("/:userID", cartHandler.GetCart)
	cart.Delete("/:userID/:productID", cartHandler.RemoveFromCart)
	
	// Order routes
	orders := api.Group("/orders")
	orders.Get("/user/:userID", orderHandler.GetOrders)
	orders.Get("/:orderID", orderHandler.GetOrder)
	orders.Post("/:orderID/cancel", orderHandler.CancelOrder)
	
	// Admin only routes
	adminOrders := orders.Group("/", middleware.Role("admin"))
	adminOrders.Patch("/:orderID/status", orderHandler.UpdateOrderStatus)
	
	// Checkout route
	api.Post("/checkout", orderHandler.Checkout)
	
	// Recommendation routes
	recommendations := api.Group("/recommendations")
	recommendations.Get("/", recHandler.GetRecommendations)
	recommendations.Post("/feedback", recHandler.SubmitFeedback)
	
	// User profile routes
	profiles := api.Group("/profiles")
	profiles.Get("/", userProfileHandler.GetProfile)
	profiles.Put("/", userProfileHandler.UpdateProfile)
	
	// User preferences routes
	preferences := api.Group("/preferences")
	preferences.Put("/", userProfileHandler.UpdatePreferences)
	
	// Wishlist routes
	wishlist := api.Group("/wishlist")
	wishlist.Get("/", wishlistHandler.GetWishlist)
	wishlist.Post("/", wishlistHandler.AddToWishlist)
	wishlist.Delete("/:id", wishlistHandler.RemoveFromWishlist)
	wishlist.Delete("/", wishlistHandler.ClearWishlist)
	
	// Address book routes
	addresses := api.Group("/addresses")
	addresses.Get("/", addressBookHandler.GetAddresses)
	addresses.Get("/:id", addressBookHandler.GetAddress)
	addresses.Post("/", addressBookHandler.CreateAddress)
	addresses.Put("/:id", addressBookHandler.UpdateAddress)
	addresses.Delete("/:id", addressBookHandler.DeleteAddress)
	addresses.Put("/:id/default", addressBookHandler.SetDefaultAddress)
}

// HealthHandler handles the health check endpoint
func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Server is healthy",
	})
}

// WelcomeHandler handles the welcome endpoint
func WelcomeHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Welcome to Pehnaw API",
	})
}
