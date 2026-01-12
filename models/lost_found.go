package models

import "time"

type LostFound struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProductSKU string    `gorm:"not null;index;type:varchar(255)" json:"product_sku"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	Reason     string    `gorm:"not null;type:text" json:"reason"`
	CreatedBy  uint      `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	CreateUser *User    `gorm:"foreignKey:CreatedBy" json:"create_user,omitempty"`
	Product    *Product `gorm:"-" json:"product,omitempty"`
}

type LostFoundresponse struct {
	ID         uint             `json:"id"`
	ProductSKU string           `json:"productSKU"`
	Quantity   int              `json:"quantity"`
	Reason     string           `json:"reason"`
	CreatedBy  string           `json:"createdBy"`
	CreatedAt  string           `json:"createdAt"`
	UpdatedAt  string           `json:"updatedAt"`
	Product    *ProductResponse `json:"product,omitempty"`
}

func (lf *LostFound) ToResponse() LostFoundresponse {
	// user visual handler
	var createdBy string
	if lf.CreateUser != nil {
		createdBy = lf.CreateUser.FullName
	}

	// Include product details if productSKU is exists in SKU Product
	var productResp *ProductResponse
	if lf.Product != nil {
		productResp = lf.Product.ToResponse()
	}

	return LostFoundresponse{
		ID:         lf.ID,
		ProductSKU: lf.ProductSKU,
		Quantity:   lf.Quantity,
		Reason:     lf.Reason,
		CreatedBy:  createdBy,
		CreatedAt:  lf.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt:  lf.UpdatedAt.Format("02-01-2006 15:04:05"),
		Product:    productResp,
	}
}
