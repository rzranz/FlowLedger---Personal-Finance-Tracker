package httpapi

import (
	"math"
	"net/http"
	"strconv"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

type TransactionHandler struct {
	uc *usecase.TransactionUsecase
}

func NewTransactionHandler(uc *usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := usecase.TransactionFilter{
		Month: q.Get("month"),
		Query: q.Get("q"),
	}

	if tStr := q.Get("type"); tStr != "" {
		tt := domain.TransactionType(tStr)
		if !tt.IsValid() {
			var ve domain.ValidationErrors
			ve.Add("type", "must be 'income' or 'expense'")
			RespondError(w, r, ve)
			return
		}
		filter.Type = &tt
	}

	if catIDStr := q.Get("category_id"); catIDStr != "" {
		catID, err := strconv.ParseInt(catIDStr, 10, 64)
		if err != nil || catID <= 0 {
			var ve domain.ValidationErrors
			ve.Add("category_id", "invalid category ID")
			RespondError(w, r, ve)
			return
		}
		filter.CategoryID = &catID
	}

	page := 1
	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	filter.Page = page

	limit := 20
	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	filter.Limit = limit

	items, total, err := h.uc.List(r.Context(), filter)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	RespondJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"pagination": map[string]any{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in usecase.CreateTransactionInput
	if err := DecodeJSON(w, r, &in); err != nil {
		RespondError(w, r, err)
		return
	}

	tx, err := h.uc.Create(r.Context(), in)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		var ve domain.ValidationErrors
		ve.Add("id", "invalid transaction ID")
		RespondError(w, r, ve)
		return
	}

	tx, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		var ve domain.ValidationErrors
		ve.Add("id", "invalid transaction ID")
		RespondError(w, r, ve)
		return
	}

	var in usecase.UpdateTransactionInput
	if err := DecodeJSON(w, r, &in); err != nil {
		RespondError(w, r, err)
		return
	}

	tx, err := h.uc.Update(r.Context(), id, in)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		var ve domain.ValidationErrors
		ve.Add("id", "invalid transaction ID")
		RespondError(w, r, ve)
		return
	}

	if err := h.uc.Delete(r.Context(), id); err != nil {
		RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

