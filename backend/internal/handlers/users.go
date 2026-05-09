package handlers

import (
	"github.com/alpyxn/aeterna/backend/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// UserHandlers groups administrative user management route handlers.
type UserHandlers struct {
	users ports.UserAdminServicePort
}

func NewUserHandlers(users ports.UserAdminServicePort) *UserHandlers {
	return &UserHandlers{users: users}
}

// List returns all accounts (primary administrator only).
func (h *UserHandlers) List(c *fiber.Ctx) error {
	actorID, err := currentUserID(c)
	if err != nil {
		return writeError(c, err)
	}
	users, err := h.users.List(actorID)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(users)
}

// Delete removes a non-primary user (primary administrator only).
func (h *UserHandlers) Delete(c *fiber.Ctx) error {
	actorID, err := currentUserID(c)
	if err != nil {
		return writeError(c, err)
	}
	targetID := c.Params("id")
	if err := h.users.Delete(actorID, targetID); err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}
