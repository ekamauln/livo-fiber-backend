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

type ComplainController struct {
	DB *gorm.DB
}

func NewComplainController(db *gorm.DB) *ComplainController {
	return &ComplainController{DB: db}
}

// Request structs
type CreateComplainRequest struct {
	TrackingNumber string `json:"trackingNumber" validate:"required,min=3,max=100"`
	ChannelID      uint   `json:"channelId" validate:"required"`
	StoreID        uint   `json:"storeId" validate:"required"`
	Reason         string `json:"reason" validate:"required"`
}

type UpdateComplainRequest struct {
	Solution    string                      `json:"solution" validate:"omitempty"`
	TotalFee    int                         `json:"totalFee" validate:"omitempty,min=0"`
	UserDetails []ComplainUserDetailRequest `json:"userDetails" validate:"omitempty,dive"`
}

type ComplainUserDetailRequest struct {
	UserID    uint `json:"userId" validate:"required"`
	FeeCharge int  `json:"feeCharge" validate:"required,min=0"`
}

type UpdateComplainCheckRequest struct {
	Checked bool `json:"checked" validate:"required"`
}

// GetComplains retrieves a list of complains with pagination and search
// @Summary Get Complains
// @Description Retrieve a list of complains with pagination and search
// @Tags Complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of complains per page" default(10)
// @Param startDate query string false "Start date (YYYY-MM-DD format)"
// @Param endDate query string false "End date (YYYY-MM-DD format)"
// @Param search query string false "Search term for tracking number or order ginee ID"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.ComplainResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/complains [get]
func (cc *ComplainController) GetComplains(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var complains []models.Complain

	// Build base query
	query := cc.DB.Preload("ComplainProductDetails").Preload("ComplainUserDetails").Preload("Channel").Preload("Store").Preload("CreateUser").Model(&models.Complain{}).Order("created_at DESC")

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
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("tracking_number LIKE ? OR order_ginee_id LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Limit(limit).Offset(offset).Find(&complains).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve complains",
		})
	}

	// Format response
	complainList := make([]models.ComplainResponse, len(complains))
	for i, complain := range complains {
		complainList[i] = *complain.ToComplainResponse()
	}

	// Build success message
	message := "Complains retrieved successfully"
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
		Data:    complainList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetComplain retrieves a single complain by ID
// @Summary Get Complain
// @Description Retrieve a single complain by ID
// @Tags Complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Complain ID"
// @Success 200 {object} utils.SuccessResponse{data=models.ComplainResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/complains/{id} [get]
func (cc *ComplainController) GetComplain(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var complain models.Complain
	if err := cc.DB.Preload("ComplainProductDetails").Preload("ComplainUserDetails").Preload("Channel").Preload("Store").Preload("CreateUser").Where("id = ?", id).First(&complain).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Complain with id " + id + " not found.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Complain retrieved successfully",
		Data:    complain.ToComplainResponse(),
	})
}

// CreateComplain handles the creation of a new complain
// @Summary Create Complain
// @Description Create a new complain
// @Tags Complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param complain body CreateComplainRequest true "Complain details"
// @Success 201 {object} utils.SuccessResponse{data=models.ComplainResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/complains [post]
func (cc *ComplainController) CreateComplain(c fiber.Ctx) error {
	// Parse request body
	var req CreateComplainRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Get the username of the creator from context (assuming middleware sets it)
	username := c.Locals("username").(string)

	// Check if tracking number already exists
	var existingComplain models.Complain
	if err := cc.DB.Where("tracking_number = ?", req.TrackingNumber).First(&existingComplain).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Complain with the same tracking number already exists.",
		})
	}

	// Generate complain code
	complainCode := utils.GenerateComplainCode(cc.DB, username, "")

	// Start transaction
	tx := cc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find order by tracking to get OrderGineeID and populate product details
	var order models.Order
	if err := tx.Preload("OrderDetails").Where("tracking_number = ?", req.TrackingNumber).First(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with tracking number " + req.TrackingNumber + " not found.",
		})
	}

	// Create complain record
	complain := models.Complain{
		Code:           complainCode,
		TrackingNumber: req.TrackingNumber,
		OrderGineeID:   order.OrderGineeID,
		ChannelID:      req.ChannelID,
		StoreID:        req.StoreID,
		CreatedBy:      c.Locals("userID").(uint),
		Reason:         req.Reason,
	}

	if err := tx.Create(&complain).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create complain",
		})
	}

	// Populate complain product details from order
	for _, orderDetail := range order.OrderDetails {
		complainProductDetail := models.ComplainProductDetail{
			ComplainID: complain.ID,
			ProductSKU: orderDetail.SKU,
			Quantity:   orderDetail.Quantity,
			Price:      orderDetail.Price,
		}

		if err := tx.Create(&complainProductDetail).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Failed to create complain product details",
			})
		}
	}

	// Populate complain user details with zero fee charge initially
	userIDs := make(map[uint]bool) // To avoid duplicate user details

	// Check qc ribbon
	var qcRibbon models.QCRibbon
	if err := tx.Where("tracking_number = ?", req.TrackingNumber).First(&qcRibbon).Error; err == nil && qcRibbon.QCBy != 0 {
		userIDs[qcRibbon.QCBy] = true

		// Update qc ribbon if tracking number is complained
		qcRibbon.Complained = true
		tx.Save(&qcRibbon)
	}

	// Check qc online
	var qcOnline models.QCOnline
	if err := tx.Where("tracking_number = ?", req.TrackingNumber).First(&qcOnline).Error; err == nil && qcOnline.QCBy != 0 {
		userIDs[qcOnline.QCBy] = true

		// Update qc online if tracking number is complained
		qcOnline.Complained = true
		tx.Save(&qcOnline)
	}

	// Check Outbound
	var outbound models.Outbound
	if err := tx.Where("tracking_number = ?", req.TrackingNumber).First(&outbound).Error; err == nil && outbound.OutboundBy != 0 {
		userIDs[outbound.OutboundBy] = true

		// Update outbound if tracking number is complained
		outbound.Complained = true
		tx.Save(&outbound)
	}

	// Check Order Assigned User
	var orderUser models.Order
	if err := tx.Where("tracking_number = ?", req.TrackingNumber).First(&orderUser).Error; err == nil && orderUser.PickedBy != nil && orderUser.AssignedBy != nil {
		userIDs[*orderUser.PickedBy] = true
		userIDs[*orderUser.AssignedBy] = true

		// Update order if tracking number is complained
		orderUser.Complained = true
		tx.Save(&orderUser)
	}

	// Create user details for each unique user found
	for userIDValue := range userIDs {
		userDetail := models.ComplainUserDetail{
			ComplainID: complain.ID,
			UserID:     userIDValue,
			FeeCharge:  0,
		}

		if err := tx.Create(&userDetail).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Failed to create complain user details",
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

	// Load created complain with related data
	if err := cc.DB.Preload("ComplainProductDetails").Preload("ComplainUserDetails").Preload("Channel").Preload("Store").Preload("CreateUser").Where("id = ?", complain.ID).First(&complain, complain.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve created complain",
		})
	}

	complain.Order = &order

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Complain created successfully",
		Data:    complain.ToComplainResponse(),
	})
}

