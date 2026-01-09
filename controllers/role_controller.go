package controllers

import (
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type RoleController struct {
	DB *gorm.DB
}

func NewRoleController(db *gorm.DB) *RoleController {
	return &RoleController{DB: db}
}

// Request structs
type CreateRoleRequest struct {
	RoleName  string `json:"roleName" validate:"required,min=3,max=50"`
	Hierarchy int    `json:"hierarchy" validate:"required,min=1"`
}

type UpdateRoleRequest struct {
	RoleName  string `json:"roleName" validate:"required,min=3,max=50"`
	Hierarchy int    `json:"hierarchy" validate:"required,min=1"`
}

// GetRoles retrieves a list of roles with pagination and search
// @Summary Get Roles
// @Description Retrieve a list of roles with pagination and search
// @Tags Roles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of roles per page" default(10)
// @Param search query string false "Search term for role name"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.Role}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/roles [get]
func (rc *RoleController) GetRoles(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var roles []models.Role

	// Build base query
	query := rc.DB.Model(&models.Role{}).Order("created_at DESC")

	// Search condition if provided
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("role_name ILIKE ?", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve roles",
		})
	}

	// Format response
	roleList := make([]models.RoleResponse, len(roles))
	for i, role := range roles {
		roleList[i] = *role.ToResponse()
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessPaginatedResponse{
		Success: true,
		Message: "Roles retrieved successfully",
		Data:    roleList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetUser retrieves a single role by ID
// @Summary Get Role
// @Description Retrieve a single role by ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Role}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/roles/{id} [get]
func (rc *RoleController) GetRole(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var role models.Role
	if err := rc.DB.Where("id = ?", id).First(&role).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Role with id " + id + " not found.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Role retrieved successfully",
		Data:    role.ToResponse(),
	})
}

// CreateRole creates a new role
// @Summary Create Role
// @Description Create a new role
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body CreateRoleRequest true "Role details"
// @Success 201 {object} utils.SuccessResponse{data=models.Role}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/roles [post]
func (rc *RoleController) CreateRole(c fiber.Ctx) error {
	// Binding request body
	var req CreateRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Check for existing role with same name
	var existingRole models.Role
	if err := rc.DB.Where("role_name = ?", req.RoleName).First(&existingRole).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Role with name " + req.RoleName + " already exists.",
		})
	}

	// Check permission hierarchy - current user only can create roles with equal or lower hierarchy
	currUserRoles := c.Locals("userRoles").([]string)
	currUserMinHierarchy := 999
	for _, currUserRoleName := range currUserRoles {
		var currRole models.Role
		if err := rc.DB.Where("role_name = ?", currUserRoleName).First(&currRole).Error; err == nil {
			if currRole.Hierarchy < currUserMinHierarchy {
				currUserMinHierarchy = currRole.Hierarchy
			}
		}
	}

	// Validate new role hierarchy
	if req.Hierarchy < currUserMinHierarchy {
		return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions to create role with higher privilege",
		})
	}

	// Create new role
	newRole := models.Role{
		RoleName:  req.RoleName,
		Hierarchy: req.Hierarchy,
	}

	if err := rc.DB.Create(&newRole).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create role",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Role created successfully",
		Data:    newRole.ToResponse(),
	})
}

// UpdateRole updates an existing role by ID
// @Summary Update Role
// @Description Update an existing role by ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param request body UpdateRoleRequest true "Updated role details"
// @Success 200 {object} utils.SuccessResponse{data=models.Role}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/roles/{id} [put]
func (rc *RoleController) UpdateRole(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var role models.Role
	if err := rc.DB.Where("id = ?", id).First(&role).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Role with id " + id + " not found.",
		})
	}

	// Binding request body
	var req UpdateRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Check for existing role with same name (excluding current role)
	var existingRole models.Role
	if err := rc.DB.Where("role_name = ? AND id != ?", req.RoleName, id).First(&existingRole).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Role with name " + req.RoleName + " already exists.",
		})
	}

	// Check permission hierarchy - current user only can create roles with equal or lower hierarchy
	currUserRoles := c.Locals("userRoles").([]string)
	currUserMinHierarchy := 999
	for _, currUserRoleName := range currUserRoles {
		var currRole models.Role
		if err := rc.DB.Where("role_name = ?", currUserRoleName).First(&currRole).Error; err == nil {
			if currRole.Hierarchy < currUserMinHierarchy {
				currUserMinHierarchy = currRole.Hierarchy
			}
		}
	}

	// Validate new role hierarchy
	if req.Hierarchy < currUserMinHierarchy {
		return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions to create role with higher privilege",
		})
	}

	// Update role fields
	role.RoleName = req.RoleName
	role.Hierarchy = req.Hierarchy

	if err := rc.DB.Save(&role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update role",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Role updated successfully",
		Data:    role.ToResponse(),
	})
}

// DeleteRole deletes a role by ID
// @Summary Delete Role
// @Description Delete a role by ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/roles/{id} [delete]
func (rc *RoleController) DeleteRole(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var role models.Role
	if err := rc.DB.Where("id = ?", id).First(&role).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Role with id " + id + " not found.",
		})
	}

	// Delete role (also deletes user_roles due to foreign key constraint with ON DELETE CASCADE)
	if err := rc.DB.Delete(&role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to delete role",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Role deleted successfully",
	})
}
