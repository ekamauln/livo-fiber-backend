package models

import "time"

type Expedition struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ExpeditionCode  string    `gorm:"uniqueIndex;not null;type:varchar(50)" json:"expedition_code"`
	ExpeditionName  string    `gorm:"not null;type:varchar(100)" json:"expedition_name"`
	ExpeditionSlug  string    `gorm:"uniqueIndex;not null;type:varchar(100)" json:"expedition_slug"`
	ExpeditionColor string    `gorm:"not null;type:varchar(20)" json:"expedition_color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ExpeditionResponse represents the expedition data returned in API responses
type ExpeditionResponse struct {
	ID              uint   `json:"id"`
	ExpeditionCode  string `json:"expeditionCode"`
	ExpeditionName  string `json:"expeditionName"`
	ExpeditionSlug  string `json:"expeditionSlug"`
	ExpeditionColor string `json:"expeditionColor"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// ToResponse converts an Expedition model to an ExpeditionResponse
func (e *Expedition) ToResponse() *ExpeditionResponse {
	return &ExpeditionResponse{
		ID:              e.ID,
		ExpeditionCode:  e.ExpeditionCode,
		ExpeditionName:  e.ExpeditionName,
		ExpeditionSlug:  e.ExpeditionSlug,
		ExpeditionColor: e.ExpeditionColor,
		CreatedAt:       e.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt:       e.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}
