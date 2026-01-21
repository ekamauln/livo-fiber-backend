package models

import "time"

type Attendance struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `json:"user_id"`
	Status     string     `gorm:"type:varchar(20);not null" json:"status"`
	Late       int        `gorm:"type:int;default:0" json:"late"`     // in minutes
	Overtime   int        `gorm:"type:int;default:0" json:"overtime"` // in minutes
	CheckedIn  time.Time  `json:"checked_in"`
	CheckedOut *time.Time `gorm:"default:null" json:"checked_out"`
	Checked    bool       `gorm:"default:true" json:"checked"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

// AttendanceResponse represents the attendance data returned in API responses
type AttendanceResponse struct {
	ID         uint   `json:"id"`
	User       string `json:"user"`
	Status     string `json:"status"`
	Late       int    `json:"late"`
	Overtime   int    `json:"overtime"`
	CheckedIn  string `json:"checkedIn"`
	CheckedOut string `json:"checkedOut"`
	Checked    bool   `json:"checked"`
}

// ToResponse converts an Attendance model to an AttendanceResponse
func (a *Attendance) ToResponse() *AttendanceResponse {
	// Visual user handler
	var userName string
	if a.User.ID != 0 {
		userName = a.User.FullName
	} else {
		userName = "Unknown User"
	}

	return &AttendanceResponse{
		ID:         a.ID,
		User:       userName,
		Status:     a.Status,
		Late:       a.Late,
		Overtime:   a.Overtime,
		CheckedIn:  a.CheckedIn.Format("02-01-2006 15:04:05"),
		CheckedOut: a.CheckedOut.Format("02-01-2006 15:04:05"),
		Checked:    a.Checked,
	}
}
