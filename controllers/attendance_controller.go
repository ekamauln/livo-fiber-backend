package controllers

import (
	"fmt"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type AttendanceController struct {
	DB *gorm.DB
}

func NewAttendanceController(db *gorm.DB) *AttendanceController {
	return &AttendanceController{DB: db}

}

// Unique response structs
type CheckInResponse struct {
	Matched    bool                 `json:"matched" example:"true"`
	UserID     string               `json:"userId" example:"1"`
	Confidence float64              `json:"confidence" example:"0.95"`
	User       *models.UserResponse `json:"user"`
	Attendance *models.Attendance   `json:"attendance"`
	Status     string               `json:"status" example:"fullday"`
	Late       int                  `json:"late" example:"2"`
}

type CheckOutResponse struct {
	Matched    bool                 `json:"matched" example:"true"`
	UserID     string               `json:"userId" example:"1"`
	Confidence float64              `json:"confidence" example:"0.95"`
	User       *models.UserResponse `json:"user"`
	Attendance *models.Attendance   `json:"attendance"`
	Status     string               `json:"status" example:"halfday"`
	Overtime   int                  `json:"overtime" example:"30"`
}

// SearchUsersByFace searches for users by face image
// @Summary Search Users by Face
// @Description Search for users by face image
// @Tags Attendances
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Face image to search for"
// @Success 200 {object} utils.SuccessResponse{data=models.UserResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/attendances/search/face [post]
func (ac *AttendanceController) SearchUsersByFace(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Image file is required",
		})
	}

	// Validate mime type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid image file type",
		})
	}

	tmpPath := "tmp/search_face.jpg"
	if err := c.SaveFile(file, tmpPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to save image file",
		})
	}
	defer os.Remove(tmpPath)

	result, err := utils.SendToDeepFaceSearch(tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Face search failed: %v", err),
		})
	}

	if !result.Matched {
		return c.JSON(fiber.Map{
			"matched": false,
		})
	}

	// Fetch user data from database
	var user models.User
	if err := ac.DB.Preload("Roles").Where("id = ?", result.UserID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"matched":    true,
		"userId":     result.UserID,
		"confidence": result.Confidence,
		"user":       user.ToResponse(),
	})
}

// CheckInUserByFace checks in a user by face image
// @Summary Check In Users by Face
// @Description Check In for users by face image
// @Tags Attendances
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Face image to search for"
// @Success 200 {object} utils.SuccessResponse{data=CheckInResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/attendances/checkin/face [post]
func (ac *AttendanceController) CheckInUserByFace(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Image file is required",
		})
	}

	// Validate mime type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid image file type",
		})
	}

	tmpPath := "tmp/search_face.jpg"
	if err := c.SaveFile(file, tmpPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to save image file",
		})
	}
	defer os.Remove(tmpPath)

	result, err := utils.SendToDeepFaceSearch(tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Face search failed: %v", err),
		})
	}

	if !result.Matched {
		return c.JSON(fiber.Map{
			"matched": false,
		})
	}

	// Fetch user data from database
	var user models.User
	if err := ac.DB.Preload("Roles").Where("id = ?", result.UserID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	}

	// Check if user already checked in today
	var attendance models.Attendance
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := ac.DB.Where("user_id = ? AND checked_in >= ? AND checked_in < ? AND checked = ?", user.ID, startOfDay, endOfDay, true).First(&attendance).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User already checked in today",
		})
	}

	// Automatically determine status based on check-in time
	checkedInTime := time.Now()

	// Define time windows for fullday and halfday
	fulldayStart := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
	fulldayEnd := fulldayStart.Add(5 * time.Minute)

	halfdayStart := time.Date(now.Year(), now.Month(), now.Day(), 12, 30, 0, 0, now.Location())
	halfdayEnd := halfdayStart.Add(5 * time.Minute)

	var status string
	var workStartTime time.Time
	var lateMinutes int

	// Check which time window the check-in falls into
	if checkedInTime.After(fulldayStart.Add(-1*time.Minute)) && checkedInTime.Before(fulldayEnd.Add(1*time.Minute)) {
		// Within fullday window
		status = "fullday"
		workStartTime = fulldayStart

		if checkedInTime.After(fulldayEnd) {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   fmt.Sprintf("Check-in time has expired for fullday shift. Deadline was %s", fulldayEnd.Format("15:04")),
			})
		}

		if checkedInTime.After(workStartTime) {
			lateMinutes = int(checkedInTime.Sub(workStartTime).Minutes())
		}
	} else if checkedInTime.After(halfdayStart.Add(-1*time.Minute)) && checkedInTime.Before(halfdayEnd.Add(1*time.Minute)) {
		// Within halfday window
		status = "halfday"
		workStartTime = halfdayStart

		if checkedInTime.After(halfdayEnd) {
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   fmt.Sprintf("Check-in time has expired for halfday shift. Deadline was %s", halfdayEnd.Format("15:04")),
			})
		}

		if checkedInTime.After(workStartTime) {
			lateMinutes = int(checkedInTime.Sub(workStartTime).Minutes())
		}
	} else {
		// Not within any valid check-in window
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error: fmt.Sprintf("Not within valid check-in time. Fullday: %s-%s, Halfday: %s-%s",
				fulldayStart.Format("15:04"), fulldayEnd.Format("15:04"),
				halfdayStart.Format("15:04"), halfdayEnd.Format("15:04")),
		})
	}

	// Create attendance record
	newAttendance := models.Attendance{
		UserID:    user.ID,
		CheckedIn: checkedInTime,
		Checked:   true,
		Status:    status,
		Late:      lateMinutes,
	}

	if err := ac.DB.Create(&newAttendance).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create attendance record",
		})
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "User checked in successfully",
		Data: CheckInResponse{
			Matched:    true,
			UserID:     result.UserID,
			Confidence: result.Confidence,
			User:       user.ToResponse(),
			Attendance: &newAttendance,
			Status:     status,
			Late:       lateMinutes,
		},
	})
}

