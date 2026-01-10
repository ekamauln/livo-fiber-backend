package controllers

import (
	"fmt"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

// Request structs
type UpdateUserRequest struct {
	FullName string `json:"fullName" validate:"omitempty,min=3,max=100" example:"John Doe"`
	Email    string `json:"email" validate:"omitempty,email" example:"john@example.com"`
	IsActive *bool  `json:"isActive" validate:"omitempty" example:"true"`
}

type UpdatePasswordRequest struct {
	NewPassword        string `json:"newPassword" validate:"required,min=8" example:"SecurePass123"`
	ConfirmNewPassword string `json:"confirmNewPassword" validate:"required,eqfield=NewPassword" example:"SecurePass123"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50" example:"john_doe"`
	Password string `json:"password" validate:"required,min=8" example:"SecurePass123"`
	FullName string `json:"fullName" validate:"required,min=3,max=100" example:"John Doe"`
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	RoleName string `json:"roleName,omitempty" example:"guest"` // Optional role assignment
}

type AssignRoleRequest struct {
	RoleName string `json:"roleName" validate:"required" example:"guest"`
}

type RemoveRoleRequest struct {
	RoleName string `json:"roleName" validate:"required" example:"guest"`
}

// GetUsers retrieves a paginated list of users with optional search and role filtering
// @Summary Get Users
// @Description Retrieve a paginated list of users with optional search and role filtering
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of users per page" default(10)
// @Param search query string false "Search term for username or full name"
// @Param role query string false "Filter users by role name"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.UserResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users [get]
func (uc *UserController) GetUsers(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var users []models.User

	// Build base query
	query := uc.DB.Model(&models.User{}).Order("created_at DESC").Preload("Roles")

	// Filter by role if provided
	roleName := strings.TrimSpace(c.Query("roleName", ""))
	if roleName != "" {
		query = query.Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Joins("JOIN roles ON user_roles.role_id = roles.id").
			Where("roles.role_name = ?", roleName)
	}

	// Search condition if provided
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("username ILIKE ? OR full_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated users
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve users",
		})
	}

	// Format response
	userList := make([]models.UserResponse, len(users))
	for i, user := range users {
		userList[i] = *user.ToResponse()
	}

	// Build success message
	message := "Users retrieved successfully"
	var filters []string

	if search != "" {
		filters = append(filters, "search: "+search)
	}

	if len(filters) > 0 {
		message += fmt.Sprintf(" (filtered by %s)", strings.Join(filters, " | "))
	}

	// Return success response
	return c.JSON(utils.SuccessPaginatedResponse{
		Success: true,
		Message: message,
		Data:    userList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetUser retrieves a single user by ID
// @Summary Get User
// @Description Retrieve a single user by ID
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} utils.SuccessResponse{data=models.UserResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users/{id} [get]
func (uc *UserController) GetUser(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var user models.User
	if err := uc.DB.Preload("Roles").Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User with id " + id + " not found.",
		})
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user.ToResponse(),
	})
}

// CreateUser creates a new user with optional role assignment
// @Summary Create User
// @Description Create a new user with optional role assignment
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "User details"
// @Success 201 {object} utils.SuccessResponse{data=models.UserResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users [post]
func (uc *UserController) CreateUser(c fiber.Ctx) error {
	// Binding request body
	var req CreateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Check for existing username or email
	var existingUser models.User
	if err := uc.DB.Preload("Roles").Where("username = ?", req.Username).Or("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Username or email already exists",
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
	}

	// Start database transaction
	tx := uc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	newUser := models.User{
		Username: req.Username,
		Password: hashedPassword,
		FullName: req.FullName,
		Email:    req.Email,
		IsActive: true,
	}

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create user",
		})
	}

	// Assign role if provided
	if req.RoleName != "" {
		var role models.Role
		if err := tx.Where("role_name = ?", req.RoleName).First(&role).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Invalid role name",
			})
		}

		// Check permission hierarchy - current user must have higher or equal privilege
		currUserRoles := c.Locals("userRoles").([]string)
		currUserMinHierarchy := 999
		for _, currUserRoleName := range currUserRoles {
			var currRole models.Role
			if err := tx.Where("role_name = ?", currUserRoleName).First(&currRole).Error; err == nil {
				if currRole.Hierarchy < currUserMinHierarchy {
					currUserMinHierarchy = currRole.Hierarchy
				}
			}
		}

		// Current user must have equal or higher privilege (lower or equal hierarchy number)
		if role.Hierarchy < currUserMinHierarchy {
			tx.Rollback()
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Insufficient permissions",
			})
		}

		// Assign role to user
		userRole := models.UserRole{
			UserID: newUser.ID,
			RoleID: role.ID,
		}

		if err := tx.Create(&userRole).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Failed to assign role to user",
			})
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create user",
		})
	}

	// Reload the data
	if err := uc.DB.Preload("Roles").Where("id = ?", newUser.ID).First(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to load user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "User created successfully",
		Data:    newUser.ToResponse(),
	})
}

