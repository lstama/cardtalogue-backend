package models

import (
	"time"
)

type Media struct {
	ID          uint      `json:"id" gorm:"column:id;primaryKey"`
	URL         string    `json:"url" gorm:"column:url;not null;default:https://picsum.photos/230/330.webp" validate:"required,min=1,max=255"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	Status      string    `json:"status" gorm:"column:status;not null;size:20;default:pending" validate:"required,oneof=pending attached archived disabled"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
	Provider    string    `json:"provider" gorm:"column:provider;not null"`
}

type CreateMediaRequest struct {
	URL         string `json:"url" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"oneof=pending attached archived disabled"`
	Provider    string `json:"provider" validate:"required,min=1,max=255"`
}

type UpdateMediaRequest struct {
	URL         *string `json:"url,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=pending attached archived disabled"`
	Provider    *string `json:"provider,omitempty" validate:"omitempty,min=1,max=255"`
}

type MediaResponse struct {
	ID          uint      `json:"id"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Provider    string    `json:"provider"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
