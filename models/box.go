package models

import "time"

type Box struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BoxCode   string    `gorm:"uniqueIndex;not null;type:varchar(100)" json:"box_code"`
	BoxName   string    `gorm:"not null;type:varchar(100)" json:"box_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BoxResponse represents the box data returned in API responses
type BoxResponse struct {
	ID        uint   `json:"id"`
	BoxCode   string `json:"boxCode"`
	BoxName   string `json:"boxName"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ToResponse converts a Box model to a BoxResponse
func (b *Box) ToResponse() *BoxResponse {
	return &BoxResponse{
		ID:        b.ID,
		BoxCode:   b.BoxCode,
		BoxName:   b.BoxName,
		CreatedAt: b.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: b.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}
