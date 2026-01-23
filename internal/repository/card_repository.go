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
	GetByIDWithStatus(id uint, includeDisabled bool) (*models.Card, error)
	GetAll(limit, offset int) ([]*models.Card, error)
	GetAllWithStatus(limit, offset int, includeDisabled bool) ([]*models.Card, error)
	Update(card *models.Card) error
	Delete(id uint) error
	Count() (int64, error)
	CountWithStatus(includeDisabled bool) (int64, error)
	Search(game_id int, language_id int, pack_id int, query_string string, limit, offset int) ([]*models.Card, int64, error)
	SearchWithStatus(game_id int, language_id int, pack_id int, query_string string, limit, offset int, includeDisabled bool) ([]*models.Card, int64, error)
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
	return r.GetByIDWithStatus(id, false)
}

func (r *cardRepository) GetByIDWithStatus(id uint, includeDisabled bool) (*models.Card, error) {
	var card models.Card
	query := r.db.Preload("Game").Preload("Language").Preload("Pack")
	if !includeDisabled {
		query = query.Where("status = ?", "enabled")
	}
	if err := query.First(&card, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("card not found")
		}
		return nil, fmt.Errorf("failed to get card: %w", err)
	}
	return &card, nil
}

func (r *cardRepository) GetAll(limit, offset int) ([]*models.Card, error) {
	return r.GetAllWithStatus(limit, offset, false)
}

func (r *cardRepository) GetAllWithStatus(limit, offset int, includeDisabled bool) ([]*models.Card, error) {
	var cards []*models.Card
	query := r.db.Preload("Game").Preload("Language").Preload("Pack").Limit(limit).Offset(offset).Order("created_at DESC")
	if !includeDisabled {
		query = query.Where("status = ?", "enabled")
	}
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
	return r.CountWithStatus(false)
}

func (r *cardRepository) CountWithStatus(includeDisabled bool) (int64, error) {
	var count int64
	query := r.db.Model(&models.Card{})
	if !includeDisabled {
		query = query.Where("status = ?", "enabled")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count cards: %w", err)
	}
	return count, nil
}

func (r *cardRepository) Search(gameID, languageID, packID int, queryString string, limit, offset int) ([]*models.Card, int64, error) {
	return r.SearchWithStatus(gameID, languageID, packID, queryString, limit, offset, false)
}

func (r *cardRepository) SearchWithStatus(gameID, languageID, packID int, queryString string, limit, offset int, includeDisabled bool) ([]*models.Card, int64, error) {
	var cards []*models.Card
	var total int64

	// Base query (filters only, no preload, no limit/offset/order)
	baseQuery := r.db.Model(&models.Card{})

	if gameID > 0 {
		baseQuery = baseQuery.Where("game_id = ?", gameID)
	}
	if languageID > 0 {
		baseQuery = baseQuery.Where("language_id = ?", languageID)
	}
	if packID > 0 {
		baseQuery = baseQuery.Where("pack_id = ?", packID)
	}
	if queryString != "" {
		queryString = strings.ToLower(queryString)
		likeQuery := fmt.Sprintf("%%%s%%", queryString)
		baseQuery = baseQuery.Where("lowercase_name LIKE ? OR lowercase_code LIKE ?", likeQuery, likeQuery)
	}
	if !includeDisabled {
		baseQuery = baseQuery.Where("status = ?", "enabled")
	}

	// Count (on base query only)
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count searched cards: %w", err)
	}

	// Fetch data with preload + pagination + order
	findQuery := baseQuery.Preload("Game").Preload("Language").Preload("Pack").Preload("Rarity").Preload("Storage").
		Limit(limit).Offset(offset).Order("created_at DESC")

	if err := findQuery.Find(&cards).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search cards: %w", err)
	}

	return cards, total, nil
}