// UpdateComplain updates an existing complain by ID
// @Summary Update Complain
// @Description Update an existing complain by ID
// @Tags Complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Complain ID"
// @Param complain body UpdateComplainRequest true "Updated complain details"
// @Success 200 {object} utils.SuccessResponse{data=models.ComplainResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/complains/{id} [put]
func (cc *ComplainController) UpdateComplain(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var complain models.Complain
	if err := cc.DB.Where("id = ?", id).First(&complain).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Complain with id " + id + " not found.",
		})
	}

	// Parse request body
	var req UpdateComplainRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Start transaction
	tx := cc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update complain fields if provided
	complain.Solution = &req.Solution
	complain.TotalFee = &req.TotalFee

	if err := tx.Save(&complain).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update complain",
		})
	}

	// handle user details update
	if len(req.UserDetails) > 0 {
		// Clear existing user details
		if err := tx.Where("complain_id = ?", complain.ID).Delete(&models.ComplainUserDetail{}).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Failed to clear existing complain user details",
			})
		}

		// Create new user details
		for _, userDetailReq := range req.UserDetails {
			userDetail := models.ComplainUserDetail{
				ComplainID: complain.ID,
				UserID:     userDetailReq.UserID,
				FeeCharge:  userDetailReq.FeeCharge,
			}

			if err := tx.Create(&userDetail).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
					Success: false,
					Error:   "Failed to create complain user details",
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

	// Load updated complain with related data
	if err := cc.DB.Preload("ComplainProductDetails").Preload("ComplainUserDetails").Preload("Channel").Preload("Store").Preload("CreateUser").Where("id = ?", complain.ID).First(&complain, complain.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve updated complain",
		})
	}

	// Load order data if tracking number exists
	if complain.TrackingNumber != "" {
		var order models.Order
		if err := cc.DB.Where("tracking_number = ?", complain.TrackingNumber).First(&order).Error; err == nil {
			complain.Order = &order
		}
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Complain updated successfully",
		Data:    complain.ToComplainResponse(),
	})
}

// UpdateComplainCheck updates the 'checked' status of a complain
// @Summary Update Complain Checked Status
// @Description Update the 'checked' status of a complain
// @Tags Complains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Complain ID"
// @Param checked body UpdateComplainCheckRequest true "Checked status"
// @Success 200 {object} utils.SuccessResponse{data=models.ComplainResponse}
// @Success 200 {object} utils.ErrorResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/complains/{id}/check [put]
func (cc *ComplainController) UpdateComplainCheck(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var complain models.Complain
	if err := cc.DB.Where("id = ?", id).First(&complain).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Complain with id " + id + " not found.",
		})
	}

	// Parse request body
	var req UpdateComplainCheckRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Update checked status
	complain.Checked = req.Checked
	if err := cc.DB.Save(&complain).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update complain checked status",
		})
	}

	// Load related data
	if err := cc.DB.Preload("ComplainProductDetails").Preload("ComplainUserDetails").Preload("Channel").Preload("Store").Preload("CreateUser").Where("id = ?", complain.ID).First(&complain, complain.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve updated complain",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Complain checked status updated successfully",
		Data:    complain.ToComplainResponse(),
	})
}
