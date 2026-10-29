package domain_test

import (
	"errors"
	"testing"

	"flowledger/internal/domain"
)

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty validation errors", func(t *testing.T) {
		t.Parallel()
		var ve domain.ValidationErrors

		if ve.HasErrors() {
			t.Errorf("HasErrors() = true, want false")
		}
		if err := ve.AsError(); err != nil {
			t.Errorf("AsError() = %v, want nil", err)
		}
		if s := ve.Error(); s != "" {
			t.Errorf("Error() = %q, want empty string", s)
		}
	})

	t.Run("multiple field errors", func(t *testing.T) {
		t.Parallel()
		var ve domain.ValidationErrors
		ve.Add("name", "cannot be empty")
		ve.Add("amount_idr", "must be greater than 0")

		if !ve.HasErrors() {
			t.Errorf("HasErrors() = false, want true")
		}
		err := ve.AsError()
		if err == nil {
			t.Fatalf("AsError() = nil, want error")
		}

		var target domain.ValidationErrors
		if !errors.As(err, &target) {
			t.Errorf("errors.As failed to match ValidationErrors")
		}

		expected := "validation failed: name: cannot be empty; amount_idr: must be greater than 0"
		if got := err.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})
}

