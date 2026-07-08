package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewConnection(databaseURL, databaseDriver string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch databaseDriver {
	case "postgres":
		dialector = postgres.Open(databaseURL)
	case "mysql":
		dialector = mysql.Open(databaseURL)
	case "sqlite":
		dialector = sqlite.Open(databaseURL)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", databaseDriver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
