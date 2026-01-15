package controllers

import (
	"fmt"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type MobileStoreController struct {
	DB *gorm.DB
}

func NewMobileStoreController(db *gorm.DB) *MobileStoreController {
	return &MobileStoreController{DB: db}
}

// GetMobileStores retrieves all mobile stores from the database
// @Summary Get Mobile Stores
// @Description Retrieve all mobile stores from the database
// @Tags Mobile Returns
// @Accept json
// @Produce json
// @Param search query string false "Search term to filter stores by name or code"
// @Success 200 {object} utils.SuccessTotaledResponse{data=[]models.StoreResponse}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/mobile-returns/stores [get]
func (mcs *MobileStoreController) GetMobileStores(c fiber.Ctx) error {
	var mobileStores []models.Store

	// Build base query
	query := mcs.DB.Model(&models.Store{})

	// Parse search query parameter
	search := c.Query("search")
	if search != "" {
		likeSearch := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", likeSearch, likeSearch)
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute the query to fetch mobile stores
	if err := query.Find(&mobileStores).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve mobile stores")
	}

	// Response format
	mobileStoreList := make([]models.StoreResponse, len(mobileStores))
	for i, store := range mobileStores {
		mobileStoreList[i] = *store.ToResponse()
	}

	// Build success message
	message := "Complains retrieved successfully"
	var filters []string

	if search != "" {
		filters = append(filters, "search: "+search)
	}

	if len(filters) > 0 {
		message += fmt.Sprintf(" (filtered by %s)", strings.Join(filters, " | "))
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(utils.SuccessTotaledResponse{
		Success: true,
		Message: message,
		Data:    mobileStoreList,
		Total:   total,
	})
}
