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

type RibbonFlowController struct {
	DB *gorm.DB
}

func NewRibbonFlowController(db *gorm.DB) *RibbonFlowController {
	return &RibbonFlowController{DB: db}
}

// Uniqe response struct for RibbonFlow
type RibbonUserFlowInfo struct {
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

type QcRibbonFlowInfo struct {
	User      *RibbonUserFlowInfo `json:"qcBy,omitempty"`
	CreatedAt string              `json:"createdAt"`
}

type RibbonOutboundFlowInfo struct {
	User            *RibbonUserFlowInfo `json:"outboundBy,omitempty"`
	Expedition      string              `json:"expedition"`
	ExpeditionColor string              `json:"expeditionColor"`
	CreatedAt       string              `json:"createdAt"`
}

type RibbonOrderFlowInfo struct {
	TrackingNumber   string              `json:"trackingNumber"`
	ProcessingStatus string              `json:"processingStatus"`
	OrderGineeID     string              `json:"orderGineeId"`
	Complained       bool                `json:"complained"`
	CreatedAt        string              `json:"createdAt"`
	AssignedBy       *RibbonUserFlowInfo `json:"assignedBy,omitempty"`
	AssignedAt       *string             `json:"assignedAt,omitempty"`
	PickedBy         *RibbonUserFlowInfo `json:"pickedBy,omitempty"`
	PickedAt         *string             `json:"pickedAt,omitempty"`
	PendingBy        *RibbonUserFlowInfo `json:"pendingBy,omitempty"`
	PendingAt        *string             `json:"pendingAt,omitempty"`
	ChangedBy        *RibbonUserFlowInfo `json:"changedBy,omitempty"`
	ChangedAt        *string             `json:"changedAt,omitempty"`
	DuplicatedBy     *RibbonUserFlowInfo `json:"duplicatedBy,omitempty"`
	DuplicatedAt     *string             `json:"duplicatedAt,omitempty"`
	CancelledBy      *RibbonUserFlowInfo `json:"cancelledBy,omitempty"`
	CancelledAt      *string             `json:"cancelledAt,omitempty"`
}

type RibbonFlowResponse struct {
	TrackingNumber string                  `json:"trackingNumber"`
	QCRibbon       *QcRibbonFlowInfo       `json:"qcRibbon,omitempty"`
	Outbound       *RibbonOutboundFlowInfo `json:"outbound,omitempty"`
	Order          *RibbonOrderFlowInfo    `json:"order,omitempty"`
}

type RibbonFlowsListResponse struct {
	RibbonFlows []RibbonFlowResponse `json:"ribbonFlows"`
}

// Helper function to build RibbonFlowResponse for a given tracking number
func (rfc *RibbonFlowController) BuildRibbonFlow(trackingNumber string) RibbonFlowResponse {
	var response RibbonFlowResponse
	response.TrackingNumber = trackingNumber

	// 1. Query QC Ribbon (Primary Table)
	var qcRibbon models.QCRibbon
	if err := rfc.DB.Preload("QCRibbonDetails").Preload("QCUser").Where("tracking_number = ?", trackingNumber).First(&qcRibbon).Error; err == nil {
		var user *RibbonUserFlowInfo
		if qcRibbon.QCUser != nil {
			user = &RibbonUserFlowInfo{
				Username: qcRibbon.QCUser.Username,
				FullName: qcRibbon.QCUser.FullName,
			}
		}

		response.QCRibbon = &QcRibbonFlowInfo{
			User:      user,
			CreatedAt: qcRibbon.CreatedAt.Format("02-01-2006 15:04:05"),
		}
	}

	// 2. Query Outbound table
	var outbound models.Outbound
	if err := rfc.DB.Preload("OutboundUser").Where("tracking_number = ?", trackingNumber).First(&outbound).Error; err == nil {
		var user *RibbonUserFlowInfo
		if outbound.OutboundUser != nil {
			user = &RibbonUserFlowInfo{
				Username: outbound.OutboundUser.Username,
				FullName: outbound.OutboundUser.FullName,
			}
		}

		response.Outbound = &RibbonOutboundFlowInfo{
			User:            user,
			Expedition:      outbound.Expedition,
			ExpeditionColor: outbound.ExpeditionColor,
			CreatedAt:       outbound.CreatedAt.Format("02-01-2006 15:04:05"),
		}
	}

	// 3. Query Order table
	var order models.Order
	if err := rfc.DB.Preload("OrderDetails").Preload("AssignUser").Preload("PickUser").Preload("PendingUser").Preload("ChangeUser").Preload("DuplicateUser").Preload("CancelUser").Preload("DuplicateUser").Where("tracking_number = ?", trackingNumber).First(&order).Error; err == nil {
		orderInfo := &RibbonOrderFlowInfo{
			TrackingNumber:   order.TrackingNumber,
			ProcessingStatus: order.ProcessingStatus,
			OrderGineeID:     order.OrderGineeID,
			Complained:       order.Complained,
			CreatedAt:        order.CreatedAt.Format("02-01-2006 15:04:05"),
		}

		// user visual handlers
		if order.AssignedBy != nil && order.AssignUser != nil {
			orderInfo.AssignedBy = &RibbonUserFlowInfo{
				Username: order.AssignUser.Username,
				FullName: order.AssignUser.FullName,
			}
			if order.AssignedAt != nil {
				assignedAt := order.AssignedAt.Format("02-01-2006 15:04:05")
				orderInfo.AssignedAt = &assignedAt
			}
		}

		if order.PickedBy != nil && order.PickUser != nil {
			orderInfo.PickedBy = &RibbonUserFlowInfo{
				Username: order.PickUser.Username,
				FullName: order.PickUser.FullName,
			}
			if order.PickedAt != nil {
				pickedAt := order.PickedAt.Format("02-01-2006 15:04:05")
				orderInfo.PickedAt = &pickedAt
			}
		}

		if order.PendingBy != nil && order.PendingUser != nil {
			orderInfo.PendingBy = &RibbonUserFlowInfo{
				Username: order.PendingUser.Username,
				FullName: order.PendingUser.FullName,
			}
			if order.PendingAt != nil {
				pendingAt := order.PendingAt.Format("02-01-2006 15:04:05")
				orderInfo.PendingAt = &pendingAt
			}
		}

		if order.ChangedBy != nil && order.ChangeUser != nil {
			orderInfo.ChangedBy = &RibbonUserFlowInfo{
				Username: order.ChangeUser.Username,
				FullName: order.ChangeUser.FullName,
			}
			if order.ChangedAt != nil {
				changedAt := order.ChangedAt.Format("02-01-2006 15:04:05")
				orderInfo.ChangedAt = &changedAt
			}
		}

		if order.CanceledBy != nil && order.CancelUser != nil {
			orderInfo.CancelledBy = &RibbonUserFlowInfo{
				Username: order.CancelUser.Username,
				FullName: order.CancelUser.FullName,
			}
			if order.CanceledAt != nil {
				canceledAt := order.CanceledAt.Format("02-01-2006 15:04:05")
				orderInfo.CancelledAt = &canceledAt
			}
		}

		if order.DuplicatedBy != nil && order.DuplicateUser != nil {
			orderInfo.DuplicatedBy = &RibbonUserFlowInfo{
				Username: order.DuplicateUser.Username,
				FullName: order.DuplicateUser.FullName,
			}
			if order.DuplicatedAt != nil {
				duplicatedAt := order.DuplicatedAt.Format("02-01-2006 15:04:05")
				orderInfo.DuplicatedAt = &duplicatedAt
			}
		}
		response.Order = orderInfo
	}
	return response
}

// GetRibbonFlows retrieves all ribbon flows with their associated QC Ribbon, Outbound, and Order details (with pagination and search)
// @Summary Get all ribbon flows
// @Description Retrieve all ribbon flows with their associated QC Ribbon, Outbound, and Order details (with pagination and search)
// @Tags Ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number for pagination" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param startDate query string false "Filter by start date (YYYY-MM-DD format)"
// @Param endDate query string false "Filter by end date (YYYY-MM-DD format)"
// @Param search query string false "Search term for tracking number"
// @Success 200 {array} utils.SuccessPaginatedResponse{data=[]RibbonFlowsListResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/ribbons/flows [get]
func (rfc *RibbonFlowController) GetRibbonFlows(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var qcRibbons []models.QCRibbon

	// Build base query
	query := rfc.DB.Preload("QCRibbonDetails").Preload("QCUser").Model(&models.QCRibbon{}).Order("created_at DESC")

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
		query = query.Where("tracking_number LIKE ?", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Offset(offset).Limit(limit).Find(&qcRibbons).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve ribbon flows",
		})
	}

	// Format response
	var ribbonFlows []RibbonFlowResponse
	for _, qcRibbon := range qcRibbons {
		flow := rfc.BuildRibbonFlow(qcRibbon.TrackingNumber)
		ribbonFlows = append(ribbonFlows, flow)
	}

	response := RibbonFlowsListResponse{
		RibbonFlows: ribbonFlows,
	}

	// Build success message with all filters
	message := "Ribbon flows retrieved successfully"
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

	return c.Status(fiber.StatusOK).JSON(utils.SuccessPaginatedResponse{
		Success: true,
		Message: message,
		Data:    response,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetRibbonFlow retrieves a single ribbon flow by tracking number
// @Summary Get a single ribbon flow
// @Description Retrieve a single ribbon flow by tracking number
// @Tags Ribbons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trackingNumber path string true "Tracking number of the ribbon flow"
// @Success 200 {object} utils.SuccessResponse{data=RibbonFlowResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/ribbons/flows/{trackingNumber} [get]
func (rfc *RibbonFlowController) GetRibbonFlow(c fiber.Ctx) error {
	trackingNumber := c.Params("trackingNumber")

	if trackingNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Tracking number is required",
		})
	}

	flow := rfc.BuildRibbonFlow(trackingNumber)

	// Check if qc-ribbon exists
	if flow.QCRibbon == nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "No QC Ribbon found with the provided tracking number",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Ribbon flow retrieved successfully",
		Data:    flow,
	})
}
