package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type QCPerformance struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	UserID       uint            `gorm:"not null" json:"user_id"`
	SessionID    *uuid.UUID      `gorm:"type:uuid;default:null" json:"session_id"`
	Role         string          `gorm:"not null" json:"role"`
	LoginTime    time.Time       `gorm:"not null" json:"login_time"`
	LogoutTime   *time.Time      `gorm:"default:null" json:"logout_time"`
	TotalTime    time.Duration   `json:"total_time"`
	TotalQC      int             `json:"total_qc"`
	AverageScore decimal.Decimal `json:"average_score"`
	CreatedAt    time.Time       `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"not null" json:"updated_at"`

	User    User                  `gorm:"foreignKey:UserID" json:"user"`
	Session *Session              `gorm:"foreignKey:SessionID;constraint:OnUpdate:SET NULL,OnDelete:SET NULL;" json:"-"`
	Details []QCPerformanceDetail `gorm:"foreignKey:QCPerformanceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"details,omitempty"`
}

type QCPerformanceResponse struct {
	ID           uint                          `json:"id"`
	UserID       uint                          `json:"user_id"`
	User         *UserResponse                 `json:"user,omitempty"`
	SessionID    *uuid.UUID                    `json:"session_id,omitempty"`
	Role         string                        `json:"role"`
	LoginTime    string                        `json:"login_time"`
	LogoutTime   *string                       `json:"logout_time,omitempty"`
	TotalTime    string                        `json:"total_time_minutes"`
	TotalQC      int                           `json:"total_qc"`
	AverageScore string                        `json:"average_score"`
	Details      []QCPerformanceDetailResponse `json:"details,omitempty"`
}

func (qcp *QCPerformance) ToResponse() *QCPerformanceResponse {
	response := &QCPerformanceResponse{
		ID:           qcp.ID,
		UserID:       qcp.UserID,
		SessionID:    qcp.SessionID,
		Role:         qcp.Role,
		LoginTime:    qcp.LoginTime.Format("02-01-2006 15:04:05"),
		TotalTime:    fmt.Sprintf("%02d:%02d:%02d", int(qcp.TotalTime.Hours()), int(qcp.TotalTime.Minutes())%60, int(qcp.TotalTime.Seconds())%60),
		TotalQC:      qcp.TotalQC,
		AverageScore: qcp.AverageScore.String(),
	}

	if qcp.LogoutTime != nil {
		logoutTimeStr := qcp.LogoutTime.Format("02-01-2006 15:04:05")
		response.LogoutTime = &logoutTimeStr
	}

	if qcp.User.ID != 0 {
		response.User = qcp.User.ToResponse()
	}

	if len(qcp.Details) > 0 {
		detailsResponse := make([]QCPerformanceDetailResponse, len(qcp.Details))
		for i, detail := range qcp.Details {
			detailsResponse[i] = *detail.ToResponse()
		}
		response.Details = detailsResponse
	}

	return response
}
