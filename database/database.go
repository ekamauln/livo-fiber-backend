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
		&models.Box{},
		// &models.Store{},
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

// Seeds initial data into the database
func SeedInitialRole() error {
	log.Println("🌱 Seeding initial role data into the database...")

	// Create initial roles
	roles := []models.Role{
		// Highest privilege roles
		{RoleName: "developer", Hierarchy: 1},
		// Management roles
		{RoleName: "superadmin", Hierarchy: 10},
		{RoleName: "coordinator", Hierarchy: 10},
		{RoleName: "hrd", Hierarchy: 10},
		// Operational roles
		{RoleName: "admin", Hierarchy: 15},
		{RoleName: "finance", Hierarchy: 15},
		// Worker roles
		{RoleName: "picker", Hierarchy: 20},
		{RoleName: "qc-ribbon", Hierarchy: 20},
		{RoleName: "qc-online", Hierarchy: 20},
		{RoleName: "outbound", Hierarchy: 20},
		// Lowest privilege role
		{RoleName: "guest", Hierarchy: 99},
	}

	for _, roleData := range roles {
		var existingRole models.Role
		result := DB.Where("role_name = ?", roleData.RoleName).First(&existingRole)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new role
			role := models.Role{
				RoleName:  roleData.RoleName,
				Hierarchy: roleData.Hierarchy,
			}

			if err := DB.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %w", roleData.RoleName, err)
			}
		}
	}

	log.Println("✅ Roles seeding completed successfully")
	return nil
}

func SeedInitialBox() error {
	log.Println("🌱 Seeding initial box data into the database...")

	// Create initial boxes
	boxes := []models.Box{
		{BoxCode: "AX", BoxName: "Axon"},
		{BoxCode: "DR", BoxName: "DeParcel Ribbon"},
		{BoxCode: "AS", BoxName: "Axon Store"},
		{BoxCode: "AL", BoxName: "Aqualivo"},
		{BoxCode: "LM", BoxName: "Livo Mall"},
		{BoxCode: "LI", BoxName: "Livo ID"},
		{BoxCode: "BI", BoxName: "Bion"},
		{BoxCode: "AI", BoxName: "Axon ID"},
		{BoxCode: "AM", BoxName: "Axon Mall"},
		{BoxCode: "AS", BoxName: "Aqualivo Store"},
		{BoxCode: "RP", BoxName: "Rumah Pita"},
		{BoxCode: "SL", BoxName: "Sporti Livo"},
		{BoxCode: "LT", BoxName: "Livotech"},
	}

	for _, boxData := range boxes {
		var existingBox models.Box
		result := DB.Where("box_code = ?", boxData.BoxCode).First(&existingBox)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new box
			box := models.Box{
				BoxCode: boxData.BoxCode,
				BoxName: boxData.BoxName,
			}

			if err := DB.Create(&box).Error; err != nil {
				return fmt.Errorf("failed to create box %s: %w", boxData.BoxCode, err)
			}
		}
	}

	log.Println("✅ Boxes seeding completed successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
