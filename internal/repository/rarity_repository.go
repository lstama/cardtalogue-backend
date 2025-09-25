package repository

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"gorm.io/gorm"
)

type RarityRepository interface {
	Create(rarity *models.Rarity) error
	GetByID(id uint) (*models.Rarity, error)
	GetAll(limit, offset int) ([]*models.Rarity, error)
	Update(rarity *models.Rarity) error
	Delete(id uint) error
	Count() (int64, error)
}

type rarityRepository struct {
	db *gorm.DB
}

func NewRarityRepository(db *gorm.DB) RarityRepository {
	return &rarityRepository{db: db}
}

func (r *rarityRepository) Create(rarity *models.Rarity) error {
	if err := r.db.Create(rarity).Error; err != nil {
		return fmt.Errorf("failed to create rarity: %w", err)
	}
	return nil
}

func (r *rarityRepository) GetByID(id uint) (*models.Rarity, error) {
	var rarity models.Rarity
	if err := r.db.First(&rarity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("rarity not found")
		}
		return nil, fmt.Errorf("failed to get rarity: %w", err)
	}
	return &rarity, nil
}

func (r *rarityRepository) GetAll(limit, offset int) ([]*models.Rarity, error) {
	var rarities []*models.Rarity
	query := r.db.Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&rarities).Error; err != nil {
		return nil, fmt.Errorf("failed to get rarities: %w", err)
	}
	return rarities, nil
}

func (r *rarityRepository) Update(rarity *models.Rarity) error {
	if err := r.db.Save(rarity).Error; err != nil {
		return fmt.Errorf("failed to update rarity: %w", err)
	}
	return nil
}

func (r *rarityRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Rarity{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete rarity: %w", err)
	}
	return nil
}

func (r *rarityRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Rarity{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count rarities: %w", err)
	}
	return count, nil
}
