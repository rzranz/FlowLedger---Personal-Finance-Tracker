package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"flowledger/internal/domain"
)

// ProblemDetails represents an RFC 9457 HTTP Problem Details payload.
type ProblemDetails struct {
	Type             string                  `json:"type"`
	Title            string                  `json:"title"`
	Status           int                     `json:"status"`
	Detail           string                  `json:"detail"`
	Instance         string                  `json:"instance"`
	ValidationErrors []domain.FieldViolation `json:"validation_errors,omitempty"`
}

// RespondJSON writes a JSON response with the given status code.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// RespondError maps Go and domain errors into RFC 9457 Problem Details JSON responses.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	problem := ProblemDetails{
		Instance: r.URL.Path,
	}

	var ve domain.ValidationErrors
	if errors.As(err, &ve) {
		problem.Type = "https://flowledger.dev/errors/validation-failed"
		problem.Title = "Unprocessable Entity"
		problem.Status = http.StatusUnprocessableEntity
		problem.Detail = "Validation failed on one or more fields."
		problem.ValidationErrors = ve
		RespondJSON(w, problem.Status, problem)
		return
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		problem.Type = "https://flowledger.dev/errors/not-found"
		problem.Title = "Not Found"
		problem.Status = http.StatusNotFound
		problem.Detail = err.Error()

	case errors.Is(err, domain.ErrConflict):
		problem.Type = "https://flowledger.dev/errors/conflict"
		problem.Title = "Conflict"
		problem.Status = http.StatusConflict
		problem.Detail = err.Error()

	case errors.Is(err, domain.ErrReferenced):
		problem.Type = "https://flowledger.dev/errors/referenced-conflict"
		problem.Title = "Conflict"
		problem.Status = http.StatusConflict
		problem.Detail = "Resource is referenced by transactions or budgets and cannot be deleted."

	case errors.Is(err, domain.ErrTypeMismatch):
		problem.Type = "https://flowledger.dev/errors/type-mismatch"
		problem.Title = "Unprocessable Entity"
		problem.Status = http.StatusUnprocessableEntity
		problem.Detail = err.Error()

	case errors.Is(err, domain.ErrIdempotencyConflict):
		problem.Type = "https://flowledger.dev/errors/idempotency-conflict"
		problem.Title = "Idempotency Conflict"
		problem.Status = http.StatusConflict
		problem.Detail = "Idempotency key has already been used with a different request payload."

	case errors.Is(err, domain.ErrInvalidInput):
		problem.Type = "https://flowledger.dev/errors/invalid-input"
		problem.Title = "Bad Request"
		problem.Status = http.StatusBadRequest
		problem.Detail = err.Error()

	default:
		// Unknown or unexpected server error
		problem.Type = "https://flowledger.dev/errors/internal-error"
		problem.Title = "Internal Server Error"
		problem.Status = http.StatusInternalServerError
		problem.Detail = "An unexpected error occurred processing your request."
	}

	RespondJSON(w, problem.Status, problem)
}

// DecodeJSON decodes request body while rejecting unknown fields and enforcing 1 MB limit.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Limit request body to 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	return nil
}

