package database

import (
	"fmt"
	"livo-fiber-backend/config"
	"livo-fiber-backend/models"
	"livo-fiber-backend/utils"
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
		&models.Channel{},
		&models.Expedition{},
		&models.Store{},
		&models.Product{},
		&models.Order{},
		&models.OrderDetail{},
		&models.QCRibbon{},
		&models.QCRibbonDetail{},
		&models.QCOnline{},
		&models.QCOnlineDetail{},
		&models.Outbound{},
		&models.LostFound{},
		&models.Return{},
		&models.ReturnDetail{},
		&models.PickedOrder{},
		&models.Return{},
		&models.ReturnDetail{},
		&models.LostFound{},
		&models.Complain{},
		&models.ComplainUserDetail{},
		&models.ComplainProductDetail{},
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
		{RoleName: "warehouse", Hierarchy: 20},
		{RoleName: "picker", Hierarchy: 20},
		{RoleName: "qc-ribbon", Hierarchy: 20},
		{RoleName: "qc-online", Hierarchy: 20},
		{RoleName: "outbound", Hierarchy: 20},
		{RoleName: "security", Hierarchy: 20},
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
		{BoxCode: "1", BoxName: "001"},
		{BoxCode: "2", BoxName: "002"},
		{BoxCode: "A", BoxName: "Polos A"},
		{BoxCode: "B", BoxName: "Polos B"},
		{BoxCode: "K", BoxName: "Kawat"},
		{BoxCode: "R", BoxName: "Ribbon"},
		{BoxCode: "PK", BoxName: "Panjang Kecil"},
		{BoxCode: "PB", BoxName: "Panjang Besar"},
		{BoxCode: "SF", BoxName: "Single Face"},
		{BoxCode: "L", BoxName: "Layer"},
		{BoxCode: "X", BoxName: "Dos Bekas"},
		{BoxCode: "KRG", BoxName: "Karung"},
		{BoxCode: "17", BoxName: "1730"},
		{BoxCode: "20", BoxName: "2030"},
		{BoxCode: "25", BoxName: "2535"},
		{BoxCode: "30", BoxName: "3040"},
		{BoxCode: "35", BoxName: "3550"},
		{BoxCode: "40", BoxName: "4050"},
		{BoxCode: "75", BoxName: "5075"},
		{BoxCode: "85", BoxName: "8525"},
		{BoxCode: "70", BoxName: "7020"},
		{BoxCode: "50", BoxName: "6050"},
		{BoxCode: "KR", BoxName: "Kantong Kresek"},
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

func SeedInitialChannel() error {
	log.Println("🌱 Seeding initial channel data into the database...")

	// Create initial channels
	channels := []models.Channel{
		{ChannelCode: "SP", ChannelName: "Shopee"},
		{ChannelCode: "TP", ChannelName: "Tokopedia"},
		{ChannelCode: "LA", ChannelName: "Lazada"},
		{ChannelCode: "BU", ChannelName: "Bukalapak"},
		{ChannelCode: "BL", ChannelName: "Blibli"},
		{ChannelCode: "TT", ChannelName: "Tiktok"},
	}

	for _, channelData := range channels {
		var existingChannel models.Channel
		result := DB.Where("channel_code = ?", channelData.ChannelCode).First(&existingChannel)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new channel
			channel := models.Channel{
				ChannelCode: channelData.ChannelCode,
				ChannelName: channelData.ChannelName,
			}

			if err := DB.Create(&channel).Error; err != nil {
				return fmt.Errorf("failed to create channel %s: %w", channelData.ChannelCode, err)
			}
		}
	}

	log.Println("✅ Channels seeding completed successfully")
	return nil
}

