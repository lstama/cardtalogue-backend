package service

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
)

type StorageService interface {
	CreateStorage(req *models.CreateStorageRequest) (*models.StorageResponse, error)
	GetStorage(id uint) (*models.StorageResponse, error)
	GetStorages(limit, offset int) ([]*models.StorageResponse, int64, error)
	UpdateStorage(id uint, req *models.UpdateStorageRequest) (*models.StorageResponse, error)
	DeleteStorage(id uint) error
	SearchStorages(query_string string, limit, offset int) ([]*models.StorageResponse, int64, error)
}

type storageService struct {
	storageRepo repository.StorageRepository
}

func NewStorageService(storageRepo repository.StorageRepository) StorageService {
	return &storageService{storageRepo: storageRepo}
}

func (s *storageService) CreateStorage(req *models.CreateStorageRequest) (*models.StorageResponse, error) {
	storage := &models.Storage{
		Name: req.Name,
		Code: req.Code,
	}
	if err := s.storageRepo.Create(storage); err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	return s.toStorageResponse(storage), nil
}

func (s *storageService) GetStorage(id uint) (*models.StorageResponse, error) {
	storage, err := s.storageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toStorageResponse(storage), nil
}

func (s *storageService) GetStorages(limit, offset int) ([]*models.StorageResponse, int64, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	storages, err := s.storageRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.storageRepo.Count()
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*models.StorageResponse, len(storages))
	for i, storage := range storages {
		responses[i] = s.toStorageResponse(storage)
	}
	return responses, total, nil
}

func (s *storageService) UpdateStorage(id uint, req *models.UpdateStorageRequest) (*models.StorageResponse, error) {
	storage, err := s.storageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		storage.Name = *req.Name
	}
	if req.Code != nil {
		storage.Code = *req.Code
	}
	if err := s.storageRepo.Update(storage); err != nil {
		return nil, fmt.Errorf("failed to update storage: %w", err)
	}
	return s.toStorageResponse(storage), nil
}

func (s *storageService) DeleteStorage(id uint) error {
	if _, err := s.storageRepo.GetByID(id); err != nil {
		return err
	}
	return s.storageRepo.Delete(id)
}

func (s *storageService) toStorageResponse(storage *models.Storage) *models.StorageResponse {
	return &models.StorageResponse{
		ID:          storage.ID,
		Code:        storage.Code,
		Name:        storage.Name,
		Description: storage.Description,
		CreatedAt:   storage.CreatedAt,
		UpdatedAt:   storage.UpdatedAt,
	}
}

func (s *storageService) SearchStorages(query_string string, limit, offset int) ([]*models.StorageResponse, int64, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	storages, total, err := s.storageRepo.Search(query_string, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*models.StorageResponse, len(storages))
	for i, storage := range storages {
		responses[i] = s.toStorageResponse(storage)
	}
	return responses, total, nil
}
