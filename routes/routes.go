package routes

import (
	"livo-fiber-backend/config"
	"livo-fiber-backend/controllers"
	"livo-fiber-backend/middleware"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, cfg *config.Config, db *gorm.DB) {

	// Controllers
	authController := controllers.NewAuthController(cfg, db)
	userController := controllers.NewUserController(db)
	roleController := controllers.NewRoleController(db)
	// boxController := controllers.NewBoxController()
	// storeController := controllers.NewStoreController()
	// channelController := controllers.NewChannelController()
	// productController := controllers.NewProductController()
	// expeditionController := controllers.NewExpeditionController()
	// orderController := controllers.NewOrderController()
	// qcOnlineController := controllers.NewQCOnlineController()
	// qcRibbonController := controllers.NewQCRibbonController()
	// outboundController := controllers.NewOutboundController()
	// reportController := controllers.NewReportController()
	// returnController := controllers.NewReturnController()
	// complainController := controllers.NewComplainController()

	// Public routes
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"Aplication": "Livo Fiber",
			"Version":    "1.0.0",
			"message":    "Health check successful",
			"status":     "ok",
			"Time":       time.Now().Format("02-01-2006 15:04:05"),
		})
	})

	// API Documentation routes - Serve static swagger files
	app.Get("/docs/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	app.Get("/docs/swagger.yaml", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.yaml")
	})

	// Swagger UI HTML page
	app.Get("/docs", func(c fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="SwaggerUI" />
  <title>Livo API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '/docs/swagger.json',
      dom_id: '#swagger-ui',
    });
  };
