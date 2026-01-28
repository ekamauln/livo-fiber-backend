package controllers

import (
	"fmt"
	"livo-fiber-backend/database"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type QCRibbonController struct {
	DB *gorm.DB
}

func NewQCRibbonController(db *gorm.DB) *QCRibbonController {
	return &QCRibbonController{DB: db}
}

// Request structs
type CreateQCRibbonRequest struct {
	TrackingNumber string                  `json:"trackingNumber" validate:"required"`
	Details        []QCRibbonDetailRequest `json:"details" validate:"required,dive,required"`
}

type QCRibbonDetailRequest struct {
	BoxID    uint `json:"boxId" validate:"required"`
	Quantity int  `json:"quantity" validate:"required,min=1"`
}

// Unique response structs
// QcRibbonDailyCount represents the count of qc-ribbons for a specific date
type QcRibbonDailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// QcRibbonsDailyCountResponse represents the response for daily qc-ribbon counts
type QcRibbonsDailyCountResponse struct {
	Month       string               `json:"month"`
	Year        int                  `json:"year"`
	DailyCounts []QcRibbonDailyCount `json:"dailyCounts"`
	TotalCount  int                  `json:"totalCount"`
}

// GetQCRibbons retrieves a list of qc ribbons with pagination and search
// @Summary Get QC Ribbons
// @Description Retrieve a list of QC Ribbons with pagination and search
// @Tags Ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of QC Ribbons per page" default(10)
// @Param search query string false "Search term for tracking number"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.QCRibbonResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/ribbons/qc-ribbons [get]
func (qcrc *QCRibbonController) GetQCRibbons(c fiber.Ctx) error {
	log.Println("GetQCRibbons called")
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var qcRibbons []models.QCRibbon

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
	query := qcrc.DB.Model(&models.QCRibbon{}).Preload("QCRibbonDetails.Box").Preload("QCUser").Order("created_at DESC").Where("qc_by = ?", uint(userID)).Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)

	// Search condition if provided
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("tracking_number ILIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&qcRibbons).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve QC Ribbons",
		})
	}

	// Load orders for each QC Ribbon by tracking number
	for i := range qcRibbons {
		var order models.Order
		if err := qcrc.DB.Preload("OrderDetails").Where("tracking_number = ?", qcRibbons[i].TrackingNumber).First(&order).Error; err == nil {
			qcRibbons[i].Order = &order
		}
	}

	// Format response
	qcRibbonList := make([]models.QCRibbonResponse, len(qcRibbons))
	for i, qcRibbon := range qcRibbons {
		qcRibbonList[i] = *qcRibbon.ToResponse()
	}

	// Build success message
	message := "QC Ribbons retrieved successfully"
	var filters []string

	if search != "" {
		filters = append(filters, "search: "+search)
	}

	if len(filters) > 0 {
		message += fmt.Sprintf(" (filtered by %s)", strings.Join(filters, " | "))
	}

	// Return success response
	log.Println("GetQCRibbons completed successfully")
	return c.Status(fiber.StatusOK).JSON(utils.SuccessPaginatedResponse{
		Success: true,
		Message: message,
		Data:    qcRibbonList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetQCRibbon retrieves a single qc ribbon by ID
// @Summary Get QC Ribbon
// @Description Retrieve a single QC Ribbon by ID
// @Tags Ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "QC Ribbon ID"
// @Success 200 {object} utils.SuccessResponse{data=models.QCRibbonResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/ribbons/qc-ribbons/{id} [get]
func (qcrc *QCRibbonController) GetQCRibbon(c fiber.Ctx) error {
	log.Println("GetQCRibbon called")
	// Parse id parameter
	id := c.Params("id")
	var qcRibbon models.QCRibbon
	if err := qcrc.DB.Preload("QCRibbonDetails.Box").Preload("QCUser").Where("id = ?", id).First(&qcRibbon).Error; err != nil {
		log.Println("GetQCRibbon - QC Ribbon not found:", err)
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "QC Ribbon with id " + id + " not found.",
		})
	}

	// Load order by tracking number
	var order models.Order
	if err := qcrc.DB.Preload("OrderDetails").Where("tracking_number = ?", qcRibbon.TrackingNumber).First(&order).Error; err == nil {
		qcRibbon.Order = &order
	}

	log.Println("GetQCRibbon completed successfully")
	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "QC Ribbon retrieved successfully",
		Data:    qcRibbon.ToResponse(),
	})
}

