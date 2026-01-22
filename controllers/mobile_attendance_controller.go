package controllers

import (
	"fmt"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type MobileAttendanceController struct {
	DB *gorm.DB
}

func NewMobileAttendanceController(db *gorm.DB) *MobileAttendanceController {
	return &MobileAttendanceController{DB: db}
}

// Unique response structs
// Unique response structs
type MobileCheckInResponse struct {
	Matched    bool                 `json:"matched" example:"true"`
	UserID     string               `json:"userId" example:"1"`
	Confidence float64              `json:"confidence" example:"0.95"`
	User       *models.UserResponse `json:"user"`
	Attendance *models.Attendance   `json:"attendance"`
	Status     string               `json:"status" example:"fullday"`
	Late       int                  `json:"late" example:"2"`
}

type MobileCheckOutResponse struct {
	Matched    bool                 `json:"matched" example:"true"`
	UserID     string               `json:"userId" example:"1"`
	Confidence float64              `json:"confidence" example:"0.95"`
	User       *models.UserResponse `json:"user"`
	Attendance *models.Attendance   `json:"attendance"`
	Status     string               `json:"status" example:"halfday"`
	Overtime   int                  `json:"overtime" example:"30"`
}

// VerifyUserFace verifies a user's face
// @Summary Verify User Face
// @Description Verify the logged-in user's face against their registered face
// @Tags Mobile Attendances
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param image formData file true "Face image to verify"
// @Success 200 {object} utils.SuccessResponse{data=utils.VerifyResult}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/mobile-attendances/face-verify [post]
func (mac *MobileAttendanceController) VerifyUserFace(c fiber.Ctx) error {
	// Get current user ID from context
	currUserID := c.Locals("userId").(string)

	// Get user from database
	var user models.User
	if err := mac.DB.Where("id = ?", currUserID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	}

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

	tmpPath := fmt.Sprintf("tmp/verify_%d.jpg", user.ID)
	if err := c.SaveFile(file, tmpPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to save image file",
		})
	}
	defer os.Remove(tmpPath)

	result, err := utils.SendToDeepFaceVerify(user.ID, tmpPath)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Face verification failed: %v", err),
		})
	}

	if !result.Matched {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success":    false,
			"error":      "Face verification failed - face does not match",
			"matched":    result.Matched,
			"userId":     result.UserID,
			"confidence": result.Confidence,
		})
	}

	// Attendance logging can be implemented here
	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "Face verified successfully",
		Data:    result,
	})
}