// UpdateUser updates user details
// @Summary Update User
// @Description Update user details
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdateUserRequest true "Updated user details"
// @Success 200 {object} utils.SuccessResponse{data=models.UserResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users/{id} [put]
func (uc *UserController) UpdateUser(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var user models.User
	if err := uc.DB.Preload("Roles").Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User with id " + id + " not found.",
		})
	}

	// Binding request body
	var req UpdateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Users can only update their own profile unless they have developer/superadmin/hrd role
	currUserID := c.Locals("userId").(string)
	if id != currUserID {
		if !utils.HasPermission(c, []string{"developer", "superadmin", "hrd"}) {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Insufficient permissions to update other user's profile",
			})
		}
	}

	// Update fields if provided
	user.FullName = req.FullName
	// If updating email, check for uniqueness
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := uc.DB.Where("email = ? AND id != ?", req.Email, id).First(&existingUser).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Email already in use",
			})
		}
		user.Email = req.Email
	}
	// Only developer/superadmin/hrd can update IsActive
	if req.IsActive != nil {
		if !utils.HasPermission(c, []string{"developer", "superadmin", "hrd"}) {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Insufficient permissions to update user status",
			})
		}
		user.IsActive = *req.IsActive
	}

	if err := uc.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update user",
		})
	}

	// Reload the data with fresh query
	var reloadedUser models.User
	if err := uc.DB.Preload("Roles").Where("id = ?", user.ID).First(&reloadedUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to load user",
		})
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    reloadedUser.ToResponse(),
	})
}

// UpdatePassword updates a user's password
// @Summary Update Password
// @Description Update a user's password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdatePasswordRequest true "Updated password details"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users/{id}/password [put]
func (uc *UserController) UpdatePassword(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var user models.User
	if err := uc.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User with id " + id + " not found.",
		})
	}

	// Binding request body
	var req UpdatePasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Users can only update their own password unless they have developer/superadmin/hrd role
	currUserID := c.Locals("userId").(string)
	if id != currUserID {
		if !utils.HasPermission(c, []string{"developer", "superadmin", "hrd"}) {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Insufficient permissions to update other user's password",
			})
		}
	}

	// Check if new password and confirm password match
	if req.NewPassword != req.ConfirmNewPassword {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "New password and confirm password do not match",
		})
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
	}

	user.Password = hashedPassword
	if err := uc.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update password",
		})
	}

	// Clear sessions or tokens if user updated their own password
	uc.DB.Where("user_id = ?", user.ID).Delete(&models.Session{})

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}

