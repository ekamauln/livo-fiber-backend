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

type AttendanceController struct {
	DB *gorm.DB
}

func NewAttendanceController(db *gorm.DB) *AttendanceController {
	return &AttendanceController{DB: db}

}

// SearchUsersByFace searches for users by face image
// @Summary Search Users by Face
// @Description Search for users by face image
// @Tags Attendances
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
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
