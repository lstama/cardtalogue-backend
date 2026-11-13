package models

import (
	"time"
)

type Media struct {
	ID            uint      `json:"id" gorm:"column:id;primaryKey"`
	URL           string    `json:"url" gorm:"column:url;not null;default:https://picsum.photos/230/330.webp" validate:"required,min=1,max=255"`
	Description   string    `json:"description" gorm:"column:description;type:text"`
	Status        string    `json:"status" gorm:"column:status;not null;size:20;default:pending" validate:"required,oneof=pending attached archived disabled"`
	Provider      string    `json:"provider" gorm:"column:provider;not null"`
	Type          string    `json:"type" gorm:"column:type;size:20"`
	FileName      string    `json:"file_name" gorm:"column:file_name;size:255"`
	FileThumbURL  string    `json:"file_thumb_url" gorm:"column:file_thumb_url;size:255"`
	FileSize      int64     `json:"file_size" gorm:"column:file_size"`
	IsDeleted     bool      `json:"is_deleted" gorm:"column:is_deleted;default:false"`
	IsPending     bool      `json:"is_pending" gorm:"column:is_pending;default:true"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`
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
	ID            uint      `json:"id"`
	URL           string    `json:"url"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Provider      string    `json:"provider"`
	Type          string    `json:"type"`
	FileName      string    `json:"file_name"`
	FileThumbURL  string    `json:"file_thumb_url"`
	FileSize      int64     `json:"file_size"`
	IsDeleted     bool      `json:"is_deleted"`
	IsPending     bool      `json:"is_pending"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
