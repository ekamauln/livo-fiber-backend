package models

import "time"

type QCOnline struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TrackingNumber string    `gorm:"uniqueIndex;not null;type:varchar(100)" json:"tracking_number"`
	QCBy           uint      `gorm:"not null" json:"qc_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Complained     bool      `gorm:"default:false" json:"complained"`

	QCOnlineDetails []QCOnlineDetail `gorm:"foreignKey:QCOnlineID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"qc_online_details,omitempty"`
	QCUser          *User            `gorm:"foreignKey:QCBy" json:"qc_user,omitempty"`
	Order           *Order           `gorm:"-" json:"order,omitempty"`
}

type QCOnlineDetail struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	QCOnlineID uint `gorm:"not null" json:"qc_online_id"`
	BoxID      uint `gorm:"not null" json:"box_id"`
	Quantity   uint `gorm:"not null" json:"quantity"`

	Box      *Box      `gorm:"foreignKey:BoxID" json:"box,omitempty"`
	QCOnline *QCOnline `gorm:"foreignKey:QCOnlineID" json:"-"`
}

// QCOnlineResponse represents the QC Online data returned in API responses
type QCOnlineResponse struct {
	ID              uint                     `json:"id"`
	TrackingNumber  string                   `json:"trackingNumber"`
	QCBy            string                   `json:"qcBy"`
	CreatedAt       string                   `json:"createdAt"`
	UpdatedAt       string                   `json:"updatedAt"`
	Complained      bool                     `json:"complained"`
	QCOnlineDetails []QCOnlineDetailResponse `json:"qcOnlineDetails,omitempty"`
	Order           *OrderResponse           `json:"order,omitempty"`
}

type QCOnlineDetailResponse struct {
	Box      string `json:"box,omitempty"`
	Quantity uint   `json:"quantity"`
}

// ToResponse converts a QCOnline model to a QCOnlineResponse
func (qcr *QCOnline) ToResponse() *QCOnlineResponse {
	// Convert QC Online Details
	details := make([]QCOnlineDetailResponse, len(qcr.QCOnlineDetails))
	for i, detail := range qcr.QCOnlineDetails {
		detailResp := QCOnlineDetailResponse{
			Box:      detail.Box.BoxName,
			Quantity: detail.Quantity,
		}
		details[i] = detailResp
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

	return &QCOnlineResponse{
		ID:              qcr.ID,
		TrackingNumber:  qcr.TrackingNumber,
		QCBy:            qcBy,
		CreatedAt:       qcr.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt:       qcr.UpdatedAt.Format("02-01-2006 15:04:05"),
		Complained:      qcr.Complained,
		QCOnlineDetails: details,
		Order:           orderResponse,
	}
}
