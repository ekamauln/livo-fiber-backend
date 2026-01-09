package models

import "time"

type Store struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StoreCode string    `gorm:"uniqueIndex;not null;type:varchar(50)" json:"store_code"`
	StoreName string    `gorm:"not null;type:varchar(100)" json:"store_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StoreResponse represents the store data returned in API responses
type StoreResponse struct {
	ID        uint   `json:"id"`
	StoreCode string `json:"storeCode"`
	StoreName string `json:"storeName"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ToResponse converts a Store model to a StoreResponse
func (s *Store) ToResponse() *StoreResponse {
	return &StoreResponse{
		ID:        s.ID,
		StoreCode: s.StoreCode,
		StoreName: s.StoreName,
		CreatedAt: s.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: s.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}
