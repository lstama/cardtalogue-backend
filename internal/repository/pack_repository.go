package repository

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"

	"gorm.io/gorm"
)

type PackRepository interface {
	Create(pack *models.Pack) error
	GetByID(id uint) (*models.Pack, error)
	GetAll(limit, offset int) ([]*models.Pack, error)
	Update(pack *models.Pack) error
	Delete(id uint) error
	Count() (int64, error)
}

type packRepository struct {
	db *gorm.DB
}

func NewPackRepository(db *gorm.DB) PackRepository {
	return &packRepository{db: db}
}

func (r *packRepository) Create(pack *models.Pack) error {
	if err := r.db.Create(pack).Error; err != nil {
		return fmt.Errorf("failed to create pack: %w", err)
	}
	return nil
}

func (r *packRepository) GetByID(id uint) (*models.Pack, error) {
	var pack models.Pack
	if err := r.db.Preload("Game").Preload("Language").First(&pack, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pack not found")
		}
		return nil, fmt.Errorf("failed to get pack: %w", err)
	}
	return &pack, nil
}

func (r *packRepository) GetAll(limit, offset int) ([]*models.Pack, error) {
	var packs []*models.Pack
	query := r.db.Preload("Game").Preload("Language").Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&packs).Error; err != nil {
		return nil, fmt.Errorf("failed to get packs: %w", err)
	}
	return packs, nil
}

func (r *packRepository) Update(pack *models.Pack) error {
	if err := r.db.Save(pack).Error; err != nil {
		return fmt.Errorf("failed to update pack: %w", err)
	}
	return nil
}

func (r *packRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Pack{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete pack: %w", err)
	}
	return nil
}

func (r *packRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Pack{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count packs: %w", err)
	}
	return count, nil
}
