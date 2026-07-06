package controllers

import (
	"fmt"
	"livo-fiber-backend/config"
	"livo-fiber-backend/database"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AuthController struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewAuthController(cfg *config.Config, db *gorm.DB) *AuthController {
	return &AuthController{Config: cfg, DB: db}
}

// Request structs
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50" example:"john_doe"`
	Password string `json:"password" validate:"required,min=8" example:"SecurePass123"`
	FullName string `json:"fullName" validate:"required,min=2,max=100" example:"John Doe"`
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"john_doe"`
	Password string `json:"password" validate:"required" example:"SecurePass123"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required" example:"v4.local.xxx"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with username, password, full name, and email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} utils.SuccessResponse{data=models.UserResponse} "User registered successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 409 {object} utils.ErrorResponse "Username or email already exists"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /api/auth/register [post]
func (ac *AuthController) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Println("Invalid request body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Isi permintaan tidak valid",
		})
	}

	// Check if username exists
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		log.Println("Username already exists:", req.Username)
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Username sudah terdaftar",
		})
	}

	// Check if email exists
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		log.Println("Email already exists:", req.Email)
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Email sudah terdaftar",
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Println("Failed to hash password:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Gagal melakukan hash kata sandi",
		})
	}

	// Get guest role first to ensure it exists
	var guestRole models.Role
	if err := database.DB.Where("role_name = ?", "guest").First(&guestRole).Error; err != nil {
		log.Println("Failed to get default role:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Gagal mengambil peran default",
		})
	}

	// Use transaction to ensure user and role assignment are created together
	var user models.User
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Create user
		user = models.User{
			Username: req.Username,
			Password: hashedPassword,
			FullName: req.FullName,
			Email:    req.Email,
			IsActive: true,
		}

		if err := tx.Create(&user).Error; err != nil {
			log.Panicln(err)
			return err
		}

		// Associate role with user
		if err := tx.Table("user_roles").Create(map[string]interface{}{
			"user_id": user.ID,
			"role_id": guestRole.ID,
		}).Error; err != nil {
			log.Panicln(err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Println("Failed to create user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Gagal membuat pengguna",
		})
	}

	// Load roles for response
	database.DB.Preload("Roles").First(&user, "id = ?", user.ID)

	log.Println("User registered successfully:", req.Username)
	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Pengguna berhasil terdaftar",
		Data:    user.ToResponse(),
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return access token with optional refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} utils.LoginResponse "Login successful"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 401 {object} utils.ErrorResponse "Invalid credentials"
// @Failure 403 {object} utils.ErrorResponse "User account is disabled"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /api/auth/login [post]
func (ac *AuthController) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Println("Invalid request body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Isi permintaan tidak valid",
		})
	}

	// Find user
	var user models.User
	if err := database.DB.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
		log.Println("Invalid credentials for user:", req.Username)
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Username/Katasandi tidak valid",
		})
	}

	// Check if user is active
	if !user.IsActive {
		log.Println("User account is disabled:", req.Username)
		return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Akun pengguna tidak aktif",
		})
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		log.Println("Invalid credentials for user:", req.Username)
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Username/katasandi tidak valid",
		})
	}

	// Check if user already has an active session
	var activeSession models.Session
	if err := database.DB.Where("user_id = ? AND expires_at > ?", user.ID, time.Now()).First(&activeSession).Error; err == nil {
		log.Println("User already logged in on another device:", req.Username)
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Akun sedang digunakan di perangkat lain. Silakan logout terlebih dahulu.",
		})
	} else if err != gorm.ErrRecordNotFound {
		log.Println("Failed to check active sessions:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Terjadi kesalahan pada server",
		})
	}

	// Get role names
	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.RoleName
	}

	// Generate tokens
	claims := utils.TokenClaims{
		UserID:   fmt.Sprintf("%d", user.ID),
		Username: user.Username,
		Roles:    roleNames,
	}

	accessToken, err := utils.GenerateAccessToken(claims, ac.Config)
	if err != nil {
		log.Println("Failed to generate access token:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Gagal membuat akses token",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(claims, ac.Config)
	if err != nil {
		log.Println("Failed to generate refresh token:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Gagam membuat refresh token",
		})
	}

	// Detect device type
	userAgent := c.Get("User-Agent")
	deviceType := "web"

	// Check custom header first (for mobile apps like React Native/Expo)
	customDeviceType := c.Get("X-Device-Type")
	if customDeviceType != "" {
		if customDeviceType == "mobile" || customDeviceType == "ios" || customDeviceType == "android" {
			deviceType = "mobile"
		}
	} else if strings.Contains(strings.ToLower(userAgent), "mobile") {
		// Fallback to User-Agent detection
		deviceType = "mobile"
	}

	// Create session
	session := models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    c.IP(),
		DeviceType:   deviceType,
		ExpiresAt:    time.Now().Add(time.Duration(ac.Config.RefreshTokenTTL) * 24 * time.Hour),
	}

	if err := database.DB.Create(&session).Error; err != nil {
		log.Println("Failed to create session:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Gagal membuat sesi",
		})
	}

	// Create QC Performance record if user has qc-ribbon or qc-online roles
	var qcRole string
	hasQCRibbon := false
	hasQCOnline := false
	for _, roleName := range roleNames {
		switch roleName {
		case "qc-ribbon":
			hasQCRibbon = true
		case "qc-online":
			hasQCOnline = true
		}
	}
	if hasQCRibbon && hasQCOnline {
		qcRole = "qc-ribbon, qc-online"
	} else if hasQCRibbon {
		qcRole = "qc-ribbon"
	} else if hasQCOnline {
		qcRole = "qc-online"
	}

	if qcRole != "" {
		qcPerformance := models.QCPerformance{
			UserID:    user.ID,
			SessionID: &session.ID,
			Role:      qcRole,
			LoginTime: time.Now(),
		}
		if err := database.DB.Create(&qcPerformance).Error; err != nil {
			log.Println("Failed to create QC Performance record:", err)
		}
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("last_login", now)

	userResponse := user.ToResponse()
	response := utils.LoginResponse{
		Success:     true,
		AccessToken: accessToken,
		User:        userResponse,
	}

	// If web app, set refresh token in httponly cookie
	if deviceType == "web" {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			MaxAge:   ac.Config.RefreshTokenTTL * 24 * 3600,
		})
	} else {
		// If mobile app, include refresh token in response
		response.RefreshToken = refreshToken
	}

	log.Println("User logged in successfully:", req.Username)
	return c.JSON(response)
}

