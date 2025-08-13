package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	var action string
	flag.StringVar(&action, "action", "up", "Migration action: up, down, status")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Change to project root directory
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		slog.Error("Failed to change directory", "error", err)
		os.Exit(1)
	}

	db, err := initDB()
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}(db)

	switch action {
	case "up":
		if err := runMigrations(db); err != nil {
			slog.Error("Failed to run migrations", "error", err)
			os.Exit(1)
		}
		slog.Info("Migrations completed successfully")
	case "status":
		if err := showMigrationStatus(db); err != nil {
			slog.Error("Failed to show migration status", "error", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown action: %s\n", action)
		fmt.Println("Available actions: up, status")
		os.Exit(1)
	}
}
