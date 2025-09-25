package repository

import (
	"fmt"

	"github.com/lstama/cardtalogue-backend/internal/models"

	"gorm.io/gorm"
)

type CardRepository interface {
	Create(card *models.Card) error
	GetByID(id uint) (*models.Card, error)
	GetAll(limit, offset int) ([]*models.Card, error)
	Update(card *models.Card) error
	Delete(id uint) error
	Count() (int64, error)
}

type cardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) Create(card *models.Card) error {
	if err := r.db.Create(card).Error; err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}
	return nil
}

func (r *cardRepository) GetByID(id uint) (*models.Card, error) {
	var card models.Card
	if err := r.db.Preload("Game").Preload("Language").Preload("Pack").First(&card, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("card not found")
		}
		return nil, fmt.Errorf("failed to get card: %w", err)
	}
	return &card, nil
}

func (r *cardRepository) GetAll(limit, offset int) ([]*models.Card, error) {
	var cards []*models.Card
	query := r.db.Preload("Game").Preload("Language").Preload("Pack").Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&cards).Error; err != nil {
		return nil, fmt.Errorf("failed to get cards: %w", err)
	}
	return cards, nil
}

func (r *cardRepository) Update(card *models.Card) error {
	if err := r.db.Save(card).Error; err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}
	return nil
}

func (r *cardRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Card{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}
	return nil
}

func (r *cardRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Card{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count cards: %w", err)
	}
	return count, nil
}
