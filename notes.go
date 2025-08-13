package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
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

func getNotesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

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

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(notes)
		if err != nil {
			return
		}
	}
}

func createNoteHandler(db *sql.DB) http.HandlerFunc {
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

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(note)
		if err != nil {
			return
		}
	}
}

func deleteNoteHandler(db *sql.DB) http.HandlerFunc {
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

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(map[string]string{"message": "Note deleted successfully"})
		if err != nil {
			return
		}
	}
}
