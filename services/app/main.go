package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	db, err := initDB()
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	rdb := initRedis()
	if rdb != nil {
		defer func(rdb *redis.Client) {
			err := rdb.Close()
			if err != nil {
				slog.Error("Failed to close redis connection", "error", err)
			}
		}(rdb)
	}

	if err := runMigrations(db); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	r := mux.NewRouter()

	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)

	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/ping", pingHandler).Methods("GET")
	r.HandleFunc("/api/notes", createNoteHandler(db, rdb)).Methods("POST")
	r.HandleFunc("/api/notes", getNotesHandler(db, rdb)).Methods("GET")
	r.HandleFunc("/api/notes/{id}", deleteNoteHandler(db, rdb)).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped")
}
