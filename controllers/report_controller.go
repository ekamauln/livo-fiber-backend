package controllers

import (
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"

	"gorm.io/gorm"
)

type ReportController struct {
	DB *gorm.DB
}

func NewReportController(db *gorm.DB) *ReportController {
	return &ReportController{DB: db}
}

// Unique response structs
type BoxUsageDetail struct {
	TrackingNumber string `json:"trackingNumber"`
	OrderGineeID   string `json:"orderGineeId"`
	BoxName        string `json:"boxName"`
	Quantity       int    `json:"quantity"`
	QcBy           uint   `json:"qcBy"`
	Username       string `json:"username"`
	FullName       string `json:"fullName"`
	CreatedAt      string `json:"createdAt"`
	Source         string `json:"source"`
}

type BoxCountReport struct {
	BoxID       uint             `json:"boxId"`
	BoxCode     string           `json:"boxCode"`
	BoxName     string           `json:"boxName"`
	TotalCount  int              `json:"totalCount"`
	RibbonCount int              `json:"ribbonCount"`
	OnlineCount int              `json:"onlineCount"`
	Details     []BoxUsageDetail `json:"details" gorm:"-"`
}

type BoxCountReportsListResponse struct {
	Reports    []BoxCountReport `json:"reports"`
	Pagination utils.Pagination `json:"pagination"`
}

type OutboundReportsListResponse struct {
	Outbounds []models.OutboundResponse `json:"outbounds"`
	Total     int
}
