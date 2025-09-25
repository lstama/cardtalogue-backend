package service

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
)

type PackService interface {
	CreatePack(req *models.CreatePackRequest) (*models.PackResponse, error)
	GetPack(id uint) (*models.PackResponse, error)
	GetPacks(limit, offset int) ([]*models.PackResponse, int64, error)
	UpdatePack(id uint, req *models.UpdatePackRequest) (*models.PackResponse, error)
	DeletePack(id uint) error
	SearchPacks(game_id int, language_id int, query_string string, limit, offset int) ([]*models.PackResponse, error)
}

type packService struct {
	packRepo repository.PackRepository
}

func NewPackService(packRepo repository.PackRepository) PackService {
	return &packService{packRepo: packRepo}
}

func (s *packService) CreatePack(req *models.CreatePackRequest) (*models.PackResponse, error) {
	// Set default status if not provided

	pack := &models.Pack{
		Code:        req.Code,
		Name:        req.Name,
		ImageURL:    req.ImageURL,
		Description: req.Description,
		GameID:      req.GameID,
		LanguageID:  req.LanguageID,
	}

	if err := s.packRepo.Create(pack); err != nil {
		return nil, fmt.Errorf("failed to create pack: %w", err)
	}

	return s.toPackResponse(pack), nil
}

func (s *packService) GetPack(id uint) (*models.PackResponse, error) {
	pack, err := s.packRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toPackResponse(pack), nil
}

func (s *packService) GetPacks(limit, offset int) ([]*models.PackResponse, int64, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	packs, err := s.packRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.packRepo.Count()
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.PackResponse, len(packs))
	for i, pack := range packs {
		responses[i] = s.toPackResponse(pack)
	}

	return responses, total, nil
}

func (s *packService) UpdatePack(id uint, req *models.UpdatePackRequest) (*models.PackResponse, error) {
	pack, err := s.packRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Code != nil {
		pack.Code = *req.Code
	}
	if req.Name != nil {
		pack.Name = *req.Name
	}
	if req.ImageURL != nil {
		pack.ImageURL = *req.ImageURL
	}
	if req.GameID != nil {
		pack.GameID = *req.GameID
	}
	if req.LanguageID != nil {
		pack.LanguageID = *req.LanguageID
	}

	if err := s.packRepo.Update(pack); err != nil {
		return nil, fmt.Errorf("failed to update pack: %w", err)
	}

	return s.toPackResponse(pack), nil
}

func (s *packService) DeletePack(id uint) error {
	// Check if pack exists
	if _, err := s.packRepo.GetByID(id); err != nil {
		return err
	}

	return s.packRepo.Delete(id)
}

func (s *packService) toPackResponse(pack *models.Pack) *models.PackResponse {
	return &models.PackResponse{
		ID:           pack.ID,
		Code:         pack.Code,
		Name:         pack.Name,
		ImageURL:     pack.ImageURL,
		Description:  pack.Description,
		GameID:       pack.GameID,
		GameName:     pack.Game.Name,
		LanguageID:   pack.LanguageID,
		LanguageName: pack.Language.Name,
		LanguageCode: pack.Language.Code,
		CreatedAt:    pack.CreatedAt,
		UpdatedAt:    pack.UpdatedAt,
	}
}

func (s *packService) SearchPacks(game_id int, language_id int, query_string string, limit, offset int) ([]*models.PackResponse, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	packs, err := s.packRepo.Search(game_id, language_id, query_string, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.PackResponse, len(packs))
	for i, pack := range packs {
		responses[i] = s.toPackResponse(pack)
	}

	return responses, nil
}
