package service

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
)

type RarityService interface {
	CreateRarity(req *models.CreateRarityRequest) (*models.RarityResponse, error)
	GetRarity(id uint) (*models.RarityResponse, error)
	GetRarities(limit, offset int) ([]*models.RarityResponse, int64, error)
	UpdateRarity(id uint, req *models.UpdateRarityRequest) (*models.RarityResponse, error)
	DeleteRarity(id uint) error
}

type rarityService struct {
	rarityRepo repository.RarityRepository
}

func NewRarityService(rarityRepo repository.RarityRepository) RarityService {
	return &rarityService{rarityRepo: rarityRepo}
}

func (s *rarityService) CreateRarity(req *models.CreateRarityRequest) (*models.RarityResponse, error) {
	rarity := &models.Rarity{
		Name: req.Name,
		Code: req.Code,
	}
	if err := s.rarityRepo.Create(rarity); err != nil {
		return nil, fmt.Errorf("failed to create rarity: %w", err)
	}
	return s.toRarityResponse(rarity), nil
}

func (s *rarityService) GetRarity(id uint) (*models.RarityResponse, error) {
	rarity, err := s.rarityRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toRarityResponse(rarity), nil
}

func (s *rarityService) GetRarities(limit, offset int) ([]*models.RarityResponse, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	rarities, err := s.rarityRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.rarityRepo.Count()
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*models.RarityResponse, len(rarities))
	for i, rarity := range rarities {
		responses[i] = s.toRarityResponse(rarity)
	}
	return responses, total, nil
}

func (s *rarityService) UpdateRarity(id uint, req *models.UpdateRarityRequest) (*models.RarityResponse, error) {
	rarity, err := s.rarityRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		rarity.Name = *req.Name
	}
	if req.Code != nil {
		rarity.Code = *req.Code
	}
	if err := s.rarityRepo.Update(rarity); err != nil {
		return nil, fmt.Errorf("failed to update rarity: %w", err)
	}
	return s.toRarityResponse(rarity), nil
}

func (s *rarityService) DeleteRarity(id uint) error {
	if _, err := s.rarityRepo.GetByID(id); err != nil {
		return err
	}
	return s.rarityRepo.Delete(id)
}

func (s *rarityService) toRarityResponse(rarity *models.Rarity) *models.RarityResponse {
	return &models.RarityResponse{
		ID:        rarity.ID,
		Code:      rarity.Code,
		Name:      rarity.Name,
		CreatedAt: rarity.CreatedAt,
		UpdatedAt: rarity.UpdatedAt,
	}
}
