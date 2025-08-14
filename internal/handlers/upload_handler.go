package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UploadHandler handles multipart image uploads and stores them locally (can be adapted for S3)
func UploadHandler(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid multipart form", "error": err.Error()})
	}
	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "No images provided"})
	}

	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to create upload directory", "error": err.Error()})
	}

	urls := make([]string, 0, len(files))
	scheme := "http"
	if c.Secure() {
		scheme = "https"
	}
	host := c.Hostname()
	for _, f := range files {
		name := fmt.Sprintf("%d-%s", time.Now().UnixNano(), f.Filename)
		path := filepath.Join(uploadDir, name)
		if err := c.SaveFile(f, path); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to save file", "error": err.Error()})
		}
		// Build absolute URL (strip default ports for aesthetics)
		h := host
		h = strings.TrimSuffix(strings.TrimSuffix(h, ":80"), ":443")
		urls = append(urls, fmt.Sprintf("%s://%s/uploads/%s", scheme, h, name))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Upload successful", "data": fiber.Map{"urls": urls}})
}
