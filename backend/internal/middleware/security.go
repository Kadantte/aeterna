package middleware

import (
	"strings"

	"github.com/alpyxn/aeterna/backend/internal/config"
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders returns a middleware that adds security-related HTTP headers.
func SecurityHeaders(cfg config.Config) fiber.Handler {
	isProd := cfg.IsProduction()
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=(), payment=()")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'")

		if isProd {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		if strings.HasPrefix(c.Path(), "/api") {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		}

		return c.Next()
	}
}
