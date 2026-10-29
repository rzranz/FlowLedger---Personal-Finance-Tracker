package httpapi

import (
	"net/http"
	"strconv"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

type CategoryHandler struct {
	uc *usecase.CategoryUsecase
}

func NewCategoryHandler(uc *usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	typeParam := r.URL.Query().Get("type")
	var catType *domain.CategoryType
	if typeParam != "" {
		ct := domain.CategoryType(typeParam)
		if !ct.IsValid() {
			var ve domain.ValidationErrors
			ve.Add("type", "must be 'income' or 'expense'")
			RespondError(w, r, ve)
			return
		}
		catType = &ct
	}

	categories, err := h.uc.List(r.Context(), catType)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]any{
		"items": categories,
	})
}

type CreateCategoryRequest struct {
	Name string              `json:"name"`
	Type domain.CategoryType `json:"type"`
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		RespondError(w, r, err)
		return
	}

	cat, err := h.uc.Create(r.Context(), req.Name, req.Type)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusCreated, cat)
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		var ve domain.ValidationErrors
		ve.Add("id", "invalid category ID")
		RespondError(w, r, ve)
		return
	}

	var req UpdateCategoryRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		RespondError(w, r, err)
		return
	}

	cat, err := h.uc.UpdateName(r.Context(), id, req.Name)
	if err != nil {
		RespondError(w, r, err)
		return
	}

	RespondJSON(w, http.StatusOK, cat)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		var ve domain.ValidationErrors
		ve.Add("id", "invalid category ID")
		RespondError(w, r, ve)
		return
	}

	if err := h.uc.Delete(r.Context(), id); err != nil {
		RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

