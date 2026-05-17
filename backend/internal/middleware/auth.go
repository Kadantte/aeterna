package middleware

import (
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/alpyxn/aeterna/backend/internal/config"
	"github.com/alpyxn/aeterna/backend/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// MasterAuth returns a middleware that validates the session cookie and enforces the origin allowlist.
func MasterAuth(auth ports.AuthServicePort, cfg config.Config) fiber.Handler {
	allowedOrigins := cfg.AllowedOriginsOrDefault()
	isProd := cfg.IsProduction()
	cookieSecureMode := cfg.Auth.CookieSecureMode
	return func(c *fiber.Ctx) error {
		if token := c.Cookies("aeterna_session"); token != "" {
			userID, err := auth.VerifySessionToken(token)
			if err == nil {
				if err := enforceOriginAllowlist(c, allowedOrigins, isProd); err != nil {
					return err
				}
				c.Locals("user_id", userID)
				return c.Next()
			}
			clearSessionCookieWith(c, cookieSecureMode)
		}

		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized access. Session required.",
			"code":  "unauthorized",
		})
	}
}

func enforceOriginAllowlist(c *fiber.Ctx, allowedOrigins string, isProd bool) error {
	origin := strings.TrimSpace(c.Get("Origin"))

	if !isProd {
		slog.Info("Origin check", "origin", origin, "allowed", allowedOrigins, "referer", c.Get("Referer"))
	}

	if allowedOrigins == "*" {
		return nil
	}

	if origin == "" {
		referer := strings.TrimSpace(c.Get("Referer"))
		if referer != "" {
			parsed, err := url.Parse(referer)
			if err == nil && parsed.Host != "" {
				origin = parsed.Scheme + "://" + parsed.Host
			}
		}
	}

	if origin == "" {
		if !isProd {
			return nil
		}
		return c.Status(403).JSON(fiber.Map{
			"error": "Origin required",
			"code":  "origin_required",
		})
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return c.Status(403).JSON(fiber.Map{
			"error": "Invalid origin",
			"code":  "invalid_origin",
		})
	}

	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173"
	}

	for _, entry := range strings.Split(allowedOrigins, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if origin == entry {
			return nil
		}
	}

	return c.Status(403).JSON(fiber.Map{
		"error": "Origin not allowed",
		"code":  "origin_not_allowed",
	})
}

func clearSessionCookieWith(c *fiber.Ctx, cookieSecureMode string) {
	secure := ShouldUseSecureCookie(c, cookieSecureMode)
	c.Cookie(&fiber.Cookie{
		Name:     "aeterna_session",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Path:     "/",
		HTTPOnly: true,
		Secure:   secure,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

// ShouldUseSecureCookie returns true when the session cookie should be flagged Secure.
func ShouldUseSecureCookie(c *fiber.Ctx, cookieSecureMode string) bool {
	switch cookieSecureMode {
	case "always":
		return true
	case "never":
		return false
	}
	return requestIsHTTPS(c)
}

func requestIsHTTPS(c *fiber.Ctx) bool {
	if c.Protocol() == "https" {
		return true
	}
	forwardedProto := strings.ToLower(strings.TrimSpace(c.Get("X-Forwarded-Proto")))
	if forwardedProto == "" {
		return false
	}
	first := strings.TrimSpace(strings.Split(forwardedProto, ",")[0])
	return first == "https"
}
