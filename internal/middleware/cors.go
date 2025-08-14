package middleware

import "github.com/gofiber/fiber/v2"

// CORSConfig holds allowed origins for future extensibility
var AllowedOrigins = []string{
	"http://localhost:3000",
	// Add more origins here as needed
}

func CORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		allowed := false
		for _, o := range AllowedOrigins {
			if o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			c.Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Set("Access-Control-Allow-Origin", AllowedOrigins[0]) // fallback to first allowed
		}
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Set("Access-Control-Allow-Credentials", "true")
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
}
