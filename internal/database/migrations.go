package database

import (
	"fmt"
	"log"

	"github.com/lstama/cardtalogue-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RunMigrations(databaseURL, databaseDriver, migrationsPath string) error {
	_ = migrationsPath

	db, err := NewConnection(databaseURL, databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return AutoMigrate(db)
}

func RollbackMigrations(databaseURL, databaseDriver, migrationsPath string, steps int) error {
	_ = databaseURL
	_ = databaseDriver
	_ = migrationsPath
	_ = steps

	return fmt.Errorf("rollback is not supported when using GORM auto-migration")
}

func GetMigrationVersion(databaseURL, databaseDriver, migrationsPath string) (uint, bool, error) {
	_ = databaseURL
	_ = databaseDriver
	_ = migrationsPath

	return 0, false, nil
}

func AutoMigrate(db *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.Game{},
		&models.Language{},
		&models.Pack{},
		&models.Rarity{},
		&models.Storage{},
		&models.Media{},
		&models.Card{},
	}

	if err := db.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("failed to auto-migrate schema: %w", err)
	}

	log.Println("GORM auto-migration completed successfully")
	return nil
}

func CreateDefaultAdmin(db *gorm.DB, email, password string) error {
	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		log.Println("Default admin user already exists")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin := models.User{
		Name:     "Administrator",
		Email:    email,
		Password: string(hashedPassword),
		Role:     "admin",
		Age:      30, // Default age for admin user
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	log.Printf("Default admin user created with email: %s", email)
	return nil
}
