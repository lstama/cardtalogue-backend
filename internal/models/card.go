package models

import (
	"time"

	"gorm.io/gorm"
)

type Card struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Code            string         `json:"code" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Name            string         `json:"name" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Price           uint           `json:"price" gorm:"not null;default:100" validate:"min=0,max=1000000000"`
	PriceMultiplier uint           `json:"price_multiplier" gorm:"not null;default:120" validate:"min=1,max=100000"`
	Quantity        uint           `json:"quantity" gorm:"not null;default:1" validate:"min=0,max=1000000"`
	ImageURL        string         `json:"image_url" gorm:"not null;default:https://picsum.photos/230/330.webp" validate:"required,min=1,max=255"`
	Description     string         `json:"description" gorm:"type:text"`
	Status          string         `json:"status" gorm:"not null;size:20;default:enabled" validate:"required,oneof=enabled disabled"`
	ForSale         string         `json:"for_sale" gorm:"not null;size:20;default:'yes'" validate:"required,oneof=yes no"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	GameID          uint           `json:"game_id" gorm:"not null"`
	Game            Game           `json:"game" gorm:"foreignKey:GameID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	LanguageID      uint           `json:"language_id,omitempty" gorm:"not null"`
	Language        Language       `json:"language,omitempty" gorm:"foreignKey:LanguageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	PackID          uint           `json:"pack_id,omitempty" gorm:"not null"`
	Pack            Pack           `json:"pack,omitempty" gorm:"foreignKey:PackID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RarityID        uint           `json:"rarity_id,omitempty" gorm:"not null;default:1"`
	Rarity          Rarity         `json:"rarity,omitempty" gorm:"foreignKey:RarityID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	StorageID       uint           `json:"storage_id,omitempty"`
	Storage         Storage        `json:"storage,omitempty" gorm:"foreignKey:StorageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type CreateCardRequest struct {
	Code            string `json:"code" validate:"required,min=1,max=255"`
	Name            string `json:"name" validate:"required,min=1,max=255"`
	Price           uint   `json:"price" validate:"min=0,max=1000000000"`
	PriceMultiplier uint   `json:"price_multiplier" validate:"min=1,max=100000"`
	Quantity        uint   `json:"quantity" validate:"min=0,max=1000000"`
	ImageURL        string `json:"image_url" validate:"required,min=1,max=255"`
	Description     string `json:"description"`
	Status          string `json:"status" validate:"oneof=enabled disabled"`
	ForSale         string `json:"for_sale" validate:"required,oneof=yes no"`
	GameID          uint   `json:"game_id" validate:"required,min=1"`
	LanguageID      uint   `json:"language_id,omitempty" validate:"omitempty,min=1"`
	PackID          uint   `json:"pack_id,omitempty" validate:"omitempty,min=1"`
	RarityID        uint   `json:"rarity_id,omitempty" validate:"omitempty,min=1"`
	StorageID       uint   `json:"storage_id,omitempty" validate:"omitempty,min=1"`
}

type UpdateCardRequest struct {
	Code            *string `json:"code,omitempty" validate:"omitempty,min=1,max=255"`
	Name            *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Price           *uint   `json:"price,omitempty" validate:"omitempty,min=0,max=1000000000"`
	PriceMultiplier *uint   `json:"price_multiplier,omitempty" validate:"omitempty,min=1,max=100000"`
	Quantity        *uint   `json:"quantity,omitempty" validate:"omitempty,min=0,max=1000000"`
	ImageURL        *string `json:"image_url,omitempty" validate:"omitempty,min=1,max=255"`
	Description     *string `json:"description,omitempty"`
	Status          *string `json:"status,omitempty" validate:"omitempty,oneof=enabled disabled"`
	ForSale         *string `json:"for_sale,omitempty" validate:"omitempty,oneof=yes no"`
	GameID          *uint   `json:"game_id,omitempty" validate:"omitempty,min=1"`
	LanguageID      *uint   `json:"language_id,omitempty" validate:"omitempty,min=1"`
	PackID          *uint   `json:"pack_id,omitempty" validate:"omitempty,min=1"`
	RarityID        *uint   `json:"rarity_id,omitempty" validate:"omitempty,min=1"`
	StorageID       *uint   `json:"storage_id,omitempty" validate:"omitempty,min=1"`
}

type CardResponse struct {
	ID              uint      `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Price           uint      `json:"price"`
	PriceMultiplier uint      `json:"price_multiplier"`
	Quantity        uint      `json:"quantity"`
	ImageURL        string    `json:"image_url"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	ForSale         string    `json:"for_sale"`
	GameID          uint      `json:"game_id"`
	GameName        string    `json:"game_name"`
	LanguageID      uint      `json:"language_id,omitempty"`
	LanguageName    string    `json:"language_name,omitempty"`
	LanguageCode    string    `json:"language_code,omitempty"`
	PackID          uint      `json:"pack_id,omitempty"`
	PackName        string    `json:"pack_name,omitempty"`
	PackCode        string    `json:"pack_code,omitempty"`
	RarityID        uint      `json:"rarity_id,omitempty"`
	RarityName      string    `json:"rarity_name,omitempty"`
	RarityCode      string    `json:"rarity_code,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// With sensitive info (storage details)
type CardProtectedResponse struct {
	ID              uint      `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Price           uint      `json:"price"`
	PriceMultiplier uint      `json:"price_multiplier"`
	Quantity        uint      `json:"quantity"`
	ImageURL        string    `json:"image_url"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	ForSale         string    `json:"for_sale"`
	GameID          uint      `json:"game_id"`
	GameName        string    `json:"game_name"`
	LanguageID      uint      `json:"language_id,omitempty"`
	LanguageName    string    `json:"language_name,omitempty"`
	LanguageCode    string    `json:"language_code,omitempty"`
	PackID          uint      `json:"pack_id,omitempty"`
	PackName        string    `json:"pack_name,omitempty"`
	PackCode        string    `json:"pack_code,omitempty"`
	RarityID        uint      `json:"rarity_id,omitempty"`
	RarityName      string    `json:"rarity_name,omitempty"`
	RarityCode      string    `json:"rarity_code,omitempty"`
	StorageID       uint      `json:"storage_id,omitempty"`
	StorageName     string    `json:"storage_name,omitempty"`
	StorageCode     string    `json:"storage_code,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
