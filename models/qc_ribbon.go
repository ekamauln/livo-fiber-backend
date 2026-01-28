package models

import "time"

type QCRibbon struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TrackingNumber string    `gorm:"uniqueIndex;not null;type:varchar(100)" json:"tracking_number"`
	QCBy           uint      `gorm:"not null" json:"qc_by"`
	Status         string    `gorm:"default:'in_progress';type:varchar(50)" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Complained     bool      `gorm:"default:false" json:"complained"`

	QCRibbonDetails []QCRibbonDetail `gorm:"foreignKey:QCRibbonID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"qc_ribbon_details,omitempty"`
	QCUser          *User            `gorm:"foreignKey:QCBy" json:"qc_user,omitempty"`
	Order           *Order           `gorm:"-" json:"order,omitempty"`
}

type QCRibbonDetail struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	QCRibbonID uint `gorm:"not null" json:"qc_ribbon_id"`
	BoxID      uint `gorm:"not null" json:"box_id"`
	Quantity   int  `gorm:"not null" json:"quantity"`

	Box      *Box      `gorm:"foreignKey:BoxID" json:"box,omitempty"`
	QCRibbon *QCRibbon `gorm:"foreignKey:QCRibbonID" json:"-"`
}

// QCRibbonResponse represents the QC Ribbon data returned in API responses
type QCRibbonResponse struct {
	ID             uint                     `json:"id"`
	TrackingNumber string                   `json:"trackingNumber"`
	QCBy           string                   `json:"qcBy"`
	Status         string                   `json:"status"`
	CreatedAt      string                   `json:"createdAt"`
	UpdatedAt      string                   `json:"updatedAt"`
	Complained     bool                     `json:"complained"`
	Details        []QCRibbonDetailResponse `json:"details,omitempty"`
	Order          *OrderResponse           `json:"order,omitempty"`
}

type QCRibbonDetailResponse struct {
	Box      string `json:"box,omitempty"`
	Quantity int    `json:"quantity"`
}

// ToResponse converts a QCRibbon model to a QCRibbonResponse
func (qcr *QCRibbon) ToResponse() *QCRibbonResponse {
	// Convert QC Ribbon Details
	details := make([]QCRibbonDetailResponse, len(qcr.QCRibbonDetails))
	for i, detail := range qcr.QCRibbonDetails {
		boxName := ""
		if detail.Box != nil {
			boxName = detail.Box.BoxName
		}
		detailResp := QCRibbonDetailResponse{
			Box:      boxName,
			Quantity: detail.Quantity,
		}
		details[i] = detailResp
	}

	// Status visual handlers
	var status string
	switch qcr.Status {
	case "in_progress":
		status = "In Progress"
	case "completed":
		status = "Completed"
	case "cancelled":
		status = "Cancelled"
	case "pending":
		status = "Pending"
	}

	// User visual handlers
	var qcBy string
	if qcr.QCUser != nil {
		qcBy = qcr.QCUser.FullName
	}

	// Include Order response if tracking number exists in Order
	var orderResponse *OrderResponse
	if qcr.Order != nil {
		orderResponse = qcr.Order.ToOrderResponse()
	}

	return &QCRibbonResponse{
		ID:             qcr.ID,
		TrackingNumber: qcr.TrackingNumber,
		QCBy:           qcBy,
		Status:         status,
		CreatedAt:      qcr.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt:      qcr.UpdatedAt.Format("02-01-2006 15:04:05"),
		Complained:     qcr.Complained,
		Details:        details,
		Order:          orderResponse,
	}
}
