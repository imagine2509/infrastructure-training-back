package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type Note struct {
	ID        int       `json:"id" db:"id"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type NoteCreateRequest struct {
	Text string `json:"text"`
}

func getNotesHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx := context.Background()

		// Try to get from cache first
		if rdb != nil {
			cachedNotes, err := rdb.Get(ctx, "notes:all").Result()
			if err == nil {
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(cachedNotes))
				if err != nil {
					return
				}
				return
			}
		}

		rows, err := db.Query("SELECT id, text, created_at, updated_at FROM notes ORDER BY created_at DESC")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
			if err != nil {
				return
			}
			return
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {

			}
		}(rows)

		var notes []Note
		for rows.Next() {
			var note Note
			err := rows.Scan(&note.ID, &note.Text, &note.CreatedAt, &note.UpdatedAt)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Database scan error"})
				if err != nil {
					return
				}
				return
			}
			notes = append(notes, note)
		}

		if notes == nil {
			notes = []Note{}
		}

		// Cache the result for 5 minutes
		if rdb != nil {
			notesJSON, _ := json.Marshal(notes)
			err = rdb.Set(ctx, "notes:all", notesJSON, 5*time.Minute).Err()
			if err != nil {
				slog.Warn("Failed to cache notes", "error", err)
			}
		}

		w.Header().Set("X-Cache", "MISS")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(notes)
		if err != nil {
			return
		}
	}
}

func createNoteHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req NoteCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
			if err != nil {
				return
			}
			return
		}

		if req.Text == "" {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Text field is required"})
			if err != nil {
				return
			}
			return
		}

		var note Note
		err := db.QueryRow(`
			INSERT INTO notes (text) 
			VALUES ($1) 
			RETURNING id, text, created_at, updated_at`,
			req.Text).Scan(&note.ID, &note.Text, &note.CreatedAt, &note.UpdatedAt)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
			if err != nil {
				return
			}
			return
		}

		// Invalidate cache after creating a note
		if rdb != nil {
			ctx := context.Background()
			err = rdb.Del(ctx, "notes:all").Err()
			if err != nil {
				slog.Warn("Failed to invalidate cache", "error", err)
			}
		}

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(note)
		if err != nil {
			return
		}
	}
}

func deleteNoteHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr, exists := vars["id"]
		if !exists {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "ID is required"})
			if err != nil {
				return
			}
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid ID format"})
			if err != nil {
				return
			}
			return
		}

		result, err := db.Exec("DELETE FROM notes WHERE id = $1", id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
			if err != nil {
				return
			}
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
			if err != nil {
				return
			}
			return
		}

		if rowsAffected == 0 {
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Note not found"})
			if err != nil {
				return
			}
			return
		}

		// Invalidate cache after deleting a note
		if rdb != nil {
			ctx := context.Background()
			err = rdb.Del(ctx, "notes:all").Err()
			if err != nil {
				slog.Warn("Failed to invalidate cache", "error", err)
			}
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(map[string]string{"message": "Note deleted successfully"})
		if err != nil {
			return
		}
	}
}
