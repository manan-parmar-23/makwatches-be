package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/shivam-mishra-20/mak-watches-be/internal/config"
	"github.com/shivam-mishra-20/mak-watches-be/internal/database"
	"github.com/shivam-mishra-20/mak-watches-be/internal/handlers"
)

func main() {
	// Initialize config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}

	log.Printf("Starting server in %s environment...", cfg.Environment)

	// Initialize MongoDB client
	mongoClient, _, err := config.InitMongoDB(cfg)
	if err != nil {
		log.Printf("MongoDB connection error: %v", err)
		log.Printf("Check if MongoDB is running at %s", cfg.MongoURI)
		log.Fatal("Cannot continue without database connection")
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize Redis client
	redisClient, err := config.InitRedis(cfg)
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Continuing without Redis - caching will be disabled")
		log.Println("This is expected if Redis is not configured")
		// Create a nil Redis client - handlers should check for nil
		redisClient = nil
	}
	defer func() {
		if redisClient != nil {
			if err := redisClient.Close(); err != nil {
				log.Printf("Error closing Redis connection: %v", err)
			}
		}
	}()

	// Create database client wrapper
	dbClient := database.NewDBClient(mongoClient, cfg.DatabaseName, redisClient)

	// Initialize Fiber app with custom error handling
	app := fiber.New(fiber.Config{
		AppName:      "Makwatches API",
		ErrorHandler: customErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10MB

	})

	// Configure CORS: allow local dev and production Vercel origin (configurable via env)
	vercelOrigin := cfg.GetEnvOrDefault("VERCEL_ORIGIN", "https://mak-watches.vercel.app")
	devOrigin := cfg.GetEnvOrDefault("DEV_ORIGIN", "http://localhost:4200")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     vercelOrigin + "," + devOrigin,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers",
	}))

	// Setup all routes and middleware
	handlers.SetupRoutes(app, dbClient, cfg)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on port %s in %s mode", cfg.Port, cfg.Environment)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give the server 5 seconds to finish ongoing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

// customErrorHandler provides consistent error responses
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code is 500
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Return JSON response with error details
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": "An error occurred",
		"error":   err.Error(),
	})
}
