package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	"github.com/lstama/cardtalogue-backend/internal/config"
	"github.com/lstama/cardtalogue-backend/internal/database"
	"github.com/lstama/cardtalogue-backend/internal/handlers"
	"github.com/lstama/cardtalogue-backend/internal/logging"
	"github.com/lstama/cardtalogue-backend/internal/middleware"
	"github.com/lstama/cardtalogue-backend/internal/otel"
	"github.com/lstama/cardtalogue-backend/internal/repository"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "github.com/lstama/cardtalogue-backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Card Catalogue API
// @version         1.0
// @description     A card catalogue API with JWT authentication and role-based access control
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.Load()

	// Structured logging (slog), driven by LOG_FORMAT / LOG_LEVEL.
	logger := logging.New(*cfg)

	// Initialize OpenTelemetry (tracer + meter providers). When exporters are
	// set to "none" the providers are no-ops but instrumentation still runs.
	otelProvider, otelCleanup, err := otel.Init(context.Background(), *cfg, logger)
	if err != nil {
		logger.Error("otel: initialization failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelCleanup(ctx); err != nil {
			logger.Warn("otel: shutdown failed", "error", err)
		}
	}()

	// Initialize database. For MySQL, route the connection through the OTel
	// instrumented driver so queries produce spans + metrics.
	var dbConn gorm.ConnPool
	if cfg.DatabaseDriver == "mysql" {
		dbConn, err = otel.InstrumentMySQLConn(otelProvider.Tracer, otelProvider.Meter, cfg.DatabaseURL)
		if err != nil {
			logger.Error("database: instrumented connection failed", "error", err)
			os.Exit(1)
		}
	}
	db, err := database.NewConnectionWithConn(cfg.DatabaseURL, cfg.DatabaseDriver, dbConn)
	if err != nil {
		logger.Error("database: connection failed", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL, cfg.DatabaseDriver, cfg.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create default admin user
	if err := database.CreateDefaultAdmin(db, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword); err != nil {
		log.Printf("Warning: Failed to create default admin: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	cardRepo := repository.NewCardRepository(db)
	gameRepo := repository.NewGameRepository(db)
	languageRepo := repository.NewLanguageRepository(db)
	packRepo := repository.NewPackRepository(db)
	rarityRepo := repository.NewRarityRepository(db)
	storageRepo := repository.NewStorageRepository(db)
	mediaRepo := repository.NewMediaRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	cardService := service.NewCardService(cardRepo, mediaRepo)
	gameService := service.NewGameService(gameRepo)
	languageService := service.NewLanguageService(languageRepo)
	packService := service.NewPackService(packRepo)
	rarityService := service.NewRarityService(rarityRepo)
	storageService := service.NewStorageService(storageRepo)

	// Initialize card recognition service if API URL is configured
	var mediaService service.MediaService
	if cfg.CardScanAPIURL != "" {
		recognitionService := service.NewCardRecognitionService(cfg.CardScanAPIURL)
		mediaService = service.NewMediaServiceWithRecognition(mediaRepo, recognitionService)
	} else {
		mediaService = service.NewMediaService(mediaRepo)
	}

	// Setup router using the contract-first strict handler.
	router := setupRouter(
		otelProvider,
		cfg,
		&handlers.Services{
			Auth:      authService,
			Users:     userService,
			Games:     gameService,
			Languages: languageService,
			Packs:     packService,
			Rarities:  rarityService,
			Storages:  storageService,
			Media:     mediaService,
			Cards:     cardService,
		},
	)

	// Start server with graceful shutdown.
	logger.Info("server starting", "port", cfg.Port, "environment", cfg.Environment)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Warn("server forced to shutdown", "error", err)
	}
}

// otelHandler returns the Prometheus /metrics handler from the otel provider,
// falling back to a minimal 200 handler when instrumentation is disabled.
func otelHandler(p *otel.Provider) http.Handler {
	if p == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# metrics disabled\n"))
		})
	}
	return p.MetricsHandler()
}

func setupRouter(
	otelProvider *otel.Provider,
	cfg *config.Config,
	services *handlers.Services,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// OpenTelemetry instrumentation. HTTP middleware records server spans and
	// per-route request metrics; the OTel Prometheus registry serves /metrics.
	if otelProvider != nil {
		router.Use(otel.HTTPMiddleware(otelProvider.Tracer, otelProvider.ServiceName))
		router.Use(otel.HTTPMetricsMiddleware(otelProvider.Meter))
	}
	router.GET("/metrics", gin.WrapH(otelHandler(otelProvider)))

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendHosts,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "X-Cap-Token", "captcha-response"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Contract validation is OPT-IN (API_CONTRACT_VALIDATE=1). When unset it is a
	// no-op passthrough, so enabling it never changes default runtime behavior.
	if os.Getenv(middleware.EnvContractValidate) == "1" {
		router.Use(middleware.ContractValidation())
	}

	// Contract-first strict handler wiring: generated routes + per-operation auth middleware.
	shared := handlers.NewShared(services, cfg)
	ginStrict := gen.NewStrictHandlerWithOptions(
		shared,
		[]gen.StrictMiddlewareFunc{shared.AuthMiddleware()},
		gen.StrictGinServerOptions{},
	)
	api := router.Group("/api/v1")
	gen.RegisterHandlersWithOptions(api, ginStrict, gen.GinServerOptions{BaseURL: ""})

	// Legacy (outside-contract) media endpoints, served from the shared handler.
	router.GET("/media/:filename", shared.ServeMedia)
	router.POST("/media/upload", shared.UploadImage)
	router.POST("/media/scan", shared.ScanImage)

	// Add swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
