package handlers

import (
	"errors"

	"github.com/alpyxn/aeterna/backend/internal/config"
	"github.com/alpyxn/aeterna/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

var isProduction bool

// SetIsProduction configures whether the handler package runs in production mode.
// Call once from main before registering routes.
func SetIsProduction(cfg config.Config) { isProduction = cfg.IsProduction() }

func currentUserID(c *fiber.Ctx) (string, error) {
	uid, ok := c.Locals("user_id").(string)
	if !ok || uid == "" {
		return "", services.NewAPIError(401, "unauthorized", "Unauthorized", nil)
	}
	return uid, nil
}

func writeError(c *fiber.Ctx, err error) error {
	isProd := isProduction
	var apiErr *services.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.Code
		if code == "" {
			code = "internal_error"
		}
		payload := fiber.Map{
			"error": apiErr.Message,
			"code":  code,
		}
		if !isProd && apiErr.Err != nil {
			payload["detail"] = apiErr.Err.Error()
		}
		return c.Status(apiErr.Status).JSON(payload)
	}
	payload := fiber.Map{
		"error": "Internal server error",
		"code":  "internal_error",
	}
	if !isProd && err != nil {
		payload["detail"] = err.Error()
	}
	return c.Status(500).JSON(payload)
}
