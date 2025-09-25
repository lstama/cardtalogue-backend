package repository

import (
	"fmt"
	"strings"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"gorm.io/gorm"
)

type LanguageRepository interface {
	Create(language *models.Language) error
	GetByID(id uint) (*models.Language, error)
	GetAll(limit, offset int) ([]*models.Language, error)
	Update(language *models.Language) error
	Delete(id uint) error
	Count() (int64, error)
	Search(query_string string, limit, offset int) ([]*models.Language, error)
}

type languageRepository struct {
	db *gorm.DB
}

func NewLanguageRepository(db *gorm.DB) LanguageRepository {
	return &languageRepository{db: db}
}

func (r *languageRepository) Create(language *models.Language) error {
	if err := r.db.Create(language).Error; err != nil {
		return fmt.Errorf("failed to create language: %w", err)
	}
	return nil
}

func (r *languageRepository) GetByID(id uint) (*models.Language, error) {
	var language models.Language
	if err := r.db.First(&language, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("language not found")
		}
		return nil, fmt.Errorf("failed to get language: %w", err)
	}
	return &language, nil
}

func (r *languageRepository) GetAll(limit, offset int) ([]*models.Language, error) {
	var languages []*models.Language
	query := r.db.Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&languages).Error; err != nil {
		return nil, fmt.Errorf("failed to get languages: %w", err)
	}
	return languages, nil
}

func (r *languageRepository) Update(language *models.Language) error {
	if err := r.db.Save(language).Error; err != nil {
		return fmt.Errorf("failed to update language: %w", err)
	}
	return nil
}

func (r *languageRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Language{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete language: %w", err)
	}
	return nil
}

func (r *languageRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Language{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count languages: %w", err)
	}
	return count, nil
}

func (r *languageRepository) Search(query_string string, limit, offset int) ([]*models.Language, error) {
	var languages []*models.Language
	query := r.db.Limit(limit).Offset(offset).Order("created_at DESC")

	query_string = strings.ToLower(query_string)

	if query_string != "" {
		likeQuery := fmt.Sprintf("%%%s%%", query_string)
		query = query.Where("lowercase_name LIKE ? OR lowercase_code LIKE ?", likeQuery, likeQuery)
	}

	if err := query.Find(&languages).Error; err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}
	return languages, nil
}
