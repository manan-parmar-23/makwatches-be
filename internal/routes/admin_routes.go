package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shivam-mishra-20/mak-watches-be/internal/handlers"
	"go.mongodb.org/mongo-driver/mongo"
)

func AdminRoutes(app *fiber.App, db *mongo.Database) {
	admin := app.Group("/admin")

	// Other admin routes...
	admin.Get("/accounts", handlers.GetAllAccounts(db))

	// Settings routes
	settingsHandler := handlers.NewSettingsHandler(db)
	admin.Get("/settings", settingsHandler.GetSettings())
	admin.Put("/settings", settingsHandler.UpdateSettings())
	admin.Post("/settings/logo", settingsHandler.UploadLogo())
}
