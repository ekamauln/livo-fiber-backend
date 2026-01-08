package utils

import "github.com/gofiber/fiber/v3"

// HasPermission checks if user has any of the allowed roles
func HasPermission(c fiber.Ctx, allowedRoles []string) bool {
	userRoles := c.Locals("userRoles").([]string)
	for _, userRole := range userRoles {
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return true
			}
		}
	}
	return false
}
