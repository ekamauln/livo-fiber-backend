package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func CSRFMiddleware() fiber.Handler {
	// Create a session store for CSRF tokens
	sessionStore := session.NewStore()

	return csrf.New(csrf.Config{
		CookieName:        "__Host-csrf_",
		CookieSameSite:    "Lax",
		CookieSecure:      true,
		CookieHTTPOnly:    true,
		CookieSessionOnly: true,
		Extractor:         extractors.FromHeader("X-Csrf-Token"),
		Session:           sessionStore,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid CSRF token",
			})
		},
	})
}
