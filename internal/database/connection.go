package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewConnection(databaseURL, databaseDriver string) (*gorm.DB, error) {
	return NewConnectionWithConn(databaseURL, databaseDriver, nil)
}

// NewConnectionWithConn opens gorm with an optional pre-built ConnPool. When
// conn is non-nil it is used directly (used to inject the OTel-instrumented
// MySQL connection); otherwise the driver opens its own connection.
func NewConnectionWithConn(databaseURL, databaseDriver string, conn gorm.ConnPool) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch databaseDriver {
	case "postgres":
		dialector = postgres.Open(databaseURL)
	case "mysql":
		if conn != nil {
			dialector = mysql.New(mysql.Config{Conn: conn, DSN: databaseURL})
		} else {
			dialector = mysql.Open(databaseURL)
		}
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
