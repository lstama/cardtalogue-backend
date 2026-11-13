package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment          string
	DatabaseURL          string
	DatabaseDriver       string
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
	CaptchaMode          string
	CapURL               string
	CapSecret            string

	// Observability (OpenTelemetry)
	OTEL_SERVICE_NAME              string
	OTEL_EXPORTER_OTLP_ENDPOINT    string
	OTEL_EXPORTER_OTLP_PROTOCOL    string
	OTEL_TRACES_EXPORTER           string
	OTEL_METRICS_EXPORTER          string
	OTEL_METRICS_ADDRESS           string
	OTEL_METRICS_PORT              string
	OTEL_RESOURCE_ATTRIBUTES       string
	OTEL_BSP_SCHEDULE_DELAY        int
	OTEL_BSP_MAX_QUEUE_SIZE        int
	OTEL_BSP_MAX_EXPORT_BATCH_SIZE int

	// Logging
	LOG_LEVEL  string
	LOG_FORMAT string
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	frontendHosts := strings.Split(getEnv("FRONTEND_HOSTS", "http://localhost:3000,http://127.0.0.1:3000,http://172.17.0.1:3000,http://100.64.0.4:3000"), ",")

	environment := getEnv("ENVIRONMENT", "development")
	var sec bool
	if environment == "production" {
		sec = true
	} else {
		sec = false
	}

	return &Config{
		Environment:          environment,
		DatabaseURL:          getEnv("DATABASE_URL", "file:./data.db"),
		DatabaseDriver:       getEnv("DATABASE_DRIVER", "sqlite"),
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
		CaptchaMode:          getEnv("CAPTCHA_MODE", "off"),
		CapURL:               getEnv("CAP_URL", ""),
		CapSecret:            getEnv("CAP_SECRET", ""),

		// Observability (OpenTelemetry)
		OTEL_SERVICE_NAME:              getEnv("OTEL_SERVICE_NAME", "cardtalogue-backend"),
		OTEL_EXPORTER_OTLP_ENDPOINT:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
		OTEL_EXPORTER_OTLP_PROTOCOL:    getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf"),
		OTEL_TRACES_EXPORTER:           getEnv("OTEL_TRACES_EXPORTER", "otlp"),
		OTEL_METRICS_EXPORTER:          getEnv("OTEL_METRICS_EXPORTER", "otlp"),
		OTEL_METRICS_ADDRESS:           getEnv("OTEL_METRICS_ADDRESS", "0.0.0.0"),
		OTEL_METRICS_PORT:              getEnv("OTEL_METRICS_PORT", "8888"),
		OTEL_RESOURCE_ATTRIBUTES:       getEnv("OTEL_RESOURCE_ATTRIBUTES", ""),
		OTEL_BSP_SCHEDULE_DELAY:        getEnvInt("OTEL_BSP_SCHEDULE_DELAY", 5000),
		OTEL_BSP_MAX_QUEUE_SIZE:        getEnvInt("OTEL_BSP_MAX_QUEUE_SIZE", 2048),
		OTEL_BSP_MAX_EXPORT_BATCH_SIZE: getEnvInt("OTEL_BSP_MAX_EXPORT_BATCH_SIZE", 512),

		// Logging
		LOG_LEVEL:  getEnv("LOG_LEVEL", "info"),
		LOG_FORMAT: getEnv("LOG_FORMAT", "json"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
