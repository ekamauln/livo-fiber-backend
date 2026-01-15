package controllers

import (
	"fmt"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type MobileChannelController struct {
	DB *gorm.DB
}

func NewMobileChannelController(db *gorm.DB) *MobileChannelController {
	return &MobileChannelController{DB: db}
}

// GetMobileChannels retrieves all mobile channels from the database
// @Summary Get Mobile Channels
// @Description Retrieve all mobile channels from the database
// @Tags Mobile Returns
// @Accept json
// @Produce json
// @Param search query string false "Search term to filter channels by name or code"
// @Success 200 {object} utils.SuccessTotaledResponse{data=[]models.ChannelResponse}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/mobile-returns/channels [get]
func (mcc *MobileChannelController) GetMobileChannels(c fiber.Ctx) error {
	var mobileChannels []models.Channel

	// Build base query
	query := mcc.DB.Model(&models.Channel{})

	// Parse search query parameter
	search := c.Query("search")
	if search != "" {
		likeSearch := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", likeSearch, likeSearch)
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute the query to fetch mobile channels
	if err := query.Find(&mobileChannels).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve mobile channels")
	}

	// Response format
	mobileChannelList := make([]models.ChannelResponse, len(mobileChannels))
	for i, channel := range mobileChannels {
		mobileChannelList[i] = *channel.ToResponse()
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
		Data:    mobileChannelList,
		Total:   total,
	})
}
