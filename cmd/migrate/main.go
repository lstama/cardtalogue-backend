package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lstama/cardtalogue-backend/internal/config"
	"github.com/lstama/cardtalogue-backend/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Load()
	command := os.Args[1]
	driver := cfg.DatabaseDriver

	switch command {
	case "up":
		if err := database.RunMigrations(cfg.DatabaseURL, driver, cfg.MigrationsPath); err != nil {
			fmt.Printf("Error running migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations completed successfully")

	case "down":
		if len(os.Args) < 3 {
			fmt.Println("Please specify the number of steps to rollback")
			os.Exit(1)
		}
		steps, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid steps number: %v\n", err)
			os.Exit(1)
		}
		if err := database.RollbackMigrations(cfg.DatabaseURL, driver, cfg.MigrationsPath, steps); err != nil {
			fmt.Printf("Error rolling back migrations: %v\n", err)
			os.Exit(1)
		}

	case "version":
		version, dirty, err := database.GetMigrationVersion(cfg.DatabaseURL, driver, cfg.MigrationsPath)
		if err != nil {
			fmt.Printf("Error getting migration version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current version: %d, Dirty: %t\n", version, dirty)

	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Please specify migration name")
			os.Exit(1)
		}
		name := os.Args[2]
		if err := createMigrationFiles(name); err != nil {
			fmt.Printf("Error creating migration files: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created migration files for: %s\n", name)

	case "force":
		if len(os.Args) < 3 {
			fmt.Println("Please specify version to force")
			os.Exit(1)
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid version number: %v\n", err)
			os.Exit(1)
		}
		if err := forceMigrationVersion(cfg.DatabaseURL, driver, cfg.MigrationsPath, version); err != nil {
			fmt.Printf("Error forcing migration version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Forced migration version to: %d\n", version)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Migration tool usage:")
	fmt.Println("  go run cmd/migrate/main.go up                    - Run all pending migrations")
	fmt.Println("  go run cmd/migrate/main.go down <steps>          - Rollback migrations")
	fmt.Println("  go run cmd/migrate/main.go version               - Show current version")
	fmt.Println("  go run cmd/migrate/main.go create <name>         - Create new migration")
	fmt.Println("  go run cmd/migrate/main.go force <version>       - Force version")
}

func createMigrationFiles(name string) error {
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s", timestamp, name)

	upFile := fmt.Sprintf("migrations/%s.up.sql", filename)
	downFile := fmt.Sprintf("migrations/%s.down.sql", filename)

	// Create up migration file
	upContent := fmt.Sprintf("-- Migration: %s\n-- Created at: %s\n\n-- Add your UP migration SQL here\n", name, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return err
	}

	// Create down migration file
	downContent := fmt.Sprintf("-- Migration: %s (rollback)\n-- Created at: %s\n\n-- Add your DOWN migration SQL here\n", name, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return err
	}

	fmt.Printf("Created: %s\n", upFile)
	fmt.Printf("Created: %s\n", downFile)
	return nil
}

func forceMigrationVersion(databaseURL, databaseDriver, migrationsPath string, version int) error {
	var driver sql.Driver
	var migrateDriver interface{}

	switch databaseDriver {
	case "postgres":
		driver = postgres.New(databaseURL, &postgres.Config{})
		migrateDriver = postgres.WithInstance(driver, &postgres.Config{})
	case "mysql":
		driver = mysql.New(databaseURL, &mysql.Config{})
		migrateDriver = mysql.WithInstance(driver, &mysql.Config{})
	case "sqlite":
		driver = sqlite3.New(databaseURL)
		migrateDriver = sqlite3.WithInstance(driver, &sqlite3.Config{})
	default:
		return fmt.Errorf("unsupported database driver: %s", databaseDriver)
	}

	db, err := sql.Open(databaseDriver, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, databaseDriver, migrateDriver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	return m.Force(version)
}
