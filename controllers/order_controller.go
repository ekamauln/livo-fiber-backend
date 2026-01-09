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

type OrderController struct {
	DB *gorm.DB
}

func NewOrderController(db *gorm.DB) *OrderController {
	return &OrderController{DB: db}
}

// Request structs
type CreateOrderRequest struct {
	OrderGineeID   string                     `json:"orderGineeId" validate:"required,min=3,max=100"`
	Channel        string                     `json:"channel" validate:"required,min=3,max=100"`
	Store          string                     `json:"store" validate:"required,min=3,max=100"`
	Buyer          string                     `json:"buyer" validate:"required,min=3,max=100"`
	Address        string                     `json:"address" validate:"required,min=3,max=255"`
	Courier        string                     `json:"courier" validate:"omitempty,min=3,max=100"`
	TrackingNumber string                     `json:"trackingNumber" validate:"omitempty,min=3,max=100"`
	SentBefore     string                     `json:"sentBefore" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	OrderDetails   []CreateOrderDetailRequest `json:"orderDetails" validate:"required,dive,required"`
}

type CreateOrderDetailRequest struct {
	SKU         string `json:"sku" validate:"required,min=1,max=255"`
	ProductName string `json:"productName" validate:"required,min=1,max=255"`
	Variant     string `json:"variant" validate:"omitempty,min=1,max=100"`
	Quantity    uint   `json:"quantity" validate:"required,gt=0"`
	Price       uint   `json:"price" validate:"required,gt=0"`
}

type BulkCreateOrdersRequest struct {
	Orders []CreateOrderRequest `json:"orders" validate:"required,dive,required"`
}

type UpdateOrderRequest struct {
	OrderDetails []UpdateOrderDetailRequest `json:"orderDetails" validate:"required,dive,required"`
}

type UpdateOrderDetailRequest struct {
	SKU         string `json:"sku" validate:"required,min=1,max=255"`
	ProductName string `json:"productName" validate:"required,min=1,max=255"`
	Variant     string `json:"variant" validate:"omitempty,min=1,max=100"`
	Quantity    uint   `json:"quantity" validate:"required,gt=0"`
	Price       uint   `json:"price" validate:"required,gt=0"`
}

type UpdateProcessingStatusRequest struct {
	ProcessingStatus string `json:"processingStatus" validate:"required,min=3,max=50"`
}

type UpdateEventStatusRequest struct {
	EventStatus string `json:"eventStatus" validate:"required,min=3,max=50"`
}

type AssignPickerRequest struct {
	PickerID       uint   `json:"pickerId" validate:"required"`
	TrackingNumber string `json:"trackingNumber" validate:"required,min=3,max=100"`
}

// Unique Response structs
type BulkCreateOrdersReponse struct {
	Summary       BulkCreateSummary      `json:"summary"`
	CreatedOrders []models.OrderResponse `json:"createdOrders"`
	SkippedOrders []SkippedOrder         `json:"skippedOrders"`
	FailedOrders  []FailedOrder          `json:"failedOrders"`
}

type BulkCreateSummary struct {
	Total   uint `json:"total"`
	Created uint `json:"created"`
	Skipped uint `json:"skipped"`
	Failed  uint `json:"failed"`
}

type SkippedOrder struct {
	Index          uint   `json:"index"`
	OrderGineeID   string `json:"orderGineeId"`
	TrackingNumber string `json:"trackingNumber"`
	Reason         string `json:"reason"`
}

type FailedOrder struct {
	Index          uint   `json:"index"`
	OrderGineeID   string `json:"orderGineeId"`
	TrackingNumber string `json:"trackingNumber"`
	Error          string `json:"error"`
}

type DuplicatedOrderResponse struct {
	OriginalOrder   models.OrderResponse `json:"originalOrder"`
	DuplicatedOrder models.OrderResponse `json:"duplicatedOrder"`
}

// GetOrders retrieves a list of orders with pagination and search
// @Summary Get Orders
// @Description Retrieve a list of orders with pagination and search
// @Tags Orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of orders per page" default(10)
// @Param start_date query string false "Start date (YYYY-MM-DD format)"
// @Param end_date query string false "End date (YYYY-MM-DD format)"
// @Param search query string false "Search term for order ginee id or tracking number"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.Order}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/orders [get]
func (oc *OrderController) GetOrders(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var orders []models.Order

	// Build base query
	query := oc.DB.Model(&models.Order{}).Preload("OrderDetails").Preload("AssignUser").Preload("PickUser").Preload("PendingUser").Preload("ChangeUser").Preload("DuplicateUser").Preload("CancelUser").Order("created_at DESC")

	// Date range filter if provided
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	if startDate != "" && endDate != "" {
		// Parse start date and set time to beginning of the day
		if parsedStartDate, err := time.Parse("2006-01-02", startDate); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Invalid start_date format. Use YYYY-MM-DD.",
			})
		} else {
			startOfDay := parsedStartDate.Format("2006-01-02 00:00:00")
			query = query.Where("created_at >= ?", startOfDay)
		}
	}
	if endDate != "" {
		// Parse end date and set time to end of the day
		if parsedEndDate, err := time.Parse("2006-01-02", endDate); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Invalid end_date format. Use YYYY-MM-DD.",
			})
		} else {
			endOfDay := parsedEndDate.Format("2006-01-02 23:59:59")
			query = query.Where("created_at <= ?", endOfDay)
		}
	}

	// Search condition if provided
	search := c.Query("search", "")
	if search != "" {
		query = query.Where("ginee_id ILIKE ? OR tracking_number ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve orders",
		})
	}

	// Format response
	orderList := make([]models.OrderResponse, len(orders))
	for i, order := range orders {
		orderList[i] = *order.ToOrderResponse()
	}

	// Build success message
	message := "Orders retrieved successfully"
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
		Data:    orderList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetOrder retrieves a single order by ID
// @Summary Get Order
// @Description Retrieve a single order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Order}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/orders/{id} [get]
func (oc *OrderController) GetOrder(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var order models.Order
	if err := oc.DB.Where("id = ?", id).Preload("OrderDetails").Preload("AssignUser").Preload("PickUser").Preload("PendingUser").Preload("ChangeUser").Preload("DuplicateUser").Preload("CancelUser").First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with id " + id + " not found.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Order retrieved successfully",
		Data:    order.ToOrderResponse(),
	})
}

// CreateOrder creates a new order
// @Summary Create Order
// @Description Create a new order
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order details"
// @Success 201 {object} utils.SuccessResponse{data=models.Order}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/orders [post]
func (oc *OrderController) CreateOrder(c fiber.Ctx) error {
	// Binding request body
	var req CreateOrderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert Order Ginee ID to uppercase and trim spaces
	req.OrderGineeID = strings.ToUpper(strings.TrimSpace(req.OrderGineeID))

	// Convert Tracking Number to uppercase and trim spaces
	req.TrackingNumber = strings.ToUpper(strings.TrimSpace(req.TrackingNumber))

	// Check for existing order with same Order Ginee ID or Tracking Number
	var existingOrder models.Order
	if err := oc.DB.Where("ginee_id = ? OR tracking_number = ?", req.OrderGineeID, req.TrackingNumber).First(&existingOrder).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with Order Ginee ID " + req.OrderGineeID + " or Tracking Number " + req.TrackingNumber + " already exists.",
		})
	}

	// Parse Sent Before date if provided
	var sentBefore time.Time
	if req.SentBefore != "" {
		var err error
		sentBefore, err = time.Parse("2006-01-02 15:04:00", req.SentBefore)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Invalid sentBefore format. Use YYYY-MM-DD HH:MM:SS format.",
			})
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "sentBefore is required",
		})
	}

	// Start transaction
	tx := oc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create new order
	newOrder := models.Order{
		OrderGineeID:     req.OrderGineeID,
		ProcessingStatus: "ready to pick",
		Channel:          req.Channel,
		Store:            req.Store,
		Buyer:            req.Buyer,
		Address:          req.Address,
		Courier:          req.Courier,
		TrackingNumber:   req.TrackingNumber,
		SentBefore:       sentBefore,
	}

	if err := tx.Create(&newOrder).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create order",
		})
	}

	// Create order details
	var orderDetails []models.OrderDetail
	for _, detail := range req.OrderDetails {
		orderDetail := models.OrderDetail{
			OrderID:     newOrder.ID,
			SKU:         detail.SKU,
			ProductName: detail.ProductName,
			Variant:     detail.Variant,
			Quantity:    detail.Quantity,
			Price:       detail.Price,
		}
		orderDetails = append(orderDetails, orderDetail)
	}

	if err := tx.Create(&orderDetails).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create order details",
		})
	}

	// Commit transaction
	tx.Commit()

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Order created successfully",
		Data:    newOrder.ToOrderResponse(),
	})
}

// BulkCreateOrders creates multiple orders in a single request
// @Summary Bulk Create Orders
// @Description Create multiple orders in a single request
// @Tags Orders
// @Accept json
// @Produce json
// @Param orders body BulkCreateOrdersRequest true "List of orders to create"
// @Success 201 {object} utils.SuccessResponse{data=BulkCreateOrdersReponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/orders/bulk [post]
func (oc *OrderController) BulkCreateOrders(c fiber.Ctx) error {
	// Binding request body
	var req BulkCreateOrdersRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	var createdOrders []models.Order
	var skippedOrders []SkippedOrder
	var failedOrders []FailedOrder

	for i, orderReq := range req.Orders {
		// Convert Order Ginee ID to uppercase and trim spaces
		orderReq.OrderGineeID = strings.ToUpper(strings.TrimSpace(orderReq.OrderGineeID))

		// Convert Tracking Number to uppercase and trim spaces
		orderReq.TrackingNumber = strings.ToUpper(strings.TrimSpace(orderReq.TrackingNumber))

		// Check if order with same OrderGineeID or tracking number already exists
		var existingOrder models.Order
		if err := oc.DB.Where("order_ginee_id = ? OR tracking_number = ?", orderReq.OrderGineeID, orderReq.TrackingNumber).First(&existingOrder).Error; err == nil {
			// If order already exists, skip it
			skippedOrders = append(skippedOrders, SkippedOrder{
				Index:          uint(i),
				OrderGineeID:   orderReq.OrderGineeID,
				TrackingNumber: orderReq.TrackingNumber,
				Reason:         "Order already exists",
			})
			continue
		}

		// Create order
		order := models.Order{
			OrderGineeID:     orderReq.OrderGineeID,
			ProcessingStatus: "ready to pick",
			Channel:          orderReq.Channel,
			Store:            orderReq.Store,
			Buyer:            orderReq.Buyer,
			Address:          orderReq.Address,
			Courier:          orderReq.Courier,
			TrackingNumber:   orderReq.TrackingNumber,
		}

		if orderReq.SentBefore != "" {
			if parsedTime, err := time.Parse("2006-01-02 15:04:00", orderReq.SentBefore); err == nil {
				order.SentBefore = parsedTime
			} else {
				// Failed to parse date
				failedOrders = append(failedOrders, FailedOrder{
					Index:        uint(i),
					OrderGineeID: orderReq.OrderGineeID,
					Error:        "Invalid sentBefore format: " + err.Error(),
				})
				continue
			}
		}

		// Create order details
		for _, detailReq := range orderReq.OrderDetails {
			orderDetail := models.OrderDetail{
				SKU:         detailReq.SKU,
				ProductName: detailReq.ProductName,
				Variant:     detailReq.Variant,
				Quantity:    detailReq.Quantity,
				Price:       detailReq.Price,
			}
			order.OrderDetails = append(order.OrderDetails, orderDetail)
		}

		// Try to create the order using transaction
		tx := oc.DB.Begin()
		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			// Failed to create order
			failedOrders = append(failedOrders, FailedOrder{
				Index:        uint(i),
				OrderGineeID: orderReq.OrderGineeID,
				Error:        err.Error(),
			})
			continue
		}
		tx.Commit()

		// Load order with details for response
		oc.DB.Preload("OrderDetails").First(&order, order.ID)
		createdOrders = append(createdOrders, order)
	}

	// Format response
	createdOrderResponses := make([]models.OrderResponse, len(createdOrders))
	for i, order := range createdOrders {
		createdOrderResponses[i] = *order.ToOrderResponse()
	}

	response := BulkCreateOrdersReponse{
		Summary: BulkCreateSummary{
			Total:   uint(len(req.Orders)),
			Created: uint(len(createdOrders)),
			Skipped: uint(len(skippedOrders)),
			Failed:  uint(len(failedOrders)),
		},
		CreatedOrders: createdOrderResponses,
		SkippedOrders: skippedOrders,
		FailedOrders:  failedOrders,
	}

	// Build success message
	statusCode := fiber.StatusCreated
	message := "Bulk order creation completed"

	if len(createdOrders) == 0 {
		if len(skippedOrders) > 0 {
			statusCode = fiber.StatusOK
			message = "All orders were skipped (already exist)"
		} else {
			statusCode = fiber.StatusBadRequest
			message = "No orders could be created"
		}
	} else if len(failedOrders) > 0 || len(skippedOrders) > 0 {
		message = "Bulk order creation completed with some issues"
	}

	// Return response
	return c.Status(statusCode).JSON(utils.SuccessResponse{
		Success: true,
		Message: message,
		Data:    response,
	})
}

// UpdateOrder updates an existing order
// @Summary Update Order
// @Description Update an existing order
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param order body UpdateOrderRequest true "Updated order details"
// @Success 200 {object} utils.SuccessResponse{data=models.Order}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/orders/{id} [put]
func (oc *OrderController) UpdateOrder(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var order models.Order
	if err := oc.DB.Where("id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with id " + id + " not found.",
		})
	}

	// Binding request body
	var req UpdateOrderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
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

	// Check if order processing status allows modification
	if order.ProcessingStatus == "picking process" || order.ProcessingStatus == "qc process" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order cannot be modified in " + order.ProcessingStatus + " status.",
		})
	}

	// Check if order is canceled
	if order.EventStatus != nil && *order.EventStatus == "canceled" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Canceled order cannot be modified.",
		})
	}

	// Start transaction
	tx := oc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update changed_by and changed_at fields
	now := time.Now()
	userIDUint := uint(userID)
	order.ChangedBy = &userIDUint
	order.ChangedAt = &now

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update order",
		})
	}

	// Update order details if provided - replace all details
	if req.OrderDetails != nil {
		// Delete all existing order details
		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderDetail{}).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Failed to update order details",
			})
		}

		// Create new order details
		newDetails := make([]models.OrderDetail, 0, len(req.OrderDetails))
		for _, detailReq := range req.OrderDetails {
			detail := models.OrderDetail{
				OrderID:     order.ID,
				SKU:         detailReq.SKU,
				ProductName: detailReq.ProductName,
				Variant:     detailReq.Variant,
				Quantity:    detailReq.Quantity,
				Price:       detailReq.Price,
			}
			newDetails = append(newDetails, detail)
		}

		if len(newDetails) > 0 {
			if err := tx.Create(&newDetails).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
					Success: false,
					Error:   "Failed to update order details",
				})
			}
		}

		// Update order's OrderDetails field
		order.OrderDetails = newDetails
	}

	// Coommit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update order",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Order updated successfully",
		Data:    order.ToOrderResponse(),
	})
}

// DuplicateOrder duplicates an existing order
// @Summary Duplicate Order
// @Description Duplicate an existing order
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 201 {object} utils.SuccessResponse{data=DuplicatedOrderResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/orders/{id}/duplicate [put]
func (oc *OrderController) DuplicateOrder(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var order models.Order
	if err := oc.DB.Where("id = ?", id).Preload("OrderDetails").First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order with id " + id + " not found.",
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

	// Check if order processing status allows modification
	if order.ProcessingStatus == "picking process" || order.ProcessingStatus == "qc process" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order cannot be duplicated in " + order.ProcessingStatus + " status.",
		})
	}

	// Check if order is canceled
	if order.EventStatus != nil && *order.EventStatus == "canceled" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Canceled order cannot be duplicated.",
		})
	}

	// Check if order event status has been duplicated
	if order.EventStatus != nil && *order.EventStatus == "duplicated" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Order has already been duplicated.",
		})
	}

	// Start transaction
	tx := oc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Store original tracking number before duplication
	originalTrackingNumber := order.TrackingNumber
	newTrackingNumber := "X-" + originalTrackingNumber

	// Update original order's order ginee id by adding "-X2" suffix and tracking number with "X-" prefix
	now := time.Now()
	userIDUint := uint(userID)
	eventStatusDuplicated := "duplicated"
	order.EventStatus = &eventStatusDuplicated
	order.OrderGineeID = order.OrderGineeID + "-X2"
	order.TrackingNumber = newTrackingNumber
	order.DuplicatedBy = &userIDUint
	order.DuplicatedAt = &now

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update original order for duplication",
		})
	}

	// Update tracking number in qc ribbon, qc online, and outbound if exists (ignore errors if table doesn't exist)
	tx.Model(&models.QCRibbon{}).Where("tracking_number = ?", originalTrackingNumber).Update("tracking_number", newTrackingNumber)

	tx.Model(&models.QCOnline{}).Where("tracking_number = ?", originalTrackingNumber).Update("tracking_number", newTrackingNumber)

	tx.Model(&models.Outbound{}).Where("tracking_number = ?", originalTrackingNumber).Update("tracking_number", newTrackingNumber)

	// Create duplicated order
	duplicatedEventStatus := "duplicated"
	duplicatedOrder := models.Order{
		OrderGineeID:     order.OrderGineeID[:len(order.OrderGineeID)-3], // Remove "-X2" suffix
		ProcessingStatus: order.ProcessingStatus,
		Channel:          order.Channel,
		Store:            order.Store,
		Buyer:            order.Buyer,
		Address:          order.Address,
		Courier:          order.Courier,
		TrackingNumber:   originalTrackingNumber,
		SentBefore:       order.SentBefore,
		EventStatus:      &duplicatedEventStatus,
		DuplicatedBy:     &userIDUint,
		DuplicatedAt:     &now,
	}

	// Duplicate order details
	for _, detail := range order.OrderDetails {
		duplicatedDetail := models.OrderDetail{
			SKU:         detail.SKU,
			ProductName: detail.ProductName,
			Variant:     detail.Variant,
			Quantity:    detail.Quantity,
			Price:       detail.Price,
		}
		duplicatedOrder.OrderDetails = append(duplicatedOrder.OrderDetails, duplicatedDetail)
	}

	// Create duplicated order in database
	if err := tx.Create(&duplicatedOrder).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create duplicated order",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to commit transaction",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Order duplicated successfully",
		Data: map[string]interface{}{
			"originalOrder":   order.ToOrderResponse(),
			"duplicatedOrder": duplicatedOrder.ToOrderResponse(),
		},
	})
}
