package controllers

import (
	"gorm.io/gorm"
)

type OutboundController struct {
	DB *gorm.DB
}

func NewOutboundController(db *gorm.DB) *OutboundController {
	return &OutboundController{DB: db}
}

// Request structs
type CreateOutboundRequest struct {
	TrackingNumber string `json:"tracking_number" validate:"required,min=4,max=100"`
}

// GetOutbounds retrieves a list of outbounds with pagination and search
// @Summary Get Outbounds
// @Description Retrieve a list of outbounds with pagination and search
// @Tags Outbounds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of Outbounds per page" default(10)
// @Param search query string false "Search term for outbound Tracking Number"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.Outbound}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/outbounds [get]
// func (oc *OutboundController) GetOutbounds(c fiber.Ctx) error {
// 	// Parse pagination parameters
// 	page, _ := strconv.Atoi(c.Query("page", "1"))
// 	limit, _ := strconv.Atoi(c.Query("limit", "10"))
// 	offset := (page - 1) * limit

// 	var outbounds []models.Outbound

// 	// Build base query
// 	query := oc.DB.Model(&models.Outbound{}).Order("created_at DESC")
// }