// CheckOutUserByFace checks out a user by face image
// @Summary Check Out Users by Face
// @Description Check Out for users by face image
// @Tags Attendances
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Face image to search for"
// @Success 200 {object} utils.SuccessResponse{data=CheckOutResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/attendances/checkout/face [put]
func (ac *AttendanceController) CheckOutUserByFace(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Image file is required",
		})
	}

	// Validate mime type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid image file type",
		})
	}

	tmpPath := "tmp/search_face.jpg"
	if err := c.SaveFile(file, tmpPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to save image file",
		})
	}
	defer os.Remove(tmpPath)

	result, err := utils.SendToDeepFaceSearch(tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Face search failed: %v", err),
		})
	}

	if !result.Matched {
		return c.JSON(fiber.Map{
			"matched": false,
		})
	}

	// Fetch user data from database
	var user models.User
	if err := ac.DB.Preload("Roles").Where("id = ?", result.UserID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	}

	// Search the target user's attendance record
	var attendance models.Attendance
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	if err := ac.DB.Where("user_id = ? AND checked_in >= ? AND checked_in < ? AND checked = ?", user.ID, startOfDay, endOfDay, true).First(&attendance).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Attendance record not found or user has not checked in today",
		})
	}

	// Automatically determine checkout behavior based on time
	checkedOutTime := time.Now()

	// Define checkout time windows
	earlyCheckOut := time.Date(now.Year(), now.Month(), now.Day(), 12, 30, 0, 0, now.Location())
	earlyCheckOutEnd := earlyCheckOut.Add(5 * time.Minute)

	regularCheckOut := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location())
	regularCheckOutStart := regularCheckOut.Add(-5 * time.Minute) // Allow 5 minutes before

	overtime := 0

	// Check if checking out around 12:30 (early checkout)
	if checkedOutTime.After(earlyCheckOut.Add(-1*time.Minute)) && checkedOutTime.Before(earlyCheckOutEnd.Add(1*time.Minute)) {
		// Update status from fullday to halfday, no overtime
		attendance.Status = "halfday"
		attendance.CheckedOut = &checkedOutTime
		attendance.Checked = false
		attendance.Overtime = 0
	} else if checkedOutTime.After(regularCheckOutStart) {
		// Checking out around 17:00 or later
		switch attendance.Status {
		case "halfday":
			// Halfday status: just update checkout time, no overtime
			attendance.CheckedOut = &checkedOutTime
			attendance.Checked = false
			attendance.Overtime = 0
		case "fullday":
			// Fullday status: update checkout and calculate overtime if after 17:00
			attendance.CheckedOut = &checkedOutTime
			attendance.Checked = false

			if checkedOutTime.After(regularCheckOut) {
				overtime = int(checkedOutTime.Sub(regularCheckOut).Minutes())
			}
			attendance.Overtime = overtime
		}
	} else {
		// Not within valid checkout windows
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error: fmt.Sprintf("Not within valid check-out time. Early checkout: %s-%s, Regular checkout: %s onwards",
				earlyCheckOut.Format("15:04"), earlyCheckOutEnd.Format("15:04"),
				regularCheckOutStart.Format("15:04")),
		})
	}

	// Update attendance record
	if err := ac.DB.Save(&attendance).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update attendance record",
		})
	}

	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "User checked out successfully",
		Data: CheckOutResponse{
			Matched:    true,
			UserID:     result.UserID,
			Confidence: result.Confidence,
			User:       user.ToResponse(),
			Attendance: &attendance,
			Status:     attendance.Status,
			Overtime:   overtime,
		},
	})
}
