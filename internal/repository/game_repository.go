package repository

import (
	"fmt"
	"strings"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"gorm.io/gorm"
)

type GameRepository interface {
	Create(game *models.Game) error
	GetByID(id uint) (*models.Game, error)
	GetAll(limit, offset int) ([]*models.Game, error)
	Update(game *models.Game) error
	Delete(id uint) error
	Count() (int64, error)
	Search(query_string string, limit, offset int) ([]*models.Game, error)
}

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(game *models.Game) error {
	if err := r.db.Create(game).Error; err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}
	return nil
}

func (r *gameRepository) GetByID(id uint) (*models.Game, error) {
	var game models.Game
	if err := r.db.First(&game, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	return &game, nil
}

func (r *gameRepository) GetAll(limit, offset int) ([]*models.Game, error) {
	var games []*models.Game
	query := r.db.Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&games).Error; err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}
	return games, nil
}

func (r *gameRepository) Update(game *models.Game) error {
	if err := r.db.Save(game).Error; err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}
	return nil
}

func (r *gameRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Game{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}
	return nil
}

func (r *gameRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Game{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count games: %w", err)
	}
	return count, nil
}

func (r *gameRepository) Search(query_string string, limit, offset int) ([]*models.Game, error) {
	var games []*models.Game
	query := r.db.Limit(limit).Offset(offset).Order("created_at DESC")

	query_string = strings.ToLower(query_string)

	if query_string != "" {
		likeQuery := fmt.Sprintf("%%%s%%", query_string)
		query = query.Where("lowercase_name LIKE ?", likeQuery)
	}

	if err := query.Find(&games).Error; err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}
	return games, nil
}
