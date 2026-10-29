package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

type ExportHandler struct {
	uc *usecase.TransactionUsecase
}

func NewExportHandler(uc *usecase.TransactionUsecase) *ExportHandler {
	return &ExportHandler{uc: uc}
}

func (h *ExportHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
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

	filename := fmt.Sprintf("flowledger-transactions-%s.csv", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	if err := h.uc.StreamCSV(r.Context(), filter, w); err != nil {
		// Because headers are already sent, we cannot change status, but we log or report
		return
	}
}

