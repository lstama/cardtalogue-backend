package models

import (
	"time"
)

type Rarity struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Code          string    `json:"code" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	LowercaseCode string    `json:"lowercase_code" gorm:"not null;size:255;index"` // for case-insensitive search
	Name          string    `json:"name" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	LowercaseName string    `json:"lowercase_name" gorm:"not null;size:255;index"` // for case-insensitive search
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateRarityRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	Code string `json:"code" validate:"required,min=1,max=255"`
}

type UpdateRarityRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Code *string `json:"code,omitempty" validate:"omitempty,min=1,max=255"`
}

type RarityResponse struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
