package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	Port                 string
	MigrationsPath       string
	JWTSecret            string
	DefaultAdminEmail    string
	DefaultAdminPassword string
	CommonHost           string
	FrontendHosts        []string
	Secure               bool
	APIBaseURL           string
	MediaUploadPath      string
	CardScanAPIURL       string
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	frontendHosts := strings.Split(getEnv("FRONTEND_HOSTS", "http://localhost:3000,http://127.0.0.1:3000,http://172.17.0.1:3000"), ",")

	environment := getEnv("ENVIRONMENT", "development")
	var sec bool
	if environment == "production" {
		sec = true
	} else {
		sec = false
	}

	return &Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://user:password@localhost/dbname?sslmode=disable"),
		Port:                 getEnv("PORT", "8080"),
		MigrationsPath:       getEnv("MIGRATIONS_PATH", "file://migrations"),
		JWTSecret:            getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		DefaultAdminEmail:    getEnv("DEFAULT_ADMIN_EMAIL", "admin@example.com"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "admin123"),
		CommonHost:           getEnv("COMMON_HOST", "localhost"),
		FrontendHosts:        frontendHosts,
		Secure:               sec,
		APIBaseURL:           getEnv("API_BASE_URL", "http://localhost:"+getEnv("PORT", "8080")),
		MediaUploadPath:      getEnv("MEDIA_UPLOAD_PATH", "./uploads/media"),
		CardScanAPIURL:       getEnv("CARD_SCAN_API_URL", "http://172.17.0.4:5000/api/recognize"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
