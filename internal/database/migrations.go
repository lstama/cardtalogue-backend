package database

import (
	"errors"
	"fmt"
	"log"
	"strings"

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

	if err := CreateDefaultSeedData(db); err != nil {
		return fmt.Errorf("failed to create default seed data: %w", err)
	}

	log.Println("GORM auto-migration completed successfully")
	return nil
}

func CreateDefaultSeedData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		defaultName := "Default"
		defaultCode := "DF"

		defaultGame, err := ensureDefaultGame(tx, defaultName)
		if err != nil {
			return err
		}

		defaultLanguage, err := ensureDefaultLanguage(tx, defaultName, defaultCode)
		if err != nil {
			return err
		}

		if _, err := ensureDefaultPack(tx, defaultName, defaultCode, defaultGame.ID, defaultLanguage.ID); err != nil {
			return err
		}

		if _, err := ensureDefaultRarity(tx, defaultName, defaultCode); err != nil {
			return err
		}

		if _, err := ensureDefaultStorage(tx, defaultName, defaultCode); err != nil {
			return err
		}

		return nil
	})
}

func ensureDefaultGame(tx *gorm.DB, defaultName string) (*models.Game, error) {
	var existing models.Game
	err := tx.Where("name = ? OR lowercase_name = ?", defaultName, strings.ToLower(defaultName)).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	game := models.Game{
		Name:          defaultName,
		LowercaseName: strings.ToLower(defaultName),
	}

	if err := tx.Create(&game).Error; err != nil {
		return nil, err
	}

	return &game, nil
}

func ensureDefaultLanguage(tx *gorm.DB, defaultName, defaultCode string) (*models.Language, error) {
	var existing models.Language
	err := tx.Where("code = ? OR lowercase_code = ?", defaultCode, strings.ToLower(defaultCode)).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	language := models.Language{
		Code:          defaultCode,
		LowercaseCode: strings.ToLower(defaultCode),
		Name:          defaultName,
		LowercaseName: strings.ToLower(defaultName),
	}

	if err := tx.Create(&language).Error; err != nil {
		return nil, err
	}

	return &language, nil
}

func ensureDefaultPack(tx *gorm.DB, defaultName, defaultCode string, gameID, languageID uint) (*models.Pack, error) {
	var existing models.Pack
	err := tx.Where("code = ? OR lowercase_code = ?", defaultCode, strings.ToLower(defaultCode)).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	pack := models.Pack{
		Code:          defaultCode,
		LowercaseCode: strings.ToLower(defaultCode),
		Name:          defaultName,
		LowercaseName: strings.ToLower(defaultName),
		GameID:        gameID,
		LanguageID:    languageID,
	}

	if err := tx.Create(&pack).Error; err != nil {
		return nil, err
	}

	return &pack, nil
}

func ensureDefaultRarity(tx *gorm.DB, defaultName, defaultCode string) (*models.Rarity, error) {
	var existing models.Rarity
	err := tx.Where("code = ? OR lowercase_code = ?", defaultCode, strings.ToLower(defaultCode)).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	rarity := models.Rarity{
		Code:          defaultCode,
		LowercaseCode: strings.ToLower(defaultCode),
		Name:          defaultName,
		LowercaseName: strings.ToLower(defaultName),
	}

	if err := tx.Create(&rarity).Error; err != nil {
		return nil, err
	}

	return &rarity, nil
}

func ensureDefaultStorage(tx *gorm.DB, defaultName, defaultCode string) (*models.Storage, error) {
	var existing models.Storage
	err := tx.Where("code = ? OR lowercase_code = ?", defaultCode, strings.ToLower(defaultCode)).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	storage := models.Storage{
		Code:          defaultCode,
		LowercaseCode: strings.ToLower(defaultCode),
		Name:          defaultName,
		LowercaseName: strings.ToLower(defaultName),
	}

	if err := tx.Create(&storage).Error; err != nil {
		return nil, err
	}

	return &storage, nil
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
