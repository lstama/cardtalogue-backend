package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	Port                 string
	MigrationsPath       string
	JWTSecret            string
	DefaultAdminEmail    string
	DefaultAdminPassword string
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://user:password@localhost/dbname?sslmode=disable"),
		Port:                 getEnv("PORT", "8080"),
		MigrationsPath:       getEnv("MIGRATIONS_PATH", "file://migrations"),
		JWTSecret:            getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		DefaultAdminEmail:    getEnv("DEFAULT_ADMIN_EMAIL", "admin@example.com"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
