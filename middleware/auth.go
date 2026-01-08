package middleware

import (
	"livo-fiber-backend/config"
	"livo-fiber-backend/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token, err := utils.ValidateToken(parts[1], cfg)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Check token type
		tokenType, err := token.GetString("type")
		if err != nil || tokenType != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token type",
			})
		}

		// Extract claims
		userID, err := token.GetString("userId")
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		username, _ := token.GetString("username")

		var roles []string
		if err := token.Get("roles", &roles); err != nil {
			roles = []string{}
		}

		// Store in context
		c.Locals("userId", userID)
		c.Locals("username", username)
		c.Locals("userRoles", roles)

		return c.Next()
	}
}