// CreateQCRibbon creates a new QC Ribbon
// @Summary Create QC Ribbon
// @Description Create a new QC Ribbon
// @Tags Ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param qcRibbon body CreateQCRibbonRequest true "QC Ribbon details"
// @Success 201 {object} utils.SuccessResponse{data=models.QCRibbonResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/ribbons/qc-ribbons [post]
func (qcrc *QCRibbonController) CreateQCRibbon(c fiber.Ctx) error {
	log.Println("CreateQCRibbon called")
	// Binding request body
	var req CreateQCRibbonRequest
	if err := c.Bind().JSON(&req); err != nil {
		log.Println("CreateQCRibbon - Invalid request body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert tracking number to uppercase and trim spaces
	req.TrackingNumber = strings.ToUpper(strings.TrimSpace(req.TrackingNumber))

	// Check for existing QC Ribbon with same tracking number
	var existingQCRibbon models.QCRibbon
	if err := qcrc.DB.Where("tracking_number = ?", req.TrackingNumber).First(&existingQCRibbon).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "QC Ribbon with the same " + req.TrackingNumber + " already exists.",
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

	// Check if tracking number already exists in QC Online
	var existingQCOnline models.QCOnline
	if err := qcrc.DB.Where("tracking_number = ?", req.TrackingNumber).First(&existingQCOnline).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Tracking number " + req.TrackingNumber + " already exists in QC Online records.",
		})
	}

	// Check if tracking number exists in orders and have processing status "picking_completed"
	var order models.Order
	if err := qcrc.DB.Where("tracking_number = ? AND processing_status = ?", req.TrackingNumber, "picking_completed").First(&order).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "No order found with tracking number " + req.TrackingNumber + " in picking completed status.",
		})
	}

	// Check if order processing_status is already "qc_completed"
	if order.ProcessingStatus == "qc_completed" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with tracking number " + req.TrackingNumber + " has already been QC completed.",
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
		if err := qcrc.DB.Where("id = ?", detailReq.BoxID).First(&box).Error; err != nil {
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

	// Create QCRibbon record
	qcRibbon := models.QCRibbon{
		TrackingNumber: req.TrackingNumber,
		QCBy:           uint(userID),
	}

	// Create QCRibbonDetails records
	for _, detailReq := range req.Details {
		qcRibbonDetail := models.QCRibbonDetail{
			BoxID:    detailReq.BoxID,
			Quantity: detailReq.Quantity,
		}
		qcRibbon.QCRibbonDetails = append(qcRibbon.QCRibbonDetails, qcRibbonDetail)
	}

	// Create records in the database (GORM will cascade to details automatically)
	if err := tx.Create(&qcRibbon).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create QC ribbon record",
		})
	}

	// Update order processing status to "qc_completed"
	if err := tx.Model(&models.Order{}).Where("tracking_number = ?", req.TrackingNumber).Update("processing_status", "qc_completed").Error; err != nil {
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
	if err := qcrc.DB.Preload("QCRibbonDetails.Box").Preload("QCUser").First(&qcRibbon, qcRibbon.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to load created QC ribbon",
		})
	}

	// Load order by tracking number
	var orderResponse models.Order
	if err := qcrc.DB.Preload("OrderDetails").Where("tracking_number = ?", qcRibbon.TrackingNumber).First(&orderResponse).Error; err == nil {
		qcRibbon.Order = &orderResponse
	}

	log.Println("CreateQCRibbon completed successfully")
	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "QC ribbon created successfully",
		Data:    qcRibbon.ToResponse(),
	})
}

// GetChartQcRibbons retrieves QC Ribbon data for charting
// @Summary Get Chart QC Ribbons
// @Description Retrieve QC Ribbon data for charting
// @Tags Ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=QcRibbonsDailyCountResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/ribbons/qc-ribbons/chart [get]
func (qcrc *QCRibbonController) GetChartQCRibbons(c fiber.Ctx) error {
	log.Println("GetChartQCRibbons called")
	// Get current month start and end dates
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	// Start of the month
	startOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	// First day of next month at 00:00:00 (to use as upper bound)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)

	// Query to get daily counts for current month
	var dailyCounts []QcRibbonDailyCount

	if err := qcrc.DB.Model(&models.QCRibbon{}).Select("DATE(created_at) as date, COUNT(*) as count").Where("created_at >= ? AND created_at < ?", startOfMonth, startOfNextMonth).Group("DATE(created_at)").Order("date ASC").Scan(&dailyCounts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve QC Ribbon data",
		})
	}

	// Get total count for the current month
	var totalCount int64
	if err := qcrc.DB.Model(&models.QCRibbon{}).Where("created_at >= ? AND created_at < ?", startOfMonth, startOfNextMonth).Count(&totalCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve total QC Ribbon count",
		})
	}

	// Format response
	response := QcRibbonsDailyCountResponse{
		Month:       currentMonth.String(),
		Year:        currentYear,
		DailyCounts: dailyCounts,
		TotalCount:  int(totalCount),
	}

	message := "QC Ribbon chart data " + currentMonth.String() + "  " + strconv.Itoa(currentYear) + " retrieved successfully"

	log.Println("GetChartQCRibbons completed successfully")
	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: message,
		Data:    response,
	})
}
