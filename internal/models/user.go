package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"column:id;primaryKey"`
	Name      string         `json:"name" gorm:"column:name;not null;size:100" validate:"required,min=2,max=100"`
	Email     string         `json:"email" gorm:"column:email;uniqueIndex;not null;size:255" validate:"required,email"`
	Password  string         `json:"-" gorm:"column:password;not null;size:255"`
	Role      string         `json:"role" gorm:"column:role;not null;size:50;default:'editor'" validate:"required,oneof=admin editor"`
	Age       int            `json:"age" gorm:"column:age;not null" validate:"required,min=1,max=150"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index"`
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Role     string `json:"role" validate:"required,oneof=admin editor"`
	Age      int    `json:"age" validate:"required,min=1,max=150"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Role  *string `json:"role,omitempty" validate:"omitempty,oneof=admin editor"`
	Age   *int    `json:"age,omitempty" validate:"omitempty,min=1,max=150"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// type LoginResponseNoToken struct {
// 	User          UserResponse `json:"user"`
// 	Authenticated bool         `json:"authenticated"`
// }

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6,max=100"`
}
