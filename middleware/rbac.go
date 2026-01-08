package middleware

import (
	"livo-fiber-backend/database"
	"livo-fiber-backend/models"

	"github.com/gofiber/fiber/v3"
)

// RoleMiddleware checks if user has required role hierarchy
func RoleMiddleware(allowedRoles []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRoles := c.Locals("userRoles").([]string)

		// Get minimum hierarchy level from allowed roles
		minHierarchy := 999
		for _, allowedRole := range allowedRoles {
			var role models.Role
			if err := database.DB.Where("name = ?", allowedRole).First(&role).Error; err == nil {
				if role.Hierarchy < minHierarchy {
					minHierarchy = role.Hierarchy
				}
			}
		}

		// Check if user has any role with equal or higher privilege
		for _, userRole := range userRoles {
			var role models.Role
			if err := database.DB.Where("name = ?", userRole).First(&role).Error; err == nil {
				if role.Hierarchy <= minHierarchy {
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}
