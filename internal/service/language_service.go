package service

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"github.com/lstama/cardtalogue-backend/internal/repository"
)

type LanguageService interface {
	CreateLanguage(req *models.CreateLanguageRequest) (*models.LanguageResponse, error)
	GetLanguage(id uint) (*models.LanguageResponse, error)
	GetLanguages(limit, offset int) ([]*models.LanguageResponse, int64, error)
	UpdateLanguage(id uint, req *models.UpdateLanguageRequest) (*models.LanguageResponse, error)
	DeleteLanguage(id uint) error
	SearchLanguages(query_string string, limit, offset int) ([]*models.LanguageResponse, int64, error)
}

type languageService struct {
	languageRepo repository.LanguageRepository
}

func NewLanguageService(languageRepo repository.LanguageRepository) LanguageService {
	return &languageService{languageRepo: languageRepo}
}

func (s *languageService) CreateLanguage(req *models.CreateLanguageRequest) (*models.LanguageResponse, error) {
	language := &models.Language{
		Name: req.Name,
		Code: req.Code,
	}
	if err := s.languageRepo.Create(language); err != nil {
		return nil, fmt.Errorf("failed to create language: %w", err)
	}
	return s.toLanguageResponse(language), nil
}

func (s *languageService) GetLanguage(id uint) (*models.LanguageResponse, error) {
	language, err := s.languageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toLanguageResponse(language), nil
}

func (s *languageService) GetLanguages(limit, offset int) ([]*models.LanguageResponse, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	languages, err := s.languageRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.languageRepo.Count()
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*models.LanguageResponse, len(languages))
	for i, language := range languages {
		responses[i] = s.toLanguageResponse(language)
	}
	return responses, total, nil
}

func (s *languageService) UpdateLanguage(id uint, req *models.UpdateLanguageRequest) (*models.LanguageResponse, error) {
	language, err := s.languageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		language.Name = *req.Name
	}
	if req.Code != nil {
		language.Code = *req.Code
	}
	if err := s.languageRepo.Update(language); err != nil {
		return nil, fmt.Errorf("failed to update language: %w", err)
	}
	return s.toLanguageResponse(language), nil
}

func (s *languageService) DeleteLanguage(id uint) error {
	if _, err := s.languageRepo.GetByID(id); err != nil {
		return err
	}
	return s.languageRepo.Delete(id)
}

func (s *languageService) toLanguageResponse(language *models.Language) *models.LanguageResponse {
	return &models.LanguageResponse{
		ID:        language.ID,
		Code:      language.Code,
		Name:      language.Name,
		CreatedAt: language.CreatedAt,
		UpdatedAt: language.UpdatedAt,
	}
}

func (s *languageService) SearchLanguages(query_string string, limit, offset int) ([]*models.LanguageResponse, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	languages, total, err := s.languageRepo.Search(query_string, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*models.LanguageResponse, len(languages))
	for i, language := range languages {
		responses[i] = s.toLanguageResponse(language)
	}
	return responses, total, nil
}
