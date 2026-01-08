package utils

import "livo-fiber-backend/models"

// Pagination represents pagination details
type Pagination struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessPaginatedResponse represents a paginated success response
type SuccessPaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination Pagination  `json:"pagination"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// LoginResponse represents the response returned upon successful login
type LoginResponse struct {
	Success      bool                 `json:"success"`
	AccessToken  string               `json:"accessToken"`
	RefreshToken string               `json:"refreshToken,omitempty"`
	User         *models.UserResponse `json:"user"`
}
