package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/the-devesta/pehnaw-be/internal/config"
	"github.com/the-devesta/pehnaw-be/internal/database"
	"github.com/the-devesta/pehnaw-be/internal/handlers"
	"github.com/the-devesta/pehnaw-be/internal/middleware"
)

// SetupAccountRoutes sets up the routes for the account endpoints
func SetupAccountRoutes(app *fiber.App, db *database.DBClient, cfg *config.Config) {
	accountHandler := handlers.NewAccountHandler(db, cfg)

	// Create a group for account routes with authentication middleware
	accountGroup := app.Group("/account", middleware.Auth(cfg.JWTSecret))

	// Account overview
	accountGroup.Get("/overview", accountHandler.GetAccountOverview)

	// Profile management
	accountGroup.Get("/profile", handlers.NewUserProfileHandler(db, cfg).GetProfile)
	accountGroup.Put("/profile", accountHandler.UpdateAccountProfile)

	// Reviews management
	accountGroup.Get("/reviews", accountHandler.GetAccountReviews)
	accountGroup.Delete("/reviews/:id", accountHandler.DeleteAccountReview)

	// Wishlist management
	accountGroup.Get("/wishlist", accountHandler.GetAccountWishlist)
	accountGroup.Delete("/wishlist/:id", accountHandler.RemoveAccountWishlistItem)

	// Orders management
	accountGroup.Get("/orders", accountHandler.GetAccountOrders)
	accountGroup.Get("/orders/:orderID", accountHandler.GetAccountOrder)
}
