package models

import "time"

type QCPerformanceDetail struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	QCPerformanceID uint           `gorm:"not null" json:"qc_performance_id"`
	TrackingNumber  string         `gorm:"not null;type:varchar(100)" json:"tracking_number"`
	Type            string         `gorm:"not null;type:varchar(50)" json:"type"` // "ribbon" or "online"
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	QCPerformance   *QCPerformance `gorm:"foreignKey:QCPerformanceID" json:"-"`
}

type QCPerformanceDetailResponse struct {
	ID             uint   `json:"id"`
	TrackingNumber string `json:"trackingNumber"`
	Type           string `json:"type"`
	CreatedAt      string `json:"createdAt"`
}

func (qcpd *QCPerformanceDetail) ToResponse() *QCPerformanceDetailResponse {
	return &QCPerformanceDetailResponse{
		ID:             qcpd.ID,
		TrackingNumber: qcpd.TrackingNumber,
		Type:           qcpd.Type,
		CreatedAt:      qcpd.CreatedAt.Format("02-01-2006 15:04:05"),
	}
}
