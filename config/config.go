package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Database settings
	DbHost    string
	DbPort    string
	DbUser    string
	DbPass    string
	DbName    string
	DbSslMode string
	DbTz      string

	// Server settings
	Env      string
	Port     string
	AppUrl   string
	AppName  string
	LogLevel string

	// Security settings
	PasetoSymmetricKey string
	CorsOrigins        []string
	AccessTokenTTL     int // minutes
	RefreshTokenTTL    int // days
}

func LoadConfig() *Config {
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://192.168.31.147:3000"
	}

	accessTokenTTL, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_TTL"))
	if err != nil || accessTokenTTL <= 0 {
		accessTokenTTL = 60 // default 60 minutes
	}

	refreshTokenTTL, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_TTL"))
	if err != nil || refreshTokenTTL <= 0 {
		refreshTokenTTL = 7 // default 7 days
	}

	return &Config{
		// Database settings
		DbHost:    getEnv("DB_HOST", "localhost"),
		DbPort:    getEnv("DB_PORT", "5432"),
		DbUser:    getEnv("DB_USER", "postgres"),
		DbPass:    getEnv("DB_PASSWORD", "password"),
		DbName:    getEnv("DB_NAME", "app_db"),
		DbSslMode: getEnv("DB_SSLMODE", "disable"),
		DbTz:      getEnv("DB_TZ", "UTC"),

		// Server settings
		Env:      getEnv("ENV", "development"),
		Port:     getEnv("PORT", "8040"),
		AppUrl:   getEnv("APP_URL", "http://localhost:8040"),
		AppName:  getEnv("APP_NAME", "MyApp"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),

		// Security settings
		PasetoSymmetricKey: getEnv("PASETO_SYMMETRIC_KEY", "your-32-character-secret-key!!"), // Must be 32 chars
		CorsOrigins:        strings.Split(corsOrigins, ","),
		AccessTokenTTL:     accessTokenTTL,  // 15 minutes
		RefreshTokenTTL:    refreshTokenTTL, // 7 days
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
