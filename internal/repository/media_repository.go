package repository

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"

	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(media *models.Media) error
	GetByID(id uint) (*models.Media, error)
	GetAll(limit, offset int) ([]*models.Media, error)
	Update(media *models.Media) error
	Delete(id uint) error
	Count() (int64, error)
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(media *models.Media) error {
	if err := r.db.Create(media).Error; err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}
	return nil
}

func (r *mediaRepository) GetByID(id uint) (*models.Media, error) {
	var media models.Media
	if err := r.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("media not found")
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
	}
	return &media, nil
}

func (r *mediaRepository) GetAll(limit, offset int) ([]*models.Media, error) {
	var medias []*models.Media
	query := r.db.Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&medias).Error; err != nil {
		return nil, fmt.Errorf("failed to get medias: %w", err)
	}
	return medias, nil
}

func (r *mediaRepository) Update(media *models.Media) error {
	if err := r.db.Save(media).Error; err != nil {
		return fmt.Errorf("failed to update media: %w", err)
	}
	return nil
}

func (r *mediaRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Media{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}
	return nil
}

func (r *mediaRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Media{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count medias: %w", err)
	}
	return count, nil
}