</script>
</body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// RapiDoc HTML page
	app.Get("/rapidoc", func(c fiber.Ctx) error {
		html := `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Livo API Documentation</title>
  <script type="module" src="https://unpkg.com/rapidoc/dist/rapidoc-min.js"></script>
</head>
<body>
  <rapi-doc
        spec-url="/docs/swagger.yaml"
        theme="dark"
        bg-color="#1a1a1a"
        text-color="#f0f0f0"
        primary-color="#4caf50"
        nav-bg-color="#2d2d2d"
        nav-text-color="#ffffff"
        nav-hover-bg-color="#404040"
        render-style="read"
        layout="column"
        schema-style="tree"
        show-header="true"
        show-info="true"
        allow-try="true"
        allow-authentication="true"
        allow-spec-url-load="false"
        allow-spec-file-load="false"
        allow-search="true"
        allow-advanced-search="true"
        show-method-in-nav-bar="as-colored-block"
        use-path-in-nav-bar="true"
        response-area-height="400px"
        api-key-name="Authorization"
        api-key-location="header"
        api-key-value="Bearer "
        default-schema-tab="model"
        schema-expand-level="2"
        schema-description-expanded="true"
        schema-hide-read-only="never"
        schema-hide-write-only="never"
        fetch-credentials="include"
        heading-text="JWT Auth Service API"
        goto-path=""
        fill-request-fields-with-example="true"
        persist-auth="true"
    >
        <img 
            slot="logo" 
            src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%234caf50'%3E%3Cpath d='M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z'/%3E%3C/svg%3E"
            style="width: 40px; height: 40px; margin-right: 10px;"
        />
    </rapi-doc>
</body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.RefreshToken)

	// CSRF token endpoint for web clients
	auth.Get("/csrf-token", middleware.CSRFMiddleware(), func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"csrf_token": c.Locals("csrf_token"),
		})
	})

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Note: CSRF middleware removed for API clients (HTTPie, Postman, mobile apps)
	// If you need CSRF protection for web clients, apply it selectively to specific routes
	// protected.Use(middleware.CSRFMiddleware())

	// Auth protected routes
	protectedAuth := protected.Group("/auth")
	protectedAuth.Post("/logout", authController.Logout)

	// User routes
	users := protected.Group("/users")
	users.Get("/", userController.GetUsers)
	users.Get("/:id", userController.GetUser)
	users.Post("/", middleware.RoleMiddleware([]string{"developer", "superadmin", "hrd"}), userController.CreateUser)
	users.Put("/:id", middleware.RoleMiddleware([]string{"developer", "superadmin", "hrd"}), userController.UpdateUser)
	users.Put("/:id/password", middleware.RoleMiddleware([]string{"developer", "superadmin", "hrd"}), userController.UpdatePassword)
	users.Delete("/:id", middleware.RoleMiddleware([]string{"developer"}), userController.DeleteUser)
	users.Post("/:id/roles", middleware.RoleMiddleware([]string{"developer", "superadmin", "hrd"}), userController.AssignRole)
	users.Delete("/:id/roles", middleware.RoleMiddleware([]string{"developer", "superadmin", "hrd"}), userController.RemoveRole)
	users.Get("/:id/sessions", userController.GetSessions)

	// Role routes
	roles := protected.Group("/roles")
	roles.Get("/", roleController.GetRoles)
	roles.Get("/:id", roleController.GetRole)
	roles.Post("/", middleware.RoleMiddleware([]string{"admin", "developer"}), roleController.CreateRole)
	roles.Put("/:id", middleware.RoleMiddleware([]string{"admin", "developer"}), roleController.UpdateRole)
	roles.Delete("/:id", middleware.RoleMiddleware([]string{"admin", "developer"}), roleController.DeleteRole)

	// Box routes
	// boxRoutes := protected.Group("/boxes")
	// boxRoutes.Get("/", boxController.GetBoxes)
	// boxRoutes.Get("/:id", boxController.GetBox)
	// boxRoutes.Post("/", middleware.RoleMiddleware([]string{"developer", "superadmin"}), boxController.CreateBox)
	// boxRoutes.Put("/:id", middleware.RoleMiddleware([]string{"developer", "superadmin"}), boxController.UpdateBox)
	// boxRoutes.Delete("/:id", middleware.RoleMiddleware([]string{"developer"}), boxController.DeleteBox)

	// Store routes
	// storeRoutes := protected.Group("/stores")
	// storeRoutes.Get("/", storeController.GetStores)
	// storeRoutes.Get("/:id", storeController.GetStore)
	// storeRoutes.Post("/", middleware.RoleMiddleware([]string{"developer", "superadmin"}), storeController.CreateStore)
	// storeRoutes.Put("/:id", middleware.RoleMiddleware([]string{"developer", "superadmin"}), storeController.UpdateStore)
	// storeRoutes.Delete("/:id", middleware.RoleMiddleware([]string{"developer"}), storeController.DeleteStore)

	// Channel routes
	// channelRoutes := protected.Group("/channels")
	// channelRoutes.Get("/", channelController.GetChannels)
	// channelRoutes.Get("/:id", channelController.GetChannel)
	// channelRoutes.Post("/", middleware.RoleMiddleware([]string{"developer", "superadmin"}), channelController.CreateChannel)
	// channelRoutes.Put("/:id", middleware.RoleMiddleware([]string{"developer", "superadmin"}), channelController.UpdateChannel)
	// channelRoutes.Delete("/:id", middleware.RoleMiddleware([]string{"developer"}), channelController.DeleteChannel)

	// Product routes
	// productRoutes := protected.Group("/products")
	// productRoutes.Get("/", productController.GetProducts)
	// productRoutes.Get("/:id", productController.GetProduct)
	// productRoutes.Post("/", middleware.RoleMiddleware([]string{"developer", "superadmin"}), productController.CreateProduct)
	// productRoutes.Put("/:id", middleware.RoleMiddleware([]string{"developer", "superadmin"}), productController.UpdateProduct)
	// productRoutes.Delete("/:id", middleware.RoleMiddleware([]string{"developer"}), productController.DeleteProduct)

	// Expedition routes
	// expeditionRoutes := protected.Group("/expeditions")
	// expeditionRoutes.Get("/", expeditionController.GetExpeditions)
	// expeditionRoutes.Get("/:id", expeditionController.GetExpedition)
	// expeditionRoutes.Post("/", middleware.RoleMiddleware([]string{"developer", "superadmin"}), expeditionController.CreateExpedition)
	// expeditionRoutes.Put("/:id", middleware.RoleMiddleware([]string{"developer", "superadmin"}), expeditionController.UpdateExpedition)
	// expeditionRoutes.Delete("/:id", middleware.RoleMiddleware([]string{"developer"}), expeditionController.DeleteExpedition)

	// Order routes
	// orderRoutes := protected.Group("/orders")
	// orderRoutes.Get("/", orderController.GetOrders)
	// orderRoutes.Get("/:id", orderController.GetOrder)
	// orderRoutes.Put("/:id/status/qc-process", orderController.QCProcessStatus)
	// orderRoutes.Put("/:id/status/picking-completed", orderController.PickingCompletedStatus)

	// Order router for admin
	// orderRoutes.Post("/", middleware.RoleMiddleware([]string{"developer", "superadmin", "admin"}), orderController.CreateOrder)
	// orderRoutes.Post("/bulk", middleware.RoleMiddleware([]string{"developer", "superadmin", "admin"}), orderController.CreateBulkOrders)
	// orderRoutes.Put("/:id", middleware.RoleMiddleware([]string{"developer", "superadmin", "admin"}), orderController.UpdateOrder)
	// orderRoutes.Put("/:id/duplicate", middleware.RoleMiddleware([]string{"developer", "superadmin", "admin"}), orderController.DuplicateOrder)
	// orderRoutes.Put("/:id/cancel", middleware.RoleMiddleware([]string{"developer", "superadmin", "admin"}), orderController.CancelOrder)

	// Order router for coordinator
	// orderRoutes.Post("/:id/assign-picker", middleware.RoleMiddleware([]string{"developer", "superadmin", "coordinator"}), orderController.AssignPicker)
	// orderRoutes.Post("/:id/pending-picking", middleware.RoleMiddleware([]string{"developer", "superadmin", "coordinator"}), orderController.PendingPickingOrders)
	// orderRoutes.Get("/assigned", middleware.RoleMiddleware([]string{"developer", "superadmin", "coordinator"}), orderController.GetAssignedOrders)

	// QCRibbon routes
	// qcRibbonRoutes := protected.Group("/qc-ribbons")
	// qcRibbonRoutes.Get("/", qcRibbonController.GetQCRibbons)
	// qcRibbonRoutes.Post("/", qcRibbonController.CreateQCRibbon)

	// QCOnline routes
	// qcOnlineRoutes := protected.Group("/qc-onlines")
	// qcOnlineRoutes.Get("/", qcOnlineController.GetQCOnlines)
	// qcOnlineRoutes.Post("/", qcOnlineController.CreateQCOnline)

	// Outbound routes
	// outboundRoutes := protected.Group("/outbounds")
	// outboundRoutes.Get("/", outboundController.GetOutbounds)
	// outboundRoutes.Post("/", outboundController.CreateOutbound)
	// outboundRoutes.Put("/:id", outboundController.UpdateOutbound)

}
