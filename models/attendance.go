package models

import (
	"strconv"
	"time"
)

type Location struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Attendance struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `json:"user_id"`
	Status     string     `gorm:"type:varchar(20);not null" json:"status"`
	Late       int        `gorm:"type:int;default:0" json:"late"`     // in minutes
	Overtime   int        `gorm:"type:int;default:0" json:"overtime"` // in minutes
	LocationID uint       `json:"location_id"`
	Latitude   float64    `json:"latitude"`
	Longitude  float64    `json:"longitude"`
	Accuracy   float64    `gorm:"default:0" json:"accuracy"` // in meters
	CheckedIn  time.Time  `json:"checked_in"`
	CheckedOut *time.Time `gorm:"default:null" json:"checked_out"`
	Checked    bool       `gorm:"default:true" json:"checked"`

	Location Location `gorm:"foreignKey:LocationID" json:"location"`
	User     User     `gorm:"foreignKey:UserID" json:"user"`
}

// LocationResponse represents the location data returned in API responses
type LocationResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// ToResponse converts a Location model to a LocationResponse
func (l *Location) ToResponse() *LocationResponse {
	return &LocationResponse{
		ID:        l.ID,
		Name:      l.Name,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		CreatedAt: l.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt: l.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}

// AttendanceResponse represents the attendance data returned in API responses
type AttendanceResponse struct {
	ID         uint   `json:"id"`
	User       string `json:"user"`
	Status     string `json:"status"`
	Location   string `json:"location"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
	Accuracy   string `json:"accuracy"`
	Late       int    `json:"late"`
	Overtime   int    `json:"overtime"`
	CheckedIn  string `json:"checkedIn"`
	CheckedOut string `json:"checkedOut"`
	Checked    bool   `json:"checked"`
}

// ToResponse converts an Attendance model to an AttendanceResponse
func (a *Attendance) ToResponse() *AttendanceResponse {
	// Visual location handler
	var locationName string
	if a.Location.ID != 0 {
		locationName = a.Location.Name
	} else {
		locationName = "Unknown Location"
	}

	// Visual user handler
	var userName string
	if a.User.ID != 0 {
		userName = a.User.FullName
	} else {
		userName = "Unknown User"
	}

	// Latitude and Longitude formatting
	latitudeStr := strconv.FormatFloat(a.Latitude, 'f', 6, 64)
	longitudeStr := strconv.FormatFloat(a.Longitude, 'f', 6, 64)
	accuracyStr := strconv.FormatFloat(a.Accuracy, 'f', 2, 64)

	return &AttendanceResponse{
		ID:         a.ID,
		User:       userName,
		Status:     a.Status,
		Location:   locationName,
		Latitude:   latitudeStr,
		Longitude:  longitudeStr,
		Accuracy:   accuracyStr,
		Late:       a.Late,
		Overtime:   a.Overtime,
		CheckedIn:  a.CheckedIn.Format("02-01-2006 15:04:05"),
		CheckedOut: a.CheckedOut.Format("02-01-2006 15:04:05"),
		Checked:    a.Checked,
	}
}
