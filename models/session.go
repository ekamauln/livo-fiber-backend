package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	RefreshToken string    `gorm:"not null;uniqueIndex;type:text" json:"-"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	IPAddress    string    `gorm:"type:varchar(50)" json:"ip_address"`
	DeviceType   string    `gorm:"type:varchar(20)" json:"device_type"` // web or mobile
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// SessionResponse represents the session data returned in API responses
type SessionResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uint      `json:"userId"`
	UserAgent  string    `json:"userAgent"`
	IPAddress  string    `json:"ipAddress"`
	DeviceType string    `json:"deviceType"`
	ExpiresAt  string    `json:"expiresAt"`
	CreatedAt  string    `json:"createdAt"`
}

// ToResponse converts a Session model to a SessionResponse
func (s *Session) ToResponse() *SessionResponse {
	return &SessionResponse{
		ID:         s.ID,
		UserID:     s.UserID,
		UserAgent:  s.UserAgent,
		IPAddress:  s.IPAddress,
		DeviceType: s.DeviceType,
		ExpiresAt:  s.ExpiresAt.Format("02-01-2006 15:04:05"),
		CreatedAt:  s.CreatedAt.Format("02-01-2006 15:04:05"),
	}
}
