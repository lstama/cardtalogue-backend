package main

import (
	"log"
	"net/http"

	"github.com/lstama/cardtalogue-backend/internal/config"
	"github.com/lstama/cardtalogue-backend/internal/database"
	"github.com/lstama/cardtalogue-backend/internal/handlers"
	"github.com/lstama/cardtalogue-backend/internal/middleware"
	"github.com/lstama/cardtalogue-backend/internal/repository"
	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/lstama/cardtalogue-backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/penglongli/gin-metrics/ginmetrics"
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

	// Initialize database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, cfg)
	userHandler := handlers.NewUserHandler(userService)
	cardHandler := handlers.NewCardHandler(cardService)
	gameHandler := handlers.NewGameHandler(gameService)
	languageHandler := handlers.NewLanguageHandler(languageService)
	packHandler := handlers.NewPackHandler(packService)
	rarityHandler := handlers.NewRarityHandler(rarityService)
	storageHandler := handlers.NewStorageHandler(storageService)
	mediaHandler := handlers.NewMediaHandler(mediaService, cfg)

	// Setup router
	router := setupRouter(
		authHandler,
		userHandler,
		cardHandler,
		gameHandler,
		languageHandler,
		packHandler,
		rarityHandler,
		storageHandler,
		mediaHandler,
		cfg.JWTSecret,
		cfg.FrontendHosts,
	)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

func setupRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	cardHandler *handlers.CardHandler,
	gameHandler *handlers.GameHandler,
	languageHandler *handlers.LanguageHandler,
	packHandler *handlers.PackHandler,
	rarityHandler *handlers.RarityHandler,
	storageHandler *handlers.StorageHandler,
	mediaHandler *handlers.MediaHandler,
	jwtSecret string,
	frontendHosts []string,
) *gin.Engine {
	router := gin.Default()

	// Add metrics (this is the only addition)
	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/metrics")
	m.SetSlowTime(10)                              // Log slow requests (>10s)
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10}) // Histogram buckets
	m.Use(router)

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     frontendHosts,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health Check godoc
	// @Summary Health check endpoint
	// @Description Check if the API is running
	// @Tags System
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Media serving routes (public)
	router.GET("/media/:filename", mediaHandler.ServeMedia)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Optional auth middleware for public endpoints that need to differentiate authenticated/unauthenticated users
		v1.Use(middleware.OptionalJWTAuth(jwtSecret))

		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Public cards endpoint (optional auth)
		v1.GET("/cards", cardHandler.GetCards)
		v1.GET("cards/:id", cardHandler.GetCard)
		v1.GET("/cards/search", cardHandler.SearchCards)
		// Public games endpoint (optional auth)
		v1.GET("/games", gameHandler.GetGames)
		v1.GET("games/:id", gameHandler.GetGame)
		v1.GET("/games/search", gameHandler.SearchGames)
		// Public languages endpoint (optional auth)
		v1.GET("/languages", languageHandler.GetLanguages)
		v1.GET("languages/:id", languageHandler.GetLanguage)
		v1.GET("/languages/search", languageHandler.SearchLanguages)
		// Public packs endpoint (optional auth)
		v1.GET("/packs", packHandler.GetPacks)
		v1.GET("packs/:id", packHandler.GetPack)
		v1.GET("/packs/search", packHandler.SearchPacks)
		// Public rarities endpoint (optional auth)
		v1.GET("/rarities", rarityHandler.GetRarities)
		v1.GET("rarities/:id", rarityHandler.GetRarity)
		v1.GET("/rarities/search", rarityHandler.SearchRarities)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtSecret))
		{
			// User management (Admin only)
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin"))
			{
				users.POST("", userHandler.CreateUser)
				users.GET("", userHandler.GetUsers)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
			}

			// Profile management
			protected.GET("/profile", authHandler.GetProfile)
			protected.PUT("/profile/password", authHandler.ChangePassword)

			// Cards management
			cards := protected.Group("/cards")
			{
				// Editors and Admins can get protected data
				cards.GET("/protected", middleware.RequireRole("editor", "admin"), cardHandler.GetCardsProtected)
				cards.GET("/protected/:id", middleware.RequireRole("editor", "admin"), cardHandler.GetCardProtected)

				// Editors and Admins can create and update
				cards.POST("", middleware.RequireRole("editor", "admin"), cardHandler.CreateCard)
				cards.PUT("/:id", middleware.RequireRole("editor", "admin"), cardHandler.UpdateCard)
				cards.PUT("/protected/:id", middleware.RequireRole("editor", "admin"), cardHandler.UpdateCard)

				// Only Admins can delete
				cards.DELETE("/:id", middleware.RequireRole("admin"), cardHandler.DeleteCard)
			}

			// Games management
			games := protected.Group("/games")
			{
				games.POST("", middleware.RequireRole("editor", "admin"), gameHandler.CreateGame)
				games.PUT("/:id", middleware.RequireRole("editor", "admin"), gameHandler.UpdateGame)
				games.DELETE("/:id", middleware.RequireRole("admin"), gameHandler.DeleteGame)
			}

			// Language management
			languages := protected.Group("/languages")
			{
				languages.POST("", middleware.RequireRole("editor", "admin"), languageHandler.CreateLanguage)
				languages.PUT("/:id", middleware.RequireRole("editor", "admin"), languageHandler.UpdateLanguage)
				languages.DELETE("/:id", middleware.RequireRole("admin"), languageHandler.DeleteLanguage)
			}

			// Pack management
			packs := protected.Group("/packs")
			{
				packs.POST("", middleware.RequireRole("editor", "admin"), packHandler.CreatePack)
				packs.PUT("/:id", middleware.RequireRole("editor", "admin"), packHandler.UpdatePack)
				packs.DELETE("/:id", middleware.RequireRole("admin"), packHandler.DeletePack)
			}

			// Rarity management
			rarities := protected.Group("/rarities")
			{
				rarities.POST("", middleware.RequireRole("editor", "admin"), rarityHandler.CreateRarity)
				rarities.PUT("/:id", middleware.RequireRole("editor", "admin"), rarityHandler.UpdateRarity)
				rarities.DELETE("/:id", middleware.RequireRole("admin"), rarityHandler.DeleteRarity)
			}

			// Storage management
			storages := protected.Group("/storages")
			{
				// Storage management is only open to authenticated users
				storages.GET("", storageHandler.GetStorages)
				storages.GET("/:id", storageHandler.GetStorage)
				storages.GET("/search", storageHandler.SearchStorages)
				storages.POST("", middleware.RequireRole("editor", "admin"), storageHandler.CreateStorage)
				storages.PUT("/:id", middleware.RequireRole("editor", "admin"), storageHandler.UpdateStorage)
				storages.DELETE("/:id", middleware.RequireRole("admin"), storageHandler.DeleteStorage)
			}

			// Media management
			media := protected.Group("/media")
			{
				media.POST("/upload", middleware.RequireRole("editor", "admin"), mediaHandler.UploadImage)
			}
		}
	}

	// Add swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
