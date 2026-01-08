package database

import (
	"fmt"
	"livo-fiber-backend/config"
	"livo-fiber-backend/models"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbUser,
		cfg.DbPass,
		cfg.DbName,
		cfg.DbSslMode,
		cfg.DbTz,
	)

	// Configure GORM logger based on environment
	gormLogger := logger.Default
	if cfg.Env == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	maxRetries := 5
	retryInterval := time.Second * 10

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", attempt, maxRetries)

		var err error
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,

			NowFunc: func() time.Time {
				location, _ := time.LoadLocation(cfg.DbTz)
				return time.Now().In(location)
			},
		})

		if err == nil {
			sqlDB, err := DB.DB()

			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					// Set connection pool settings
					sqlDB.SetMaxIdleConns(10)
					sqlDB.SetMaxOpenConns(100)
					sqlDB.SetConnMaxLifetime(time.Hour)

					log.Println("Database connection established.")
					return nil
				}
				log.Printf("x Database ping failed: %v", err)
			} else {
				log.Printf("x Failed to get database instance: %v", err)
			}
		} else {
			log.Printf("x Failed to connect to database: %v", err)
		}

		if attempt < maxRetries {
			log.Printf("Retrying in %s...", retryInterval)
			time.Sleep(retryInterval)
		}

	}

	return fmt.Errorf("Failed to connect to database after %d attempts.", maxRetries)
}

// MigrateDatabase performs automatic migration of database schemas
func MigrateDatabase() error {
	log.Println("🔄 Starting database migration...")

	err := DB.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.Session{},
		// &models.Store{},
		// &models.Box{},
		// &models.Expedition{},
		// &models.Product{},
		// &models.Channel{},
		// &models.Order{},
		// &models.OrderDetail{},
		// &models.QCRibbon{},
		// &models.QCRibbonDetail{},
		// &models.QCOnline{},
		// &models.QCOnlineDetail{},
		// &models.Outbound{},
		// &models.Return{},
		// &models.Complain{},
		// &models.ComplainUserDetail{},
		// &models.ComplainProductDetail{},
		// &models.LostFound{},
		// &models.PickedOrder{},
		// &models.Return{},
		// &models.ReturnDetail{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// SeedDB seeds initial data into the database
func SeedDB() error {
	log.Println("🌱 Seeding initial data into the database...")

	// Create initial roles
	roles := []models.Role{
		// Highest privilege roles
		{Name: "developer", Hierarchy: 1},
		// Management roles
		{Name: "superadmin", Hierarchy: 10},
		{Name: "coordinator", Hierarchy: 10},
		{Name: "hrd", Hierarchy: 10},
		// Operational roles
		{Name: "admin", Hierarchy: 15},
		{Name: "finance", Hierarchy: 15},
		// Worker roles
		{Name: "picker", Hierarchy: 20},
		{Name: "qc-ribbon", Hierarchy: 20},
		{Name: "qc-online", Hierarchy: 20},
		{Name: "outbound", Hierarchy: 20},
		// Lowest privilege role
		{Name: "guest", Hierarchy: 99},
	}

	for _, roleData := range roles {
		var existingRole models.Role
		result := DB.Where("name = ?", roleData.Name).First(&existingRole)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new role
			role := models.Role{
				Name:      roleData.Name,
				Hierarchy: roleData.Hierarchy,
			}

			if err := DB.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %w", roleData.Name, err)
			}
		}
	}

	log.Println("✅ Database seeding completed successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
