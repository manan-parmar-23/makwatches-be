package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shivam-mishra-20/mak-watches-be/internal/config"
	"github.com/shivam-mishra-20/mak-watches-be/internal/firebase"
)

// UploadHandler handles multipart image uploads and stores them in Firebase Storage
func UploadHandler(c *fiber.Ctx) error {
	log.Println("[UPLOAD] Starting upload process...")

	// fallback: try loading config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("[UPLOAD] Failed to load config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to load config", "error": err.Error()})
	}
	log.Printf("[UPLOAD] Config loaded. Firebase credentials: %s, bucket: %s", cfg.FirebaseCredentialsPath, cfg.FirebaseBucketName)

	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("[UPLOAD] Multipart form error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid multipart form", "error": err.Error()})
	}
	files := form.File["images"]
	if len(files) == 0 {
		log.Println("[UPLOAD] No images provided")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "No images provided"})
	}
	log.Printf("[UPLOAD] Found %d files to upload", len(files))

	ctx := context.Background()
	fbClient, err := firebase.NewFirebaseClient(ctx, cfg.FirebaseCredentialsPath, cfg.FirebaseBucketName)
	useLocalFallback := false
	if err != nil {
		log.Printf("[UPLOAD] Failed to init Firebase client: %v", err)
		if cfg.Environment == "development" || cfg.Environment == "dev" || cfg.Environment == "local" {
			log.Println("[UPLOAD] Development mode detected; falling back to local file storage under ./uploads")
			useLocalFallback = true
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to init Firebase client", "error": err.Error()})
		}
	} else {
		log.Println("[UPLOAD] Firebase client initialized successfully")
	}

	urls := make([]string, 0, len(files))
	for i, f := range files {
		log.Printf("[UPLOAD] Processing file %d/%d: %s", i+1, len(files), f.Filename)
		file, err := f.Open()
		if err != nil {
			log.Printf("[UPLOAD] Failed to open file %s: %v", f.Filename, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to open file", "error": err.Error()})
		}
		defer file.Close()

		if useLocalFallback {
			// Ensure uploads directory exists
			if err := os.MkdirAll("uploads", 0o755); err != nil {
				log.Printf("[UPLOAD] Failed to create uploads directory: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to prepare uploads directory", "error": err.Error()})
			}
			// Unique filename similar to Firebase pathing
			unique := fmt.Sprintf("%d-%s", time.Now().UnixNano(), f.Filename)
			destPath := filepath.Join("uploads", unique)
			if err := c.SaveFile(f, destPath); err != nil {
				log.Printf("[UPLOAD] Failed to save file locally: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to save file", "error": err.Error()})
			}
			base := c.BaseURL()
			url := base + "/uploads/" + unique
			log.Printf("[UPLOAD] Saved %s locally at %s (URL: %s)", f.Filename, destPath, url)
			urls = append(urls, url)
		} else {
			log.Printf("[UPLOAD] Uploading file %s to Firebase...", f.Filename)
			url, err := fbClient.UploadFile(ctx, file, f.Filename)
			if err != nil {
				log.Printf("[UPLOAD] Failed to upload file %s to Firebase: %v", f.Filename, err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to upload to Firebase", "error": err.Error()})
			}
			log.Printf("[UPLOAD] Successfully uploaded %s, URL: %s", f.Filename, url)
			urls = append(urls, url)
		}
	}

	log.Printf("[UPLOAD] Upload process completed successfully. URLs: %v", urls)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Upload successful", "data": fiber.Map{"urls": urls}})
}
