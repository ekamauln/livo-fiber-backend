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

type StoreController struct {
	DB *gorm.DB
}

func NewStoreController(db *gorm.DB) *StoreController {
	return &StoreController{DB: db}
}

// Request structs
type CreateStoreRequest struct {
	StoreCode string `json:"storeCode" validate:"required,min=3,max=50"`
	StoreName string `json:"storeName" validate:"required,min=3,max=100"`
}

type UpdateStoreRequest struct {
	StoreCode string `json:"storeCode" validate:"required,min=3,max=50"`
	StoreName string `json:"storeName" validate:"required,min=3,max=100"`
}

// GetStores retrieves a list of stores with pagination and search
// @Summary Get Stores
// @Description Retrieve a list of stores with pagination and search
// @Tags Stores
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of stores per page" default(10)
// @Param search query string false "Search term for store code or name"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.Store}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/stores [get]
func (bc *StoreController) GetStores(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var stores []models.Store

	// Build base query
	query := bc.DB.Model(&models.Store{}).Order("created_at DESC")

	// Search condition if provided
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("store_code ILIKE ? OR store_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Limit(limit).Offset(offset).Find(&stores).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve stores",
		})
	}

	// Format response
	storeList := make([]models.StoreResponse, len(stores))
	for i, store := range stores {
		storeList[i] = *store.ToResponse()
	}

	// Build success message
	message := "Stores retrieved successfully"
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
		Data:    storeList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetStore retrieves a single store by ID
// @Summary Get Store
// @Description Retrieve a single store by ID
// @Tags Stores
// @Accept json
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Store}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/stores/{id} [get]
func (bc *StoreController) GetStore(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var store models.Store
	if err := bc.DB.Where("id = ?", id).First(&store).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Store with id " + id + " not found.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Store retrieved successfully",
		Data:    store.ToResponse(),
	})
}

// CreateStore creates a new store
// @Summary Create Store
// @Description Create a new store
// @Tags Stores
// @Accept json
// @Produce json
// @Param store body CreateStoreRequest true "Store details"
// @Success 201 {object} utils.SuccessResponse{data=models.Store}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/stores [post]
func (bc *StoreController) CreateStore(c fiber.Ctx) error {
	// Binding request body
	var req CreateStoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert store code to uppercase and trim spaces
	req.StoreCode = strings.ToUpper(strings.TrimSpace(req.StoreCode))

	// Check for existing store with same code
	var existingStore models.Store
	if err := bc.DB.Where("store_code = ?", req.StoreCode).First(&existingStore).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Store with code " + req.StoreCode + " already exists.",
		})
	}

	// Create new store
	newStore := models.Store{
		StoreCode: req.StoreCode,
		StoreName: req.StoreName,
	}

	if err := bc.DB.Create(&newStore).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create store",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Store created successfully",
		Data:    newStore.ToResponse(),
	})
}

// UpdateStore updates an existing store by ID
// @Summary Update Store
// @Description Update an existing store by ID
// @Tags Stores
// @Accept json
// @Produce json
// @Param id path int true "Store ID"
// @Param request body UpdateStoreRequest true "Updated store details"
// @Success 200 {object} utils.SuccessResponse{data=models.Store}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/stores/{id} [put]
func (bc *StoreController) UpdateStore(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var store models.Store
	if err := bc.DB.Where("id = ?", id).First(&store).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Store with id " + id + " not found.",
		})
	}

	// Binding request body
	var req UpdateStoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert store code to uppercase and trim spaces
	req.StoreCode = strings.ToUpper(strings.TrimSpace(req.StoreCode))

	// Check for existing store with same code (excluding current store)
	var existingStore models.Store
	if err := bc.DB.Where("store_code = ? AND id != ?", req.StoreCode, id).First(&existingStore).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Store with code " + req.StoreCode + " already exists.",
		})
	}

	// Update store fields
	store.StoreCode = req.StoreCode
	store.StoreName = req.StoreName

	if err := bc.DB.Save(&store).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update store",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Store updated successfully",
		Data:    store.ToResponse(),
	})
}

// DeleteStore deletes a store by ID
// @Summary Delete Store
// @Description Delete a store by ID
// @Tags Stores
// @Accept json
// @Produce json
// @Param id path int true "Store ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/stores/{id} [delete]
func (bc *StoreController) DeleteStore(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var store models.Store
	if err := bc.DB.Where("id = ?", id).First(&store).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Store with id " + id + " not found.",
		})
	}

	// Delete store (also deletes associated records if any due to foreign key constraints)
	if err := bc.DB.Delete(&store).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to delete store",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Store deleted successfully",
	})
}
