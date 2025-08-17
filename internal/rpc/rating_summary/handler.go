package rating_summary

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
)

type ratingService interface {
	GetStudentSummary(context.Context, string) (string, error)
}

type Handler struct {
	rating ratingService
}

func New(rating ratingService) *Handler {
	return &Handler{
		rating: rating,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	studentID := r.PathValue("id")
	if studentID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err := strconv.Atoi(studentID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	summary, err := h.rating.GetStudentSummary(r.Context(), studentID)
	if err != nil {
		slog.Error("failed to get student summary", "studentID", studentID, "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(summary))
}
