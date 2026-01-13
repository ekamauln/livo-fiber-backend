package controllers

import (
	"fmt"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type ReturnController struct {
	DB *gorm.DB
}

func NewReturnController(db *gorm.DB) *ReturnController {
	return &ReturnController{DB: db}
}

// Request structs
type CreateReturnRequest struct {
	NewTrackingNumber string  `json:"newTrackingNumber" validate:"required"`
	ChannelID         uint    `json:"channelId" validate:"required"`
	StoreID           uint    `json:"storeId" validate:"required"`
	TrackingNumber    *string `json:"trackingNumber,omitempty"`
	ReturnType        *string `json:"returnType,omitempty"`
	ReturnReason      *string `json:"returnReason,omitempty"`
	ReturnNumber      *string `json:"returnNumber,omitempty"`
	ScrapNumber       *string `json:"scrapNumber,omitempty"`
}

type UpdateReturnRequest struct {
	TrackingNumber *string `json:"trackingNumber,omitempty"`
	ReturnType     *string `json:"returnType,omitempty"`
	ReturnReason   *string `json:"returnReason,omitempty"`
	ReturnNumber   *string `json:"returnNumber,omitempty"`
	ScrapNumber    *string `json:"scrapNumber,omitempty"`
}

// GetReturns retrieves a list of returns with pagination and search
// @Summary Get Returns
// @Description Retrieve a list of returns with pagination and search
// @Tags Returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of returns per page" default(10)
// @Param search query string false "Search term for new tracking number, order ginee ID, or old tracking number"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.ReturnResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/returns [get]
func (rc *ReturnController) GetReturns(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var returns []models.Return

	// Build base query
	query := rc.DB.Preload("ReturnDetails").Preload("CreateUser").Preload("UpdateUser").Model(&models.Return{}).Order("created_at DESC")

	// Date range filter if provided
	startDate := c.Query("startDate", "")
	endDate := c.Query("endDate", "")
	if startDate != "" {
		// Parse start date and set time to beginning of the day
		parsedStartDate, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Invalid start_date format. Use YYYY-MM-DD.",
			})
		}
		startOfDay := time.Date(parsedStartDate.Year(), parsedStartDate.Month(), parsedStartDate.Day(), 0, 0, 0, 0, parsedStartDate.Location())
		query = query.Where("created_at >= ?", startOfDay)
	}
	if endDate != "" {
		// Parse end date and set time to end of the day
		parsedEndDate, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Invalid end_date format. Use YYYY-MM-DD.",
			})
		}
		endOfDay := time.Date(parsedEndDate.Year(), parsedEndDate.Month(), parsedEndDate.Day(), 23, 59, 59, 0, parsedEndDate.Location())
		query = query.Where("created_at <= ?", endOfDay)
	}

	// Search condition if provided
	search := c.Query("search", "")
	if search != "" {
		query = query.Where("new_tracking_number ILIKE ? OR order_ginee_id ILIKE ? OR old_tracking_number ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Offset(offset).Limit(limit).Find(&returns).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve returns",
		})
	}

	// Include Order details for each return if TrackingNumber exists in Order
	for i := range returns {
		if returns[i].TrackingNumber != nil {
			var order models.Order
			if err := rc.DB.Where("tracking_number = ?", returns[i].TrackingNumber).First(&order).Error; err == nil {
				returns[i].Order = &order
			}
		}
	}

	// Format response
	returnList := make([]models.ReturnResponse, len(returns))
	for i, ret := range returns {
		returnList[i] = ret.ToResponse()
	}

	// Build success message
	message := "Returns retrieved successfully"
	var filters []string

	if startDate != "" || endDate != "" {
		var dateRange []string
		if startDate != "" {
			dateRange = append(dateRange, "from: "+startDate)
		}
		if endDate != "" {
			dateRange = append(dateRange, "to: "+endDate)
		}
		filters = append(filters, "date: "+strings.Join(dateRange, ", "))
	}

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
		Data:    returnList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetReturn retrieves a single return by ID
// @Summary Get Return
// @Description Retrieve a single return by ID
// @Tags Returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Return ID"
// @Success 200 {object} utils.SuccessResponse{data=models.ReturnResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/returns/{id} [get]
func (rc *ReturnController) GetReturn(c fiber.Ctx) error {
	// Parse id parameters
	id := c.Params("id")
	var ret models.Return
	if err := rc.DB.Preload("ReturnDetails").Preload("CreateUser").Preload("UpdateUser").First(&ret, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Return with id " + id + " not found",
		})
	}

	// Include Order details if TrackingNumber exists in Order
	if ret.TrackingNumber != nil {
		var order models.Order
		if err := rc.DB.Where("tracking_number = ?", ret.TrackingNumber).First(&order).Error; err == nil {
			ret.Order = &order
		}
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Return retrieved successfully",
		Data:    ret.ToResponse(),
	})
}

