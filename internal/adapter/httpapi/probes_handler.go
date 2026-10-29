package httpapi

import (
	"context"
	"net/http"
	"time"

	"flowledger/internal/adapter/sqlite"
)

type ProbesHandler struct {
	db *sqlite.DB
}

func NewProbesHandler(db *sqlite.DB) *ProbesHandler {
	return &ProbesHandler{db: db}
}

func (h *ProbesHandler) Health(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (h *ProbesHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		RespondJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"error":  err.Error(),
		})
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

