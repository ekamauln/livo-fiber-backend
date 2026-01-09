package controllers

import (
	"fmt"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type ProductController struct {
	DB *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{DB: db}
}

// Request structs
type CreateProductRequest struct {
	SKU      string `json:"sku" validate:"required,min=3,max=50"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Image    string `json:"image" validate:"omitempty"`
	Variant  string `json:"variant" validate:"omitempty,min=1,max=100"`
	Location string `json:"location" validate:"omitempty,min=1,max=100"`
}

type UpdateProductRequest struct {
	SKU      string `json:"sku" validate:"required,min=3,max=50"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Image    string `json:"image" validate:"omitempty"`
	Variant  string `json:"variant" validate:"omitempty,min=1,max=100"`
	Location string `json:"location" validate:"omitempty,min=1,max=100"`
}

// GetProducts retrieves a list of products with pagination and search
// @Summary Get Products
// @Description Retrieve a list of products with pagination and search
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of Products per page" default(10)
// @Param search query string false "Search term for product SKU or name"
// @Success 200 {object} utils.SuccessPaginatedResponse{data=[]models.Product}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/products [get]
func (pc *ProductController) GetProducts(c fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var products []models.Product

	// Build base query
	query := pc.DB.Model(&models.Product{}).Order("created_at DESC")

	// Search condition if provided
	search := strings.TrimSpace(c.Query("search", ""))
	if search != "" {
		query = query.Where("sku ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Retrieve paginated results
	if err := query.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve products",
		})
	}

	// Format response
	productList := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productList[i] = *product.ToResponse()
	}

	// Build success message
	message := "Products retrieved successfully"
	var filters []string

	if search != "" {
		filters = append(filters, "search: "+search)
	}

	if len(filters) > 0 {
		message += fmt.Sprintf(" (filtered by %s)", strings.Join(filters, " | "))
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(utils.SuccessPaginatedResponse{
		Success: true,
		Message: message,
		Data:    productList,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetProduct retrieves a single product by ID
// @Summary Get Product
// @Description Retrieve a single product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Product}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/products/{id} [get]
func (pc *ProductController) GetProduct(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var product models.Product
	if err := pc.DB.Where("id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Product with id " + id + " not found.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Product retrieved successfully",
		Data:    product.ToResponse(),
	})
}

// CreateProduct creates a new product
// @Summary Create Product
// @Description Create a new product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product details"
// @Success 201 {object} utils.SuccessResponse{data=models.Product}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/products [post]
func (pc *ProductController) CreateProduct(c fiber.Ctx) error {
	// Binding request body
	var req CreateProductRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert sku to uppercase and trim spaces
	req.SKU = strings.ToUpper(strings.TrimSpace(req.SKU))

	// Check for existing product with same code
	var existingProduct models.Product
	if err := pc.DB.Where("sku = ?", req.SKU).First(&existingProduct).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Product with code " + req.SKU + " already exists.",
		})
	}

	// Create new product
	newProduct := models.Product{
		SKU:      req.SKU,
		Name:     req.Name,
		Image:    req.Image,
		Variant:  req.Variant,
		Location: req.Location,
	}

	if err := pc.DB.Create(&newProduct).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to create product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Product created successfully",
		Data:    newProduct.ToResponse(),
	})
}

// UpdateProduct updates an existing product by ID
// @Summary Update Product
// @Description Update an existing product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body UpdateProductRequest true "Updated product details"
// @Success 200 {object} utils.SuccessResponse{data=models.Product}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/products/{id} [put]
func (pc *ProductController) UpdateProduct(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var product models.Product
	if err := pc.DB.Where("id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Product with id " + id + " not found.",
		})
	}

	// Binding request body
	var req UpdateProductRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Convert SKU to uppercase and trim spaces
	req.SKU = strings.ToUpper(strings.TrimSpace(req.SKU))

	// Check for existing product with same SKU (excluding current product)
	var existingProduct models.Product
	if err := pc.DB.Where("sku = ? AND id != ?", req.SKU, id).First(&existingProduct).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Product with SKU " + req.SKU + " already exists.",
		})
	}

	// Update product fields
	product.SKU = req.SKU
	product.Name = req.Name
	product.Image = req.Image
	product.Variant = req.Variant
	product.Location = req.Location

	if err := pc.DB.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to update product",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Product updated successfully",
		Data:    product.ToResponse(),
	})
}

// DeleteProduct deletes a product by ID
// @Summary Delete Product
// @Description Delete a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/products/{id} [delete]
func (pc *ProductController) DeleteProduct(c fiber.Ctx) error {
	// Parse id parameter
	id := c.Params("id")
	var product models.Product
	if err := pc.DB.Where("id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Product with id " + id + " not found.",
		})
	}

	// Delete product (also deletes associated records if any due to foreign key constraints)
	if err := pc.DB.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse{
			Success: false,
			Error:   "Failed to delete product",
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse{
		Success: true,
		Message: "Product deleted successfully",
	})
}