// DeleteUser deletes a user by ID and all associated sessions
// @Summary Delete User
// @Description Delete a user by ID and all associated sessions
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users/{id} [delete]
func (uc *UserController) DeleteUser(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var user models.User
	if err := uc.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User with id " + id + " not found.",
		})
	}

	// Delete all sessions associated with the user
	uc.DB.Where("user_id = ?", user.ID).Delete(&models.Session{})

	// Delete user (also deletes user_roles due to foreign key constraint with ON DELETE CASCADE)
	if err := uc.DB.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to delete user",
		})
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// AssignRole assign a role to a user
// @Summary Assign Role
// @Description Assign a role to a user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body AssignRoleRequest true "Role to assign"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users/{id}/roles [post]
func (uc *UserController) AssignRole(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var user models.User
	if err := uc.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User with id " + id + " not found.",
		})
	}

	// Binding request body
	var req AssignRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Get the role to assign
	var role models.Role
	if err := uc.DB.Where("role_name = ?", req.RoleName).First(&role).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid role name",
		})
	}

	// Check if user already has the role
	var userRole models.UserRole
	if err := uc.DB.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User already has this role",
		})
	}

	// Check permission hierarchy - current user must have higher or equal privilege
	currUserRoles := c.Locals("userRoles").([]string)
	currUserMinHierarchy := 999
	for _, currUserRoleName := range currUserRoles {
		var currRole models.Role
		if err := uc.DB.Where("role_name = ?", currUserRoleName).First(&currRole).Error; err == nil {
			if currRole.Hierarchy < currUserMinHierarchy {
				currUserMinHierarchy = currRole.Hierarchy
			}
		}
	}

	// Current user must have equal or higher privilege (lower or equal hierarchy number)
	if role.Hierarchy < currUserMinHierarchy {
		return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions",
		})
	}

	// Assign role to user
	newUserRole := models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}

	if err := uc.DB.Create(&newUserRole).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to assign role to user",
		})
	}

	// Reload the data with fresh query
	var reloadedUser models.User
	if err := uc.DB.Preload("Roles").Where("id = ?", user.ID).First(&reloadedUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to load user",
		})
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "Role assigned to user successfully",
		Data:    reloadedUser.ToResponse(),
	})
}

// RemoveRole removes a role from a user
// @Summary Remove Role
// @Description Remove a role from a user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body RemoveRoleRequest true "Role to remove"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users/{id}/roles [delete]
func (uc *UserController) RemoveRole(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var user models.User
	if err := uc.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User with id " + id + " not found.",
		})
	}

	// Binding request body
	var req RemoveRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Get the role to remove
	var role models.Role
	if err := uc.DB.Where("role_name = ?", req.RoleName).First(&role).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid role name",
		})
	}

	// Check if user has the role
	var userRole models.UserRole
	if err := uc.DB.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User does not have this role",
		})
	}

	// Check permission hierarchy - current user must have higher or equal privilege
	currUserRoles := c.Locals("userRoles").([]string)
	currUserMinHierarchy := 999
	for _, currUserRoleName := range currUserRoles {
		var currRole models.Role
		if err := uc.DB.Where("role_name = ?", currUserRoleName).First(&currRole).Error; err == nil {
			if currRole.Hierarchy < currUserMinHierarchy {
				currUserMinHierarchy = currRole.Hierarchy
			}
		}
	}

	// Current user must have equal or higher privilege (lower or equal hierarchy number)
	if role.Hierarchy < currUserMinHierarchy {
		return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions",
		})
	}

	// Remove role from user
	if err := uc.DB.Delete(&models.UserRole{}, "user_id = ? AND role_id = ?", user.ID, role.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to remove role from user",
		})
	}

	// Reload the data with fresh query
	var reloadedUser models.User
	if err := uc.DB.Preload("Roles").Where("id = ?", user.ID).First(&reloadedUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to load user",
		})
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "Role removed from user successfully",
		Data:    reloadedUser.ToResponse(),
	})
}

// GetSessions retrieves all active sessions for a user
// @Summary Get User Sessions
// @Description Retrieve all active sessions for a user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} utils.SuccessResponse{data=[]models.SessionResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/users/{id}/sessions [get]
func (uc *UserController) GetSessions(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var user models.User
	if err := uc.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User with id " + id + " not found.",
		})
	}

	// Users can only view their own sessions unless they have developer/superadmin/hrd role
	currUserID := c.Locals("userId").(string)
	if id != currUserID {
		if !utils.HasPermission(c, []string{"developer", "superadmin", "hrd"}) {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Insufficient permissions to view other user's sessions",
			})
		}
	}

	var sessions []models.Session
	if err := uc.DB.Where("user_id = ?", user.ID).Find(&sessions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve user sessions",
		})
	}

	// Format response
	sessionList := make([]models.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionList[i] = *session.ToResponse()
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "User sessions retrieved successfully",
		Data:    sessionList,
	})
}
