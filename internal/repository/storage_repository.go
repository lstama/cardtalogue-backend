package repository

import (
	"fmt"
	"strings"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"gorm.io/gorm"
)

type StorageRepository interface {
	Create(storage *models.Storage) error
	GetByID(id uint) (*models.Storage, error)
	GetAll(limit, offset int) ([]*models.Storage, error)
	Update(storage *models.Storage) error
	Delete(id uint) error
	Count() (int64, error)
	Search(query_string string, limit, offset int) ([]*models.Storage, int64, error)
}

type storageRepository struct {
	db *gorm.DB
}

func NewStorageRepository(db *gorm.DB) StorageRepository {
	return &storageRepository{db: db}
}

func (r *storageRepository) Create(storage *models.Storage) error {
	if err := r.db.Create(storage).Error; err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	return nil
}

func (r *storageRepository) GetByID(id uint) (*models.Storage, error) {
	var storage models.Storage
	if err := r.db.First(&storage, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("storage not found")
		}
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}
	return &storage, nil
}

func (r *storageRepository) GetAll(limit, offset int) ([]*models.Storage, error) {
	var storages []*models.Storage
	query := r.db.Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&storages).Error; err != nil {
		return nil, fmt.Errorf("failed to get storages: %w", err)
	}
	return storages, nil
}

func (r *storageRepository) Update(storage *models.Storage) error {
	if err := r.db.Save(storage).Error; err != nil {
		return fmt.Errorf("failed to update storage: %w", err)
	}
	return nil
}

func (r *storageRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Storage{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete storage: %w", err)
	}
	return nil
}

func (r *storageRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Storage{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count storages: %w", err)
	}
	return count, nil
}

func (r *storageRepository) Search(query_string string, limit, offset int) ([]*models.Storage, int64, error) {
	var storages []*models.Storage
	query := r.db.Model(&models.Storage{})
	if query_string != "" {
		likeQuery := "%" + strings.ToLower(query_string) + "%"
		query = query.Where("lowercase_name LIKE ?", likeQuery)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count storages: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&storages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search storages: %w", err)
	}
	return storages, total, nil
}
