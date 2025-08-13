package main

import (
	"log/slog"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		slog.Info("Request started",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		slog.Info("Request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", rw.statusCode,
			"response_size", rw.size,
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
		)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
