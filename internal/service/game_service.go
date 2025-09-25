package service

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
)

type GameService interface {
	CreateGame(req *models.CreateGameRequest) (*models.GameResponse, error)
	GetGame(id uint) (*models.GameResponse, error)
	GetGames(limit, offset int) ([]*models.GameResponse, int64, error)
	UpdateGame(id uint, req *models.UpdateGameRequest) (*models.GameResponse, error)
	DeleteGame(id uint) error
	SearchGames(query_string string, limit, offset int) ([]*models.GameResponse, error)
}

type gameService struct {
	gameRepo repository.GameRepository
}

func NewGameService(gameRepo repository.GameRepository) GameService {
	return &gameService{gameRepo: gameRepo}
}

func (s *gameService) CreateGame(req *models.CreateGameRequest) (*models.GameResponse, error) {
	game := &models.Game{
		Name: req.Name,
	}
	if err := s.gameRepo.Create(game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	return s.toGameResponse(game), nil
}

func (s *gameService) GetGame(id uint) (*models.GameResponse, error) {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toGameResponse(game), nil
}

func (s *gameService) GetGames(limit, offset int) ([]*models.GameResponse, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	games, err := s.gameRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.gameRepo.Count()
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*models.GameResponse, len(games))
	for i, game := range games {
		responses[i] = s.toGameResponse(game)
	}
	return responses, total, nil
}

func (s *gameService) UpdateGame(id uint, req *models.UpdateGameRequest) (*models.GameResponse, error) {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		game.Name = *req.Name
	}
	if err := s.gameRepo.Update(game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}
	return s.toGameResponse(game), nil
}

func (s *gameService) DeleteGame(id uint) error {
	if _, err := s.gameRepo.GetByID(id); err != nil {
		return err
	}
	return s.gameRepo.Delete(id)
}

func (s *gameService) toGameResponse(game *models.Game) *models.GameResponse {
	return &models.GameResponse{
		ID:        game.ID,
		Name:      game.Name,
		CreatedAt: game.CreatedAt,
		UpdatedAt: game.UpdatedAt,
	}
}

func (s *gameService) SearchGames(query_string string, limit, offset int) ([]*models.GameResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	games, err := s.gameRepo.Search(query_string, limit, offset)
	if err != nil {
		return nil, err
	}
	responses := make([]*models.GameResponse, len(games))
	for i, game := range games {
		responses[i] = s.toGameResponse(game)
	}
	return responses, nil
}