// Logout handles user logout and clear current session
// @Summary User logout
// @Description Logout user and invalidate current session or all sessions
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body RefreshTokenRequest false "Optional refresh token to logout specific session"
// @Success 200 {object} utils.SuccessResponse "Logged out successfully"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /api/auth/logout [post]
func (ac *AuthController) Logout(c fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	// Get refresh token from cookie or body
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.Bind().Body(&body); err == nil {
			refreshToken = body.RefreshToken
		}
	}

	// Process QC Performance before deleting sessions
	var sessions []models.Session
	if refreshToken != "" {
		database.DB.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).Find(&sessions)
	} else {
		database.DB.Where("user_id = ?", userID).Find(&sessions)
	}

	for _, session := range sessions {
		var qcPerformance models.QCPerformance
		if err := database.DB.Where("session_id = ? AND logout_time IS NULL", session.ID).First(&qcPerformance).Error; err == nil {
			now := time.Now()
			totalDuration := now.Sub(qcPerformance.LoginTime)
			totalMinutes := totalDuration.Minutes()

			// Calculate TotalQC and Details
			var qcRibbons []models.QCRibbon
			database.DB.Where("qc_by = ? AND updated_at BETWEEN ? AND ?", qcPerformance.UserID, qcPerformance.LoginTime, now).Find(&qcRibbons)

			var qcOnlines []models.QCOnline
			database.DB.Where("qc_by = ? AND updated_at BETWEEN ? AND ?", qcPerformance.UserID, qcPerformance.LoginTime, now).Find(&qcOnlines)

			totalQC := len(qcRibbons) + len(qcOnlines)

			// Calculate AverageScore
			var avgScore float64
			if totalMinutes > 0 {
				avgScore = float64(totalQC) / totalMinutes
			}

			if avgScore >= 0.92 {
				avgScore = 100.0
			} else {
				// Proportional to 0.92 (e.g., 0.46 = 50)
				avgScore = (avgScore / 0.92) * 100.0
			}

			// Update QCPerformance
			qcPerformance.LogoutTime = &now
			qcPerformance.TotalTime = totalDuration
			qcPerformance.TotalQC = totalQC
			qcPerformance.AverageScore = decimal.NewFromFloat(avgScore)
			database.DB.Save(&qcPerformance)

			// Insert Details
			for _, ribbon := range qcRibbons {
				detail := models.QCPerformanceDetail{
					QCPerformanceID: qcPerformance.ID,
					TrackingNumber:  ribbon.TrackingNumber,
					Type:            "ribbon",
				}
				database.DB.Create(&detail)
			}
			for _, online := range qcOnlines {
				detail := models.QCPerformanceDetail{
					QCPerformanceID: qcPerformance.ID,
					TrackingNumber:  online.TrackingNumber,
					Type:            "online",
				}
				database.DB.Create(&detail)
			}
		}
	}

	if refreshToken != "" {
		// Delete specific session
		database.DB.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).Delete(&models.Session{})
	} else {
		// Delete all sessions for user
		database.DB.Where("user_id = ?", userID).Delete(&models.Session{})
	}

	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		MaxAge:   -1,
	})

	log.Println("User logged out successfully, userID:", userID)
	return c.JSON(utils.SuccessResponse{
		Success: true,
		Message: "Berhasil logout",
	})
}

