package httpapi

import (
	"net/http"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

type ReportHandler struct {
	uc *usecase.ReportUsecase
}

func NewReportHandler(uc *usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

func (h *ReportHandler) Monthly(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		var ve domain.ValidationErrors
		ve.Add("month", "month query parameter is required (format: YYYY-MM)")
		RespondError(w, r, ve)
		return
	}

	report, err := h.uc.GetMonthlyReport(r.Context(), month)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusOK, report)
}

