package handlers

import (
	"time"

	"github.com/alpyxn/aeterna/backend/internal/config"
	"github.com/alpyxn/aeterna/backend/internal/middleware"
	"github.com/alpyxn/aeterna/backend/internal/ports"
	"github.com/alpyxn/aeterna/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

type registerRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	OwnerEmail string `json:"owner_email"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthHandlers groups all authentication-related route handlers.
type AuthHandlers struct {
	auth ports.AuthServicePort
	cfg  config.Config
}

func NewAuthHandlers(auth ports.AuthServicePort, cfg config.Config) *AuthHandlers {
	return &AuthHandlers{auth: auth, cfg: cfg}
}

func (h *AuthHandlers) SetupStatus(c *fiber.Ctx) error {
	configured, err := h.auth.IsConfigured()
	if err != nil {
		return writeError(c, err)
	}
	out := fiber.Map{"configured": configured}
	if configured {
		allow, err := h.auth.AdditionalRegistrationOpen()
		if err != nil {
			return writeError(c, err)
		}
		out["allow_registration"] = allow
	} else {
		out["allow_registration"] = false
	}
	return c.JSON(out)
}

func (h *AuthHandlers) SetupMasterPassword(c *fiber.Ctx) error {
	configured, err := h.auth.IsConfigured()
	if err != nil {
		return writeError(c, err)
	}
	if configured {
		return writeError(c, services.NewAPIError(400, "already_configured", "An account already exists. Sign in instead.", nil))
	}

	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, services.BadRequest("Invalid request body", err))
	}
	if req.Email == "" && req.Password != "" {
		req.Email = req.OwnerEmail
	}
	recoveryKey, user, err := h.auth.RegisterFirstUser(req.Email, req.Password, req.OwnerEmail)
	if err != nil {
		return writeError(c, err)
	}
	if err := h.issueSessionCookie(c, user.ID); err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "recovery_key": recoveryKey})
}

func (h *AuthHandlers) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, services.BadRequest("Invalid request body", err))
	}

	configured, err := h.auth.IsConfigured()
	if err != nil {
		return writeError(c, err)
	}
	var recoveryKey string
	var userID string
	if !configured {
		rk, u, err := h.auth.RegisterFirstUser(req.Email, req.Password, req.OwnerEmail)
		if err != nil {
			return writeError(c, err)
		}
		recoveryKey, userID = rk, u.ID
	} else {
		rk, u, err := h.auth.RegisterAdditionalUser(req.Email, req.Password, req.OwnerEmail)
		if err != nil {
			return writeError(c, err)
		}
		recoveryKey, userID = rk, u.ID
	}
	if err := h.issueSessionCookie(c, userID); err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "recovery_key": recoveryKey})
}

func (h *AuthHandlers) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, services.BadRequest("Invalid request body", err))
	}
	user, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		middleware.RecordFailedLogin(c.IP())
		return writeError(c, err)
	}
	middleware.RecordSuccessfulLogin(c.IP())
	if err := h.issueSessionCookie(c, user.ID); err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *AuthHandlers) ResetMasterPassword(c *fiber.Ctx) error {
	var req struct {
		Email       string `json:"email"`
		RecoveryKey string `json:"recovery_key"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, services.BadRequest("Invalid request body", err))
	}

	newRecoveryKey, err := h.auth.ResetPasswordWithRecovery(req.Email, req.RecoveryKey, req.NewPassword)
	if err != nil {
		middleware.RecordFailedLogin(c.IP())
		return writeError(c, err)
	}
	middleware.RecordSuccessfulLogin(c.IP())

	user, err := h.auth.Login(req.Email, req.NewPassword)
	if err != nil {
		return writeError(c, err)
	}
	if err := h.issueSessionCookie(c, user.ID); err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "recovery_key": newRecoveryKey})
}

// VerifyMasterPassword is kept for backward compatibility: same as Login.
func (h *AuthHandlers) VerifyMasterPassword(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, services.BadRequest("Invalid request body", err))
	}
	if req.Email == "" {
		return writeError(c, services.BadRequest("Email is required", nil))
	}
	user, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		middleware.RecordFailedLogin(c.IP())
		return writeError(c, err)
	}
	middleware.RecordSuccessfulLogin(c.IP())
	if err := h.issueSessionCookie(c, user.ID); err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *AuthHandlers) SessionStatus(c *fiber.Ctx) error {
	token := c.Cookies("aeterna_session")
	userID, err := h.auth.VerifySessionToken(token)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(fiber.Map{"authorized": true, "user_id": userID})
}

func (h *AuthHandlers) Logout(c *fiber.Ctx) error {
	h.clearSessionCookie(c)
	return c.JSON(fiber.Map{"success": true})
}

func (h *AuthHandlers) issueSessionCookie(c *fiber.Ctx, userID string) error {
	token, exp, err := h.auth.IssueSessionToken(userID)
	if err != nil {
		return err
	}
	secure := middleware.ShouldUseSecureCookie(c, h.cfg.Auth.CookieSecureMode)
	c.Cookie(&fiber.Cookie{
		Name:     "aeterna_session",
		Value:    token,
		Expires:  exp,
		Path:     "/",
		HTTPOnly: true,
		Secure:   secure,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
	return nil
}

func (h *AuthHandlers) clearSessionCookie(c *fiber.Ctx) {
	secure := middleware.ShouldUseSecureCookie(c, h.cfg.Auth.CookieSecureMode)
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