// RefreshToken handles token refreshing and generating new access token
// @Summary Refresh access token
// @Description Generate a new access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest false "Refresh token (optional if using cookie)"
// @Success 200 {object} utils.LoginResponse "Token refreshed successfully"
// @Failure 400 {object} utils.ErrorResponse "Refresh token required"
// @Failure 401 {object} utils.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /api/auth/refresh [post]
func (ac *AuthController) RefreshToken(c fiber.Ctx) error {
	// Get refresh token from cookie or body
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.Bind().Body(&body); err != nil || body.RefreshToken == "" {
			log.Println("Refresh token required:", err)
			return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Refresh token diperlukan",
			})
		}
		refreshToken = body.RefreshToken
	}

	// Validate refresh token
	token, err := utils.ValidateToken(refreshToken, ac.Config)
	if err != nil {
		log.Println("Invalid or expired refresh token:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Refresh token tidak valid / kadaluwarsa",
		})
	}

	// Check token type
	tokenType, err := token.GetString("type")
	if err != nil || tokenType != "refresh" {
		log.Println("Invalid token type")
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Jenis token tidak valid",
		})
	}

	// Get session - check both current and old refresh token (for grace period)
	var session models.Session
	if err := database.DB.Preload("User.Roles").
		Where("refresh_token = ? OR old_refresh_token = ?", refreshToken, refreshToken).
		First(&session).Error; err != nil {
		log.Println("Session not found for refresh token:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Sesi tidak ditemukan",
		})
	}

	// Grace Period Check: If using old token, allow if rotated recently
	isGracePeriod := refreshToken == session.OldRefreshToken
	if isGracePeriod {
		if time.Since(session.RotatedAt) > 30*time.Second {
			log.Println("Old refresh token used after grace period for userID:", session.UserID)
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Refresh token lama sudah tidak berlaku",
			})
		}
		log.Println("Grace period hit for userID:", session.UserID)
	}

	// Check if session expired
	if time.Now().After(session.ExpiresAt) {
		database.DB.Delete(&session)
		log.Println("Session expired for userID:", session.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Sesi telah kadaluwarsa",
		})
	}

	// Get role names
	roleNames := make([]string, len(session.User.Roles))
	for i, role := range session.User.Roles {
		roleNames[i] = role.RoleName
	}

	// Generate new access token
	claims := utils.TokenClaims{
		UserID:   fmt.Sprintf("%d", session.UserID),
		Username: session.User.Username,
		Roles:    roleNames,
	}

	newAccessToken, err := utils.GenerateAccessToken(claims, ac.Config)
	if err != nil {
		log.Println("Failed to generate access token:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Gagal menghasilkan access token",
		})
	}

	var newRefreshToken string
	if isGracePeriod {
		// Use the ALREADY rotated token if within grace period
		newRefreshToken = session.RefreshToken
	} else {
		// Normal Rotation: generate NEW refresh token
		newRefreshToken, err = utils.GenerateRefreshToken(claims, ac.Config)
		if err != nil {
			log.Println("Failed to generate refresh token:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
				Success: false,
				Error:   "Gagal menghasilkan refresh token",
			})
		}

		// Update session for rotation
		session.OldRefreshToken = session.RefreshToken
		session.RefreshToken = newRefreshToken
		session.RotatedAt = time.Now()
		session.ExpiresAt = time.Now().Add(time.Duration(ac.Config.RefreshTokenTTL) * 24 * time.Hour)
		database.DB.Save(&session)
	}

	userResponse := session.User.ToResponse()
	response := utils.LoginResponse{
		Success:     true,
		AccessToken: newAccessToken,
		User:        userResponse,
	}

	// Update cookie for web or include in response for mobile
	if session.DeviceType == "web" {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			MaxAge:   ac.Config.RefreshTokenTTL * 24 * 3600,
		})
	} else {
		response.RefreshToken = newRefreshToken
	}

	log.Println("Token refreshed successfully for userID:", session.UserID)
	return c.JSON(response)
}
