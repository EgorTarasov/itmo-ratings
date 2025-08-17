package info_handler

import (
	"encoding/json"
	"net/http"
	"os"
)

type Handler struct {
	commitID string
}

func New(	) *Handler {
	commitID := os.Getenv("COMMIT_ID")

	return &Handler{
		commitID: commitID,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"commitID": h.commitID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
