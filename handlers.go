package main

import (
	"encoding/json"
	"net/http"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	if err != nil {
		return
	}
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	if err != nil {
		return
	}
}

func notesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var req NoteRequest
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

		response := NoteResponse{Text: req.Text}
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		if err != nil {
			return
		}
	}
}
