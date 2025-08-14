package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// CookieConfig defines the config for cookie-based authentication middleware
type CookieConfig struct {
	TokenCookieName     string
	RefreshCookieName   string
	JWTSecret           string
	TokenExpiry         time.Duration
	RefreshExpiry       time.Duration
	SecureCookie        bool
	CookieDomain        string
	SessionInfoEndpoint string
}

// DefaultCookieConfig returns a default config for cookie-based authentication
func DefaultCookieConfig() CookieConfig {
	return CookieConfig{
		TokenCookieName:     "access_token",
		RefreshCookieName:   "refresh_token",
		TokenExpiry:         1 * time.Hour,       // Short-lived access token
		RefreshExpiry:       30 * 24 * time.Hour, // 30 days for refresh token
		SecureCookie:        false,               // Set to true in production
		CookieDomain:        "",
		SessionInfoEndpoint: "/auth/session",
	}
}

// SetTokenCookie sets a secure HTTP-only cookie with the JWT token
func SetTokenCookie(c *fiber.Ctx, token string, name string, expiry time.Duration, config CookieConfig) {
	cookie := fiber.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(expiry),
		HTTPOnly: true, // Prevents JavaScript access
		Secure:   config.SecureCookie,
		SameSite: "Lax",
	}

	if config.CookieDomain != "" {
		cookie.Domain = config.CookieDomain
	}

	c.Cookie(&cookie)
}

// ClearTokenCookie clears the auth cookie by setting it to expire immediately
func ClearTokenCookie(c *fiber.Ctx, name string, config CookieConfig) {
	cookie := fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour), // Expired 1 hour ago
		HTTPOnly: true,
		Secure:   config.SecureCookie,
		SameSite: "Lax",
	}

	if config.CookieDomain != "" {
		cookie.Domain = config.CookieDomain
	}

	c.Cookie(&cookie)
}

// CookieAuth is middleware that checks for a valid JWT token in cookies
func CookieAuth(secret string, cookieName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the token from cookie
		tokenString := c.Cookies(cookieName)
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required",
			})
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired token",
				"error":   err.Error(),
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token claims",
			})
		}

		// Store user info in context
		c.Locals("userID", claims["id"])
		c.Locals("role", claims["role"])

		return c.Next()
	}
}
