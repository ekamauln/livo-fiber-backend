package controllers

import (
	"fmt"
	"livo-fiber-backend/database"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type QCOnlineController struct {
	DB *gorm.DB
}

func NewQCOnlineController(db *gorm.DB) *QCOnlineController {
	return &QCOnlineController{DB: db}
}

// Request structs
type CreateQCOnlineRequest struct {
	TrackingNumber string                  `json:"trackingNumber" validate:"required"`
	Details        []QCOnlineDetailRequest `json:"details" validate:"required,dive,required"`
}

type QCOnlineDetailRequest struct {
	BoxID    uint `json:"boxId" validate:"required"`
	Quantity uint `json:"quantity" validate:"required,min=1"`
}

// GetQCOnlines retrieves a list of qc onlines with pagination and search
// @Summary Get QC Onlines
// @Description Retrieve a list of QC Onlines with pagination and search
// @Tags QC Onlines
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of QC Onlines per page" default(10)
// @Param search query string false "Search term for tracking number"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.QCOnlineResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/qc-onlines [get]
func (qcoc *QCOnlineController) GetQCOnlines(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var qcOnlines []models.QCOnline

	// Get current logged in user from context
	userIDStr := c.Locals("userId").(string)
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// Get start of current day (midnight)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Build base query
	query := qcoc.DB.Model(&models.QCOnline{}).Preload("QCOnlineDetails.Box").Preload("QCUser").Order("created_at DESC").Where("qc_by = ?", uint(userID)).Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)

	// Search condition if provided
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("tracking_number ILIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&qcOnlines).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve QC Onlines",
		})
	}

	// Format response
	qcOnlineList := make([]models.QCOnlineResponse, len(qcOnlines))
	for i, qcOnline := range qcOnlines {
		qcOnlineList[i] = *qcOnline.ToResponse()
	}

	// Build success message
	message := "QC Onlines retrieved successfully"
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
		Data:    qcOnlineList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// CreateQCOnline creates a new QC Online
// @Summary Create QC Online
// @Description Create a new QC Online
// @Tags QC Onlines
// @Accept json
// @Produce json
// @Param qcOnline body CreateQCOnlineRequest true "QC Online details"
// @Success 201 {object} utils.SuccessResponse{data=models.QCOnlineResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/qc-onlines [post]
func (qcoc *QCOnlineController) CreateQCOnline(c fiber.Ctx) error {
	// Binding request body
	var req CreateQCOnlineRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert tracking number to uppercase and trim spaces
	req.TrackingNumber = strings.ToUpper(strings.TrimSpace(req.TrackingNumber))

	// Check for existing QC Online with same tracking number
	var existingQCOnline models.QCOnline
	if err := qcoc.DB.Where("tracking_number = ?", req.TrackingNumber).First(&existingQCOnline).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "QC Online with the same " + req.TrackingNumber + " already exists.",
		})
	}

	// Get current logged in user from context
	userIDStr := c.Locals("userId").(string)
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// Check if tracking number already exists in QC Ribbon
	var existingQCRibbon models.QCRibbon
	if err := qcoc.DB.Where("tracking_number = ?", req.TrackingNumber).First(&existingQCRibbon).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Tracking number " + req.TrackingNumber + " already exists in QC Ribbon records.",
		})
	}

	// Check if tracking number exists in orders and have processing status "picking completed"
	var order models.Order
	if err := qcoc.DB.Where("tracking_number = ? AND processing_status = ?", req.TrackingNumber, "picking completed").First(&order).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "No order found with tracking number " + req.TrackingNumber + " in picking completed status.",
		})
	}

	// Check if order processing status is already "qc completed"
	if order.ProcessingStatus == "qc completed" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with tracking number " + req.TrackingNumber + " is already in qc completed status.",
		})
	}

	// Validate all boxes exist and no duplicates
	boxIDSet := make(map[uint]bool)
	for _, detailReq := range req.Details {
		// Check for duplicate box IDs in the request
		if boxIDSet[detailReq.BoxID] {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Duplicate box ID in the request",
			})
		}
		boxIDSet[detailReq.BoxID] = true

		// Check if box exists
		var box models.Box
		if err := qcoc.DB.Where("id = ?", detailReq.BoxID).First(&box).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Box with ID " + strconv.FormatUint(uint64(detailReq.BoxID), 10) + " does not exist",
			})
		}

		// Validate quantity
		if detailReq.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Quantity must be greater than zero for box ID " + strconv.FormatUint(uint64(detailReq.BoxID), 10),
			})
		}
	}

	// Start database transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create QCOnline record
	qcOnline := models.QCOnline{
		TrackingNumber: req.TrackingNumber,
		QCBy:           uint(userID),
	}

	// Create QCOnlineDetails records
	for _, detailReq := range req.Details {
		qcOnlineDetail := models.QCOnlineDetail{
			BoxID:    detailReq.BoxID,
			Quantity: detailReq.Quantity,
		}
		qcOnline.QCOnlineDetails = append(qcOnline.QCOnlineDetails, qcOnlineDetail)
	}

	// Create records in the database (GORM will cascade to details automatically)
	if err := tx.Create(&qcOnline).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create QC online record",
		})
	}

	// Update order processing status to "qc completed"
	if err := tx.Model(&models.Order{}).Where("tracking_number = ?", req.TrackingNumber).Update("processing_status", "qc completed").Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update order processing status",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to commit transaction",
		})
	}

	// Reload the created record with all relationships for response
	if err := qcoc.DB.Preload("QCOnlineDetails.Box").Preload("QCUser").First(&qcOnline, qcOnline.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to load created QC online",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "QC online created successfully",
		Data:    qcOnline.ToResponse(),
	})
}
