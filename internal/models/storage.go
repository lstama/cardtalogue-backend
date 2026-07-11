package models

import (
	"time"
)

type Storage struct {
	ID            uint      `json:"id" gorm:"column:id;primaryKey"`
	Code          string    `json:"code" gorm:"column:code;not null;size:255" validate:"required,min=1,max=255"`
	LowercaseCode string    `json:"lowercase_code" gorm:"column:lowercase_code;not null;size:255;index"` // for case-insensitive search
	Name          string    `json:"name" gorm:"column:name;not null;size:255" validate:"required,min=1,max=255"`
	LowercaseName string    `json:"lowercase_name" gorm:"column:lowercase_name;not null;size:255;index"` // for case-insensitive search
	Description   string    `json:"description" gorm:"column:description;type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`
}

type CreateStorageRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Code        string `json:"code" validate:"required,min=1,max=255"`
	Description string `json:"description"`
}

type UpdateStorageRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Code        *string `json:"code,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
}

type StorageResponse struct {
	ID          uint      `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
