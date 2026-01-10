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

type ExpeditionController struct {
	DB *gorm.DB
}

func NewExpeditionController(db *gorm.DB) *ExpeditionController {
	return &ExpeditionController{DB: db}
}

// Request structs
type CreateExpeditionRequest struct {
	ExpeditionCode  string `json:"expeditionCode" validate:"required,min=3,max=50"`
	ExpeditionName  string `json:"expeditionName" validate:"required,min=3,max=100"`
	ExpeditionColor string `json:"expeditionColor" validate:"required,min=3,max=20"`
}

type UpdateExpeditionRequest struct {
	ExpeditionCode  string `json:"expeditionCode" validate:"required,min=3,max=50"`
	ExpeditionName  string `json:"expeditionName" validate:"required,min=3,max=100"`
	ExpeditionColor string `json:"expeditionColor" validate:"required,min=3,max=20"`
}

// GetExpeditions retrieves a list of expeditions with pagination and search
// @Summary Get Expeditions
// @Description Retrieve a list of expeditions with pagination and search
// @Tags Expeditions
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of expeditions per page" default(10)
// @Param search query string false "Search term for expedition code or name"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.Expedition}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/expeditions [get]
func (bc *ExpeditionController) GetExpeditions(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var expeditions []models.Expedition

	// Build base query
	query := bc.DB.Model(&models.Expedition{}).Order("created_at DESC")

	// Search condition if provided
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("expedition_code ILIKE ? OR expedition_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Limit(limit).Offset(offset).Find(&expeditions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve expeditions",
		})
	}

	// Format response
	expeditionList := make([]models.ExpeditionResponse, len(expeditions))
	for i, expedition := range expeditions {
		expeditionList[i] = *expedition.ToResponse()
	}

	// Build success message
	message := "Expeditions retrieved successfully"
	var filters []string

	if search != "" {
		filters = append(filters, "search: "+search)
	}

	if len(filters) > 0 {
		message += fmt.Sprintf(" (filtered by %s)", strings.Join(filters, " | "))
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(utils.SuccessPaginatedResponse{
		Success: true,
		Message: message,
		Data:    expeditionList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetExpedition retrieves a single expedition by ID
// @Summary Get Expedition
// @Description Retrieve a single expedition by ID
// @Tags Expeditions
// @Accept json
// @Produce json
// @Param id path int true "Expedition ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Expedition}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/expeditions/{id} [get]
func (bc *ExpeditionController) GetExpedition(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var expedition models.Expedition
	if err := bc.DB.Where("id = ?", id).First(&expedition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Expedition with id " + id + " not found.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Expedition retrieved successfully",
		Data:    expedition.ToResponse(),
	})
}

// CreateExpedition creates a new expedition
// @Summary Create Expedition
// @Description Create a new expedition
// @Tags Expeditions
// @Accept json
// @Produce json
// @Param expedition body CreateExpeditionRequest true "Expedition details"
// @Success 201 {object} utils.SuccessResponse{data=models.Expedition}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/expeditions [post]
func (bc *ExpeditionController) CreateExpedition(c fiber.Ctx) error {
	// Binding request body
	var req CreateExpeditionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert expedition code to uppercase and trim spaces
	req.ExpeditionCode = strings.ToUpper(strings.TrimSpace(req.ExpeditionCode))

	// Check for existing expedition with same code
	var existingExpedition models.Expedition
	if err := bc.DB.Where("expedition_code = ?", req.ExpeditionCode).First(&existingExpedition).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Expedition with code " + req.ExpeditionCode + " already exists.",
		})
	}

	// Create new expedition
	newExpedition := models.Expedition{
		ExpeditionCode:  req.ExpeditionCode,
		ExpeditionName:  req.ExpeditionName,
		ExpeditionSlug:  utils.GenerateSlug(req.ExpeditionName),
		ExpeditionColor: req.ExpeditionColor,
	}

	if err := bc.DB.Create(&newExpedition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create expedition",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Expedition created successfully",
		Data:    newExpedition.ToResponse(),
	})
}

// UpdateExpedition updates an existing expedition by ID
// @Summary Update Expedition
// @Description Update an existing expedition by ID
// @Tags Expeditions
// @Accept json
// @Produce json
// @Param id path int true "Expedition ID"
// @Param request body UpdateExpeditionRequest true "Updated expedition details"
// @Success 200 {object} utils.SuccessResponse{data=models.Expedition}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/expeditions/{id} [put]
func (bc *ExpeditionController) UpdateExpedition(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var expedition models.Expedition
	if err := bc.DB.Where("id = ?", id).First(&expedition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Expedition with id " + id + " not found.",
		})
	}

	// Binding request body
	var req UpdateExpeditionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert expedition code to uppercase and trim spaces
	req.ExpeditionCode = strings.ToUpper(strings.TrimSpace(req.ExpeditionCode))

	// Check for existing expedition with same code (excluding current expedition)
	var existingExpedition models.Expedition
	if err := bc.DB.Where("expedition_code = ? AND id != ?", req.ExpeditionCode, id).First(&existingExpedition).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Expedition with code " + req.ExpeditionCode + " already exists.",
		})
	}

	// Update expedition fields
	expedition.ExpeditionCode = req.ExpeditionCode
	expedition.ExpeditionName = req.ExpeditionName
	expedition.ExpeditionSlug = utils.GenerateSlug(req.ExpeditionName)
	expedition.ExpeditionColor = req.ExpeditionColor

	if err := bc.DB.Save(&expedition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update expedition",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Expedition updated successfully",
		Data:    expedition.ToResponse(),
	})
}

// DeleteExpedition deletes an expedition by ID
// @Summary Delete Expedition
// @Description Delete an expedition by ID
// @Tags Expeditions
// @Accept json
// @Produce json
// @Param id path int true "Expedition ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/expeditions/{id} [delete]
func (bc *ExpeditionController) DeleteExpedition(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var expedition models.Expedition
	if err := bc.DB.Where("id = ?", id).First(&expedition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Expedition with id " + id + " not found.",
		})
	}

	// Delete expedition (also deletes associated records if any due to foreign key constraints)
	if err := bc.DB.Delete(&expedition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to delete expedition",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Expedition deleted successfully",
	})
}
