package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/shivam-mishra-20/mak-watches-be/internal/config"
	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"github.com/shivam-mishra-20/mak-watches-be/internal/middleware"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, db *database.DBClient, cfg *config.Config) {
	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,https://makwatches.in",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH",
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
	// upload handler is plain function
	cartHandler := NewCartHandler(db, cfg)
	orderHandler := NewOrderHandler(db, cfg)
	paymentHandler := NewPaymentHandler(db, cfg)
	recHandler := NewRecommendationHandler(db, cfg)
	userProfileHandler := NewUserProfileHandler(db, cfg)
	wishlistHandler := NewWishlistHandler(db, cfg)
	addressBookHandler := NewAddressBookHandler(db, cfg)
	adminAccountHandler := &AdminAccountHandler{DB: db}
	categoryHandler := NewCategoryHandler(db, cfg)
	homeContentHandler := NewHomeContentHandler(db)

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
	// Product reviews (public)
	// GET /products/:id/reviews
	// Use ReviewHandler to serve product-level reviews
	reviewHandler := NewReviewHandler(db, cfg)
	products.Get("/:productId/reviews", reviewHandler.GetProductReviews)

	// Public catalog (optimized) product routes
	catalog := app.Group("/catalog")
	catalog.Get("/products", productHandler.GetPublicProducts)
	catalog.Get("/products/:id", productHandler.GetPublicProductByID)

	// Public category routes (no auth) - read-only for storefront
	app.Get("/categories", categoryHandler.GetPublicCategories)
	app.Get("/categories/:name/subcategories", categoryHandler.GetPublicSubcategories)
	app.Get("/home-content", homeContentHandler.GetHomeContent)

	// Public (or auth-protected) upload route for admin (requires auth+role)
	app.Static("/uploads", "uploads")
	app.Post("/upload", middleware.Auth(cfg.JWTSecret), middleware.Role("admin"), UploadHandler)

	// Admin product routes (must authenticate first, then role check)
	adminProducts := products.Group("/", middleware.Auth(cfg.JWTSecret), middleware.Role("admin"))
	adminProducts.Post("/", productHandler.CreateProduct)
	adminProducts.Put("/:id", productHandler.UpdateProduct)
	adminProducts.Delete("/:id", productHandler.DeleteProduct)

	// Protected routes
	api := app.Group("/", middleware.Auth(cfg.JWTSecret))

	// Review routes (authenticated)
	// POST /reviews -> CreateReview
	reviews := api.Group("/reviews")
	reviews.Post("/", reviewHandler.CreateReview)
	// Optional: allow updating/deleting reviews by owner
	reviews.Put("/:id", reviewHandler.UpdateReview)
	reviews.Delete("/:id", reviewHandler.DeleteReview)
	reviews.Post("/:id/helpful", reviewHandler.MarkReviewHelpful)

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
	// Admin-only: get all orders, update status
	orders.Get("/", middleware.Role("admin"), orderHandler.GetAllOrders)
	orders.Patch("/:orderID/status", middleware.Role("admin"), orderHandler.UpdateOrderStatus)

	// Payment routes
	payments := api.Group("/payments")
	payments.Post("/razorpay/order", paymentHandler.CreateRazorpayOrder)

	// Admin only routes (must authenticate first, then check role)
	admin := app.Group("/admin", middleware.Auth(cfg.JWTSecret), middleware.Role("admin"))
	admin.Get("/accounts", adminAccountHandler.GetAllAccounts)
	admin.Delete("/accounts/:id", adminAccountHandler.DeleteAccount)

	// Settings routes
	settingsHandler := NewSettingsHandler(db.MongoDB)
	admin.Get("/settings", settingsHandler.GetSettings())
	admin.Put("/settings", settingsHandler.UpdateSettings())
	admin.Post("/settings/logo", settingsHandler.UploadLogo())

	// Home content management routes
	adminHome := admin.Group("/home-content")
	adminHome.Get("/hero-slides", homeContentHandler.ListHeroSlides)
	adminHome.Post("/hero-slides", homeContentHandler.CreateHeroSlide)
	adminHome.Put("/hero-slides/:id", homeContentHandler.UpdateHeroSlide)
	adminHome.Delete("/hero-slides/:id", homeContentHandler.DeleteHeroSlide)

	adminHome.Get("/categories", homeContentHandler.ListCategoryCards)
	adminHome.Post("/categories", homeContentHandler.CreateCategoryCard)
	adminHome.Put("/categories/:id", homeContentHandler.UpdateCategoryCard)
	adminHome.Delete("/categories/:id", homeContentHandler.DeleteCategoryCard)

	adminHome.Get("/collections", homeContentHandler.ListCollectionFeatures)
	adminHome.Post("/collections", homeContentHandler.CreateCollectionFeature)
	adminHome.Put("/collections/:id", homeContentHandler.UpdateCollectionFeature)
	adminHome.Delete("/collections/:id", homeContentHandler.DeleteCollectionFeature)

	adminHome.Get("/tech-cards", homeContentHandler.ListTechCards)
	adminHome.Post("/tech-cards", homeContentHandler.CreateTechCard)
	adminHome.Put("/tech-cards/:id", homeContentHandler.UpdateTechCard)
	adminHome.Delete("/tech-cards/:id", homeContentHandler.DeleteTechCard)
	adminHome.Get("/tech-highlight", homeContentHandler.GetTechHighlight)
	adminHome.Put("/tech-highlight", homeContentHandler.UpsertTechHighlight)
	adminHome.Delete("/tech-highlight", homeContentHandler.DeleteTechHighlight)

	// Gallery images management
	adminHome.Get("/gallery", homeContentHandler.ListGalleryImages)
	adminHome.Post("/gallery", homeContentHandler.CreateGalleryImage)
	adminHome.Put("/gallery/:id", homeContentHandler.UpdateGalleryImage)
	adminHome.Delete("/gallery/:id", homeContentHandler.DeleteGalleryImage)

	// Category management routes (/admin/categories)
	adminCategories := admin.Group("/categories")
	adminCategories.Get("/", categoryHandler.GetCategories)
	adminCategories.Post("/", categoryHandler.CreateCategory)
	// Fix missing leading slashes on parameterized routes
	adminCategories.Post("/:id/subcategories", categoryHandler.AddSubcategory)
	adminCategories.Patch("/:id", categoryHandler.UpdateCategoryName)
	adminCategories.Patch("/:categoryId/subcategories/:subId", categoryHandler.UpdateSubcategoryName)
	adminCategories.Delete("/:id", categoryHandler.DeleteCategory)
	adminCategories.Delete("/:categoryId/subcategories/:subId", categoryHandler.DeleteSubcategory)
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

	// Account routes (consolidated user account functionality)
	accountHandler := NewAccountHandler(db, cfg)
	account := api.Group("/account")
	account.Get("/overview", accountHandler.GetAccountOverview)
	account.Get("/reviews", accountHandler.GetAccountReviews)
	account.Delete("/reviews/:id", accountHandler.DeleteAccountReview)
	// Create a review under account scope as well
	account.Post("/reviews", reviewHandler.CreateReview)
	account.Get("/wishlist", accountHandler.GetAccountWishlist)
	account.Delete("/wishlist/:id", accountHandler.RemoveAccountWishlistItem)
	account.Get("/orders", accountHandler.GetAccountOrders)
	account.Get("/orders/:orderID", accountHandler.GetAccountOrder)

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
		"message": "Welcome to Makwatches API",
	})
}
