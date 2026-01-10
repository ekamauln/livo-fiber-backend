package models

import "time"

type Outbound struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TrackingNumber  string    `gorm:"uniqueIndex;not null;type:varchar(100)" json:"tracking_number"`
	OutboundBy      uint      `gorm:"not null" json:"outbound_by"`
	Expedition      string    `gorm:"type:varchar(100)" json:"expedition"`
	ExpeditionSlug  string    `gorm:"type:varchar(100)" json:"expedition_slug"`
	ExpeditionColor string    `gorm:"type:varchar(50)" json:"expedition_color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Complained      bool      `gorm:"default:false" json:"complained"`

	OutboundUser *User  `gorm:"foreignKey:OutboundBy" json:"outbound_user,omitempty"`
	Order        *Order `gorm:"-" json:"order,omitempty"`
}

// OutboundResponse represents the outbound data returned in API responses
type OutboundResponse struct {
	ID              uint           `json:"id"`
	TrackingNumber  string         `json:"trackingNumber"`
	OutboundBy      string         `json:"outboundBy"`
	Expedition      string         `json:"expedition"`
	ExpeditionSlug  string         `json:"expeditionSlug"`
	ExpeditionColor string         `json:"expeditionColor"`
	CreatedAt       string         `json:"createdAt"`
	UpdatedAt       string         `json:"updatedAt"`
	Complained      bool           `json:"complained"`
	Order           *OrderResponse `json:"order,omitempty"`
}

// ToResponse converts an Outbound model to an OutboundResponse
func (o *Outbound) ToResponse() *OutboundResponse {
	// User visual handlers
	var outboundBy string
	if o.OutboundUser != nil {
		outboundBy = o.OutboundUser.FullName
	}

	// Include order response if tracking number exists in Order
	var orderResponse *OrderResponse
	if o.Order != nil {
		orderResponse = o.Order.ToOrderResponse()
	}

	return &OutboundResponse{
		ID:              o.ID,
		TrackingNumber:  o.TrackingNumber,
		OutboundBy:      outboundBy,
		Expedition:      o.Expedition,
		ExpeditionSlug:  o.ExpeditionSlug,
		ExpeditionColor: o.ExpeditionColor,
		CreatedAt:       o.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt:       o.UpdatedAt.Format("02-01-2006 15:04:05"),
		Complained:      o.Complained,
		Order:           orderResponse,
	}
}