// CreateReturn handles the creation of a new return with details from order details table
// @Summary Create Return
// @Description Create a new return
// @Tags Returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateReturnRequest true "Return details"
// @Success 201 {object} utils.SuccessResponse{data=models.Return}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/returns [post]
func (rc *ReturnController) CreateReturn(c fiber.Ctx) error {
	// Binding request body
	var req CreateReturnRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Get current user logged in user
	userIDStr := c.Locals("user_id").(string)
	UserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// Convert new tracking number to uppercase and trim spaces
	req.NewTrackingNumber = strings.ToUpper(strings.TrimSpace(req.NewTrackingNumber))

	// Check for duplicate NewTrackingNumber
	var existingReturn models.Return
	if err := rc.DB.Where("new_tracking_number = ?", req.NewTrackingNumber).First(&existingReturn).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Return with new tracking number " + req.NewTrackingNumber + " already exists",
		})
	}

	// If TrackingNumber is provided, convert to uppercase and trim spaces and check if it's exists in Order
	if req.TrackingNumber != nil {
		tracking := strings.ToUpper(strings.TrimSpace(*req.TrackingNumber))
		req.TrackingNumber = &tracking
	}

	var order models.Order
	if err := rc.DB.Preload("OrderDetails").Where("tracking_number = ?", req.TrackingNumber).First(&order).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with tracking number " + *req.TrackingNumber + " not found",
		})
	}

	// Start database transaction
	tx := rc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	ret := models.Return{
		NewTrackingNumber: req.NewTrackingNumber,
		TrackingNumber:    req.TrackingNumber,
		ChannelID:         req.ChannelID,
		StoreID:           req.StoreID,
		ReturnType:        req.ReturnType,
		ReturnReason:      req.ReturnReason,
		ReturnNumber:      req.ReturnNumber,
		ScrapNumber:       req.ScrapNumber,
		CreatedBy:         uint(UserID),
		OrderGineeID:      &order.OrderGineeID,
	}

	if err := tx.Create(&ret).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create return",
		})
	}

	// Create ReturnDetails from OrderDetails
	for _, orderDetail := range order.OrderDetails {
		returnDetail := models.ReturnDetail{
			ReturnID:   &ret.ID,
			ProductSKU: &orderDetail.SKU,
			Quantity:   &orderDetail.Quantity,
			Price:      &orderDetail.Price,
		}

		if err := tx.Create(&returnDetail).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Failed to create return details",
			})
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to commit transaction",
		})
	}

	// Reload return with details
	if err := rc.DB.Preload("ReturnDetails").Preload("CreateUser").Preload("UpdateUser").First(&ret, ret.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve created return",
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Return created successfully",
		Data:    ret.ToResponse(),
	})
}

// UpdateReturn handles updating an existing return and if details still empty, populate from order details
// @Summary Update Return
// @Description Update an existing return
// @Tags Returns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Return ID"
// @Param request body UpdateReturnRequest true "Return details to update"
// @Success 200 {object} utils.SuccessResponse{data=models.ReturnResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/returns/{id} [put]
func (rc *ReturnController) UpdateReturn(c fiber.Ctx) error {
	// Parse id parameters
	id := c.Params("id")
	var ret models.Return
	if err := rc.DB.Preload("ReturnDetails").First(&ret, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Return with id " + id + " not found",
		})
	}

	// Binding request body
	var req UpdateReturnRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Get current user logged in user
	userIDStr := c.Locals("user_id").(string)
	UserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// Check if TrackingNumber is provided, convert to uppercase and trim spaces and check if it's exists in Order
	if req.TrackingNumber != nil {
		tracking := strings.ToUpper(strings.TrimSpace(*req.TrackingNumber))
		req.TrackingNumber = &tracking
	}

	// Check if TrackingNumber exists in Order
	var order models.Order
	if req.TrackingNumber != nil {
		if err := rc.DB.Preload("OrderDetails").Where("tracking_number = ?", req.TrackingNumber).First(&order).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Order with tracking number " + *req.TrackingNumber + " not found",
			})
		}
	}

	// Check if TrackingNumber is already exists in another Return
	if req.TrackingNumber != nil {
		var existingReturn models.Return
		if err := rc.DB.Where("tracking_number = ? AND id <> ?", req.TrackingNumber, ret.ID).First(&existingReturn).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Return with tracking number " + *req.TrackingNumber + " already exists",
			})
		}
	}

	// Check for return details before transaction
	needToPopulateDetails := len(*ret.ReturnDetails) == 0 && req.TrackingNumber != nil

	// Start database transaction
	tx := rc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update return fields
	updatedBy := uint(UserID)

	ret.TrackingNumber = req.TrackingNumber
	ret.ReturnType = req.ReturnType
	ret.ReturnReason = req.ReturnReason
	ret.ReturnNumber = req.ReturnNumber
	ret.ScrapNumber = req.ScrapNumber
	ret.UpdatedBy = &updatedBy

	if err := tx.Save(&ret).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update return",
		})
	}

	// Populate ReturnDetails from OrderDetails if needed
	if needToPopulateDetails {
		for _, orderDetail := range order.OrderDetails {
			returnDetail := models.ReturnDetail{
				ReturnID:   &ret.ID,
				ProductSKU: &orderDetail.SKU,
				Quantity:   &orderDetail.Quantity,
				Price:      &orderDetail.Price,
			}

			if err := tx.Create(&returnDetail).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
					Success: false,
					Error:   "Failed to create return details",
				})
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to commit transaction",
		})
	}

	// Reload return with details
	if err := rc.DB.Preload("ReturnDetails").Preload("CreateUser").Preload("UpdateUser").First(&ret, ret.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve updated return",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Return updated successfully",
		Data:    ret.ToResponse(),
	})
}
