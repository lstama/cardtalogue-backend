package models

import (
	"time"
)

type Pack struct {
	ID            uint      `json:"id" gorm:"column:id;primaryKey"`
	Code          string    `json:"code" gorm:"column:code;not null;size:255" validate:"required,min=1,max=255"`
	LowercaseCode string    `json:"lowercase_code" gorm:"column:lowercase_code;not null;size:255;index"` // for case-insensitive search
	Name          string    `json:"name" gorm:"column:name;not null;size:255" validate:"required,min=1,max=255"`
	LowercaseName string    `json:"lowercase_name" gorm:"column:lowercase_name;not null;size:255;index"` // for case-insensitive search
	ImageURL      string    `json:"image_url" gorm:"column:image_url;not null;default:https://picsum.photos/230/330.webp" validate:"required,min=1,max=255"`
	Description   string    `json:"description" gorm:"column:description;type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`
	GameID        uint      `json:"game_id" gorm:"column:game_id;not null"`
	Game          Game      `json:"game" gorm:"column:game_id;foreignKey:GameID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	LanguageID    uint      `json:"language_id,omitempty" gorm:"column:language_id;not null"`
	Language      Language  `json:"language,omitempty" gorm:"column:language_id;foreignKey:LanguageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type CreatePackRequest struct {
	Code        string `json:"code" validate:"required,min=1,max=255"`
	Name        string `json:"name" validate:"required,min=1,max=255"`
	ImageURL    string `json:"image_url" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	GameID      uint   `json:"game_id" validate:"required,min=1"`
	LanguageID  uint   `json:"language_id,omitempty" validate:"omitempty,min=1"`
}

type UpdatePackRequest struct {
	Code       *string `json:"code,omitempty" validate:"omitempty,min=1,max=255"`
	Name       *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	ImageURL   *string `json:"image_url,omitempty" validate:"omitempty,min=1,max=255"`
	Status     *string `json:"status,omitempty" validate:"omitempty,oneof=enabled disabled"`
	GameID     *uint   `json:"game_id,omitempty" validate:"omitempty,min=1"`
	LanguageID *uint   `json:"language_id,omitempty" validate:"omitempty,min=1"`
}

type PackResponse struct {
	ID           uint      `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	ImageURL     string    `json:"image_url"`
	Description  string    `json:"description"`
	GameID       uint      `json:"game_id"`
	GameName     string    `json:"game_name"`
	LanguageID   uint      `json:"language_id,omitempty"`
	LanguageName string    `json:"language_name,omitempty"`
	LanguageCode string    `json:"language_code,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