func SeedInitialExpedition() error {
	log.Println("🌱 Seeding initial expedition data into the database...")

	// Create initial expeditions
	expeditions := []models.Expedition{
		{ExpeditionCode: "TKP0", ExpeditionName: "JNE/ID-Express", ExpeditionSlug: "jne-id-express", ExpeditionColor: "#006072"}, // JNE/ID Express
		{ExpeditionCode: "PJ", ExpeditionName: "Offline", ExpeditionSlug: "offline", ExpeditionColor: "#000000"},                 // Offline
		{ExpeditionCode: "INS", ExpeditionName: "Instant", ExpeditionSlug: "instant", ExpeditionColor: "#00d0dd"},                // Instant
		{ExpeditionCode: "BLMP", ExpeditionName: "Paxel", ExpeditionSlug: "paxel", ExpeditionColor: "#5f50a0"},                   // Paxel
		{ExpeditionCode: "LX", ExpeditionName: "LEX", ExpeditionSlug: "lex", ExpeditionColor: "#0c5eb4"},                         // LEX
		{ExpeditionCode: "NL", ExpeditionName: "LEX", ExpeditionSlug: "lex", ExpeditionColor: "#0c5eb4"},                         // LEX
		{ExpeditionCode: "JN", ExpeditionName: "LEX", ExpeditionSlug: "lex", ExpeditionColor: "#0c5eb4"},                         // LEX
		{ExpeditionCode: "JZ", ExpeditionName: "LEX", ExpeditionSlug: "lex", ExpeditionColor: "#0c5eb4"},                         // LEX
		{ExpeditionCode: "SP", ExpeditionName: "SPX", ExpeditionSlug: "spx", ExpeditionColor: "#ff7300"},                         // SPX
		{ExpeditionCode: "ID2", ExpeditionName: "SPX", ExpeditionSlug: "spx", ExpeditionColor: "#ff7300"},                        // SPX
		{ExpeditionCode: "TSA", ExpeditionName: "AnterAja", ExpeditionSlug: "anteraja", ExpeditionColor: "#ff007a"},              // AnterAja
		{ExpeditionCode: "1100", ExpeditionName: "AnterAja", ExpeditionSlug: "anteraja", ExpeditionColor: "#ff007a"},             // AnterAja
		{ExpeditionCode: "TAA", ExpeditionName: "AnterAja", ExpeditionSlug: "anteraja", ExpeditionColor: "#ff007a"},              // AnterAja
		{ExpeditionCode: "TLJX", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                       // JNE
		{ExpeditionCode: "41", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                         // JNE
		{ExpeditionCode: "CM", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                         // JNE
		{ExpeditionCode: "BLIJ", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                       // JNE
		{ExpeditionCode: "JT", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                         // JNE
		{ExpeditionCode: "TG", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                         // JNE
		{ExpeditionCode: "TLJR", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                       // JNE
		{ExpeditionCode: "TLJC", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                       // JNE
		{ExpeditionCode: "JNE", ExpeditionName: "JNE", ExpeditionSlug: "jne", ExpeditionColor: "#032078"},                        // JNE
		{ExpeditionCode: "JO", ExpeditionName: "J&T Express", ExpeditionSlug: "j&t-express", ExpeditionColor: "#ff0000"},         // J&T Express
		{ExpeditionCode: "JD", ExpeditionName: "J&T Express", ExpeditionSlug: "j&t-express", ExpeditionColor: "#ff0000"},         // J&T Express
		{ExpeditionCode: "JJ", ExpeditionName: "J&T Express", ExpeditionSlug: "j&t-express", ExpeditionColor: "#ff0000"},         // J&T Express
		{ExpeditionCode: "JB", ExpeditionName: "J&T Express", ExpeditionSlug: "j&t-express", ExpeditionColor: "#ff0000"},         // J&T Express
		{ExpeditionCode: "JP", ExpeditionName: "J&T Express", ExpeditionSlug: "j&t-express", ExpeditionColor: "#ff0000"},         // J&T Express
		{ExpeditionCode: "JX", ExpeditionName: "J&T Express", ExpeditionSlug: "j&t-express", ExpeditionColor: "#ff0000"},         // J&T Express
		{ExpeditionCode: "TKJN", ExpeditionName: "J&T Express", ExpeditionSlug: "j&t-express", ExpeditionColor: "#ff0000"},       // J&T Express
		{ExpeditionCode: "IDS", ExpeditionName: "ID Express", ExpeditionSlug: "id-express", ExpeditionColor: "#b30000"},          // ID Express
		{ExpeditionCode: "TKP8", ExpeditionName: "ID Express", ExpeditionSlug: "id-express", ExpeditionColor: "#b30000"},         // ID Express
		{ExpeditionCode: "300", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},            // J&T Cargo
		{ExpeditionCode: "2012", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "2011", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "2010", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "2009", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "2008", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "2007", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "2006", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "2005", ExpeditionName: "J&T Cargo", ExpeditionSlug: "j&t-cargo", ExpeditionColor: "#008601"},           // J&T Cargo
		{ExpeditionCode: "TS", ExpeditionName: "Wahana", ExpeditionSlug: "wahana", ExpeditionColor: "#ffa100"},                   // Wahana
		{ExpeditionCode: "SIC", ExpeditionName: "SiCepat", ExpeditionSlug: "sicepat", ExpeditionColor: "#830000"},
	}

	for _, expeditionData := range expeditions {
		var existingExpedition models.Expedition
		result := DB.Where("expedition_code = ?", expeditionData.ExpeditionCode).First(&existingExpedition)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new expedition
			expedition := models.Expedition{
				ExpeditionCode:  expeditionData.ExpeditionCode,
				ExpeditionName:  expeditionData.ExpeditionName,
				ExpeditionSlug:  expeditionData.ExpeditionSlug,
				ExpeditionColor: expeditionData.ExpeditionColor,
			}

			if err := DB.Create(&expedition).Error; err != nil {
				return fmt.Errorf("failed to create expedition %s: %w", expeditionData.ExpeditionCode, err)
			}
		}
	}

	log.Println("✅ Expeditions seeding completed successfully")
	return nil
}

func SeedInitialStore() error {
	log.Println("🌱 Seeding initial store data into the database...")

	// Create initial stores
	stores := []models.Store{
		{StoreCode: "AX", StoreName: "Axon"},
		{StoreCode: "DR", StoreName: "DeParcel Ribbon"},
		{StoreCode: "AS", StoreName: "Axon Store"},
		{StoreCode: "AL", StoreName: "Aqualivo"},
		{StoreCode: "LM", StoreName: "Livo Mall"},
		{StoreCode: "LI", StoreName: "Livo ID"},
		{StoreCode: "BI", StoreName: "Bion"},
		{StoreCode: "AI", StoreName: "Axon ID"},
		{StoreCode: "AM", StoreName: "Axon Mall"},
		{StoreCode: "AS", StoreName: "Aqualivo Store"},
		{StoreCode: "RP", StoreName: "Rumah Pita"},
		{StoreCode: "SL", StoreName: "Sporti Livo"},
		{StoreCode: "LT", StoreName: "Livotech"},
	}

	for _, storeData := range stores {
		var existingStore models.Store
		result := DB.Where("store_code = ?", storeData.StoreCode).First(&existingStore)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new store
			store := models.Store{
				StoreCode: storeData.StoreCode,
				StoreName: storeData.StoreName,
			}

			if err := DB.Create(&store).Error; err != nil {
				return fmt.Errorf("failed to create store %s: %w", storeData.StoreCode, err)
			}
		}
	}

	log.Println("✅ Stores seeding completed successfully")
	return nil
}

func SeedInitialUser() error {
	log.Println("🌱 Seeding initial user data into the database...")

	// Define initial users
	type InitialUser struct {
		Username string
		Password string
		FullName string
		Email    string
		RoleName string
	}

	initialUsers := []InitialUser{
		{
			Username: "admin",
			Password: "12345678",
			FullName: "Administrator",
			Email:    "admin@example.com",
			RoleName: "developer",
		},
		{
			Username: "security",
			Password: "12345678",
			FullName: "Security",
			Email:    "security@example.com",
			RoleName: "security",
		},
	}

	for _, userData := range initialUsers {
		// Check if user already exists
		var existingUser models.User
		result := DB.Where("username = ?", userData.Username).First(&existingUser)

		if result.Error == gorm.ErrRecordNotFound {
			// Get role
			var role models.Role
			if err := DB.Where("role_name = ?", userData.RoleName).First(&role).Error; err != nil {
				return fmt.Errorf("%s role not found. Please seed roles first: %w", userData.RoleName, err)
			}

			// Hash password
			hashedPassword, err := utils.HashPassword(userData.Password)
			if err != nil {
				return fmt.Errorf("failed to hash password for %s: %w", userData.Username, err)
			}

			// Create user
			user := models.User{
				Username: userData.Username,
				Password: hashedPassword,
				FullName: userData.FullName,
				Email:    userData.Email,
				IsActive: true,
			}

			if err := DB.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user %s: %w", userData.Username, err)
			}

			// Assign role to user
			if err := DB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID).Error; err != nil {
				return fmt.Errorf("failed to assign role to user %s: %w", userData.Username, err)
			}

			log.Printf("✅ User '%s' created successfully (password: %s, role: %s)", userData.Username, userData.Password, userData.RoleName)
		} else {
			log.Printf("ℹ️ User '%s' already exists, skipping seed", userData.Username)
		}
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
