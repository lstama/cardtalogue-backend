package handlers

import (
	"github.com/lstama/cardtalogue-backend/internal/config"
	"github.com/lstama/cardtalogue-backend/internal/middleware"
	"github.com/lstama/cardtalogue-backend/internal/service"
)

// Services holds all service dependencies for the strict (contract) handlers.

type Services struct {
	Auth      service.AuthService
	Cards     service.CardService
	Games     service.GameService
	Languages service.LanguageService
	Media     service.MediaService
	Packs     service.PackService
	Rarities  service.RarityService
	Storages  service.StorageService
	Users     service.UserService
}

type shared struct {
	services *Services
	cfg      *config.Config
	verifier middleware.Verifier
}

func NewShared(services *Services, cfg *config.Config) *shared {
	return &shared{
		services: services,
		cfg:      cfg,
		verifier: middleware.NewVerifier(cfg.CaptchaMode, cfg.Environment, cfg.CapURL, cfg.CapSecret),
	}
}
