package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Common domain sentinel errors used across FlowLedger.
var (
	// ErrNotFound is returned when a requested domain entity does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict is returned when an operation conflicts with existing data (e.g., duplicate names).
	ErrConflict = errors.New("resource conflict")

	// ErrReferenced is returned when attempting to delete a category that is still referenced by transactions or budgets.
	ErrReferenced = errors.New("resource is referenced and cannot be deleted")

	// ErrTypeMismatch is returned when an entity's type does not match its related entity's type.
	ErrTypeMismatch = errors.New("type mismatch between entities")

	// ErrIdempotencyConflict is returned when an Idempotency-Key is reused with a different request payload.
	ErrIdempotencyConflict = errors.New("idempotency key reused with different payload")

	// ErrInvalidInput is returned when input data fails general validation rules.
	ErrInvalidInput = errors.New("invalid input data")
)

// FieldViolation describes a validation error on a specific struct field.
type FieldViolation struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// ValidationErrors represents a collection of field-level validation errors.
// It implements the standard error interface.
type ValidationErrors []FieldViolation

// Error returns a formatted string containing all validation violations.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("validation failed: ")
	for i, v := range ve {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", v.Field, v.Reason))
	}
	return sb.String()
}

// HasErrors returns true if there are any field violations.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Add appends a new field violation to the collection.
func (ve *ValidationErrors) Add(field, reason string) {
	*ve = append(*ve, FieldViolation{
		Field:  field,
		Reason: reason,
	})
}

// AsError returns nil if there are no errors, or the ValidationErrors itself.
func (ve ValidationErrors) AsError() error {
	if len(ve) == 0 {
		return nil
	}
	return ve
}

