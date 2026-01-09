package models

import (
	"time"
)

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoleName  string    `gorm:"uniqueIndex;not null;type:varchar(50)" json:"role_name"`
	Hierarchy int       `gorm:"not null" json:"hierarchy"` // 1=highest, 99=lowest
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Users []User `gorm:"many2many:user_roles;" json:"-"`
}

// RoleResponse represents the role data returned in API responses
type RoleResponse struct {
	ID        uint   `json:"id"`
	RoleName  string `json:"roleName"`
	Hierarchy int    `json:"hierarchy"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ToResponse converts a Role model to a RoleResponse
func (r *Role) ToResponse() *RoleResponse {
	return &RoleResponse{
		ID:        r.ID,
		RoleName:  r.RoleName,
		Hierarchy: r.Hierarchy,
		CreatedAt: r.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: r.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}
