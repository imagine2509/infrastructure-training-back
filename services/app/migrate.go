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
)

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
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to execute migration %s: %v (rollback failed: %v)", migrationName, err, rollbackErr)
			}
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		if _, err := tx.Exec("INSERT INTO migrations (name) VALUES ($1)", migrationName); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to record migration %s: %v (rollback failed: %v)", migrationName, err, rollbackErr)
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
