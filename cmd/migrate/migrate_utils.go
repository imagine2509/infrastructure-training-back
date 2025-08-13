package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func initDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "infrastructure_training"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connected successfully",
		"host", host,
		"port", port,
		"user", user,
		"dbname", dbname,
	)

	return db, nil
}

func runMigrations(db *sql.DB) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var sqlFiles []fs.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") && !file.IsDir() {
			sqlFiles = append(sqlFiles, file)
		}
	}

	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	for _, file := range sqlFiles {
		migrationName := file.Name()

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", migrationName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check if migration %s exists: %w", migrationName, err)
		}

		if exists {
			slog.Info("Migration already applied", "migration", migrationName)
			continue
		}

		filePath := filepath.Join(migrationsDir, migrationName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationName, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migrationName, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		if _, err := tx.Exec("INSERT INTO migrations (name) VALUES ($1)", migrationName); err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migrationName, err)
		}

		slog.Info("Applied migration", "migration", migrationName)
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`

	_, err := db.Exec(query)
	return err
}

func showMigrationStatus(db *sql.DB) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	rows, err := db.Query("SELECT name, applied_at FROM migrations ORDER BY applied_at")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error("Failed to close rows", "error", err)
		}
	}(rows)

	fmt.Println("Applied migrations:")
	fmt.Println("===================")

	count := 0
	for rows.Next() {
		var name string
		var appliedAt time.Time
		if err := rows.Scan(&name, &appliedAt); err != nil {
			return fmt.Errorf("failed to scan migration row: %w", err)
		}
		fmt.Printf("✓ %s (applied: %s)\n", name, appliedAt.Format("2006-01-02 15:04:05"))
		count++
	}

	if count == 0 {
		fmt.Println("No migrations applied yet")
	}

	// Check for pending migrations
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	fmt.Println("\nPending migrations:")
	fmt.Println("==================")

	pendingCount := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") && !file.IsDir() {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", file.Name()).Scan(&exists)
			if err != nil {
				return fmt.Errorf("failed to check migration status: %w", err)
			}

			if !exists {
				fmt.Printf("✗ %s\n", file.Name())
				pendingCount++
			}
		}
	}

	if pendingCount == 0 {
		fmt.Println("No pending migrations")
	}

	return nil
}
