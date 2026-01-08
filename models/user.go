package models

import (
	"time"
)

type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"uniqueIndex;not null;type:varchar(50)" json:"username"`
	Password  string     `gorm:"not null" json:"-"`
	FullName  string     `gorm:"not null;type:varchar(100)" json:"full_name"`
	Email     string     `gorm:"uniqueIndex;not null;type:varchar(100)" json:"email"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	LastLogin *time.Time `gorm:"default:null" json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	Roles    []Role    `gorm:"many2many:user_roles;" json:"roles"`
	Sessions []Session `gorm:"foreignKey:UserID" json:"-"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username"`
	FullName  string   `json:"fullName"`
	Email     string   `json:"email"`
	IsActive  bool     `json:"isActive"`
	LastLogin *string  `json:"lastLogin,omitempty"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	Roles     []string `json:"roles"`
}

type UserRole struct {
	UserID uint `gorm:"not null" json:"user_id"`
	RoleID uint `gorm:"not null" json:"role_id"`

	User User `gorm:"foreignKey:UserID" json:"-"`
	Role Role `gorm:"foreignKey:RoleID" json:"-"`
}

// ToResponse converts a User model to a UserResponse
func (u *User) ToResponse() *UserResponse {
	// Extract role names
	roleNames := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		roleNames[i] = role.Name
	}

	// Extract session IDs
	sessionIDs := make([]string, len(u.Sessions))
	for i, session := range u.Sessions {
		sessionIDs[i] = session.ID.String()
	}

	// Format LastLogin
	var lastLoginStr *string
	if u.LastLogin != nil {
		formatted := u.LastLogin.Format("02-01-2006 15:04:05")
		lastLoginStr = &formatted
	}

	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		FullName:  u.FullName,
		Email:     u.Email,
		IsActive:  u.IsActive,
		LastLogin: lastLoginStr,
		CreatedAt: u.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: u.UpdatedAt.Format("02-01-2006 15:04:05"),
		Roles:     roleNames,
	}
}
