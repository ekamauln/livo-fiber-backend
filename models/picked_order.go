package models

import "time"

type PickedOrder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"not null;index" json:"order_id"`
	PickedBy  uint      `gorm:"not null;index" json:"picked_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	PickUser *User  `gorm:"foreignKey:PickedBy" json:"pick_user,omitempty"`
	Order    *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

type PickedOrderResponse struct {
	ID        uint           `json:"id"`
	PickedBy  string         `json:"pickedBy"`
	Order     *OrderResponse `json:"order,omitempty"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

// ToResponse converts a PickedOrder model to a PickedOrderResponse
func (po *PickedOrder) ToResponse() *PickedOrderResponse {
	// User visual handlers
	var pickedBy string
	if po.PickUser != nil {
		pickedBy = po.PickUser.FullName
	}

	// Order Visual handlers
	var orderResp *OrderResponse
	if po.Order != nil {
		orderResp = po.Order.ToOrderResponse()
	}

	return &PickedOrderResponse{
		ID:        po.ID,
		PickedBy:  pickedBy,
		Order:     orderResp,
		CreatedAt: po.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: po.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}
