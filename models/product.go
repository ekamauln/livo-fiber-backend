package models

import "time"

type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SKU       string    `gorm:"uniqueIndex;not null;type:varchar(255)" json:"sku"`
	Name      string    `gorm:"not null;type:varchar(255)" json:"name"`
	Image     string    `gorm:"type:text" json:"image"`
	Variant   string    `gorm:"type:varchar(100)" json:"variant"`
	Location  string    `gorm:"type:varchar(100)" json:"location"`
	NeedCheck bool      `gorm:"default:false" json:"need_check"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductResponse represents the product data returned in API responses
type ProductResponse struct {
	ID        uint   `json:"id"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Variant   string `json:"variant"`
	NeedCheck bool   `json:"needCheck"`
	Location  string `json:"location"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ToResponse converts a Product model to a ProductResponse
func (p *Product) ToResponse() *ProductResponse {
	return &ProductResponse{
		ID:        p.ID,
		SKU:       p.SKU,
		Name:      p.Name,
		Image:     p.Image,
		Variant:   p.Variant,
		Location:  p.Location,
		NeedCheck: p.NeedCheck,
		CreatedAt: p.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: p.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}
