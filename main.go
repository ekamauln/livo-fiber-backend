package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"livo-fiber-backend/config"
	"livo-fiber-backend/database"
	_ "livo-fiber-backend/docs" // Import generated docs
	"livo-fiber-backend/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"
)

// @title Livo Fiber Backend API
// @version 1.0
// @description This is the API documentation for Livo Fiber Backend application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@livo.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8040
// @BasePath /
// @schemes http https

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	database.ConnectDatabase(cfg)
	database.MigrateDatabase()
	database.SeedDB()

	// Create Fiber app with go-joson
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		AppName:      "Livo Fiber Backend",
		ServerHeader: "Fiber",
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(helmet.New())

	// Configure CORS based on origins
	corsConfig := cors.Config{
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
	}

	// If origins contain wildcard, don't use credentials
	if len(cfg.CorsOrigins) == 1 && cfg.CorsOrigins[0] == "*" {
		corsConfig.AllowOrigins = []string{"*"}
		corsConfig.AllowCredentials = false
	} else {
		corsConfig.AllowOrigins = cfg.CorsOrigins
		corsConfig.AllowCredentials = true
	}

	app.Use(cors.New(corsConfig))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 60 * time.Second,
	}))

	// Setup routes
	routes.SetupRoutes(app, cfg, database.DB)

	// Start server
	port := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
