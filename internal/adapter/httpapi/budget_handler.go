package httpapi

import (
	"net/http"
	"strconv"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

type BudgetHandler struct {
	uc *usecase.BudgetUsecase
}

func NewBudgetHandler(uc *usecase.BudgetUsecase) *BudgetHandler {
	return &BudgetHandler{uc: uc}
}

func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		var ve domain.ValidationErrors
		ve.Add("month", "month parameter is required (format: YYYY-MM)")
		RespondError(w, r, ve)
		return
	}

	items, err := h.uc.ListByMonth(r.Context(), month)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

type UpsertBudgetRequest struct {
	LimitIDR domain.AmountIDR `json:"limit_idr"`
}

func (h *BudgetHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	catIDStr := r.PathValue("category_id")
	categoryID, err := strconv.ParseInt(catIDStr, 10, 64)
	if err != nil || categoryID <= 0 {
		var ve domain.ValidationErrors
		ve.Add("category_id", "invalid category ID")
		RespondError(w, r, ve)
		return
	}

	month := r.PathValue("month")
	if month == "" {
		var ve domain.ValidationErrors
		ve.Add("month", "month is required")
		RespondError(w, r, ve)
		return
	}

	var req UpsertBudgetRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		RespondError(w, r, err)
		return
	}

	b, err := h.uc.Upsert(r.Context(), usecase.UpsertBudgetInput{
		CategoryID: categoryID,
		Month:      month,
		LimitIDR:   req.LimitIDR,
	})
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusOK, b)
}

