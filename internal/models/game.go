package models

import (
	"time"
)

type Game struct {
	ID            uint      `json:"id" gorm:"column:id;primaryKey"`
	Name          string    `json:"name" gorm:"column:name;not null;size:255" validate:"required,min=1,max=255"`
	LowercaseName string    `json:"lowercase_name" gorm:"column:lowercase_name;not null;size:255;index"` // for case-insensitive search
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`
}

type CreateGameRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type UpdateGameRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
}

type GameResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
