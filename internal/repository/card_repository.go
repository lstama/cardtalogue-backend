package repository

import (
	"fmt"
	"strings"

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
	Search(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.Card, error)
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

func (r *cardRepository) Search(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.Card, error) {
	var cards []*models.Card
	query := r.db.Preload("Game").Preload("Language").Preload("Pack").Limit(limit).Offset(offset).Order("created_at DESC")

	query_string = strings.ToLower(query_string)

	if game_id > 0 {
		query = query.Where("game_id = ?", game_id)
	}
	if language_id > 0 {
		query = query.Where("language_id = ?", language_id)
	}
	if pack_id > 0 {
		query = query.Where("pack_id = ?", pack_id)
	}
	if query_string != "" {
		likeQuery := fmt.Sprintf("%%%s%%", query_string)
		query = query.Where("lowercase_name LIKE ? OR lowercase_code LIKE ?", likeQuery, likeQuery)
	}

	if err := query.Find(&cards).Error; err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}
	return cards, nil
}
