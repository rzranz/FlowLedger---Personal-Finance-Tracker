package domain_test

import (
	"strings"
	"testing"
	"time"

	"flowledger/internal/domain"
)

func TestCategoryType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		catType   domain.CategoryType
		wantValid bool
		wantStr   string
	}{
		{
			name:      "valid income",
			catType:   domain.CategoryTypeIncome,
			wantValid: true,
			wantStr:   "income",
		},
		{
			name:      "valid expense",
			catType:   domain.CategoryTypeExpense,
			wantValid: true,
			wantStr:   "expense",
		},
		{
			name:      "invalid type empty",
			catType:   domain.CategoryType(""),
			wantValid: false,
			wantStr:   "",
		},
		{
			name:      "invalid type arbitrary",
			catType:   domain.CategoryType("investment"),
			wantValid: false,
			wantStr:   "investment",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.catType.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
			}
			if got := tt.catType.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestCategory_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		category   domain.Category
		wantErr    bool
		wantFields []string
	}{
		{
			name: "valid income category",
			category: domain.Category{
				Name:      "Gaji Bulanan",
				Type:      domain.CategoryTypeIncome,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid expense category",
			category: domain.Category{
				Name:      "Makanan & Minuman",
				Type:      domain.CategoryTypeExpense,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty name and invalid type",
			category: domain.Category{
				Name: "   ",
				Type: domain.CategoryType("unknown"),
			},
			wantErr:    true,
			wantFields: []string{"name", "type"},
		},
		{
			name: "name too long (exceeds 100 runes)",
			category: domain.Category{
				Name: strings.Repeat("A", 101),
				Type: domain.CategoryTypeExpense,
			},
			wantErr:    true,
			wantFields: []string{"name"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.category.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ve domain.ValidationErrors
				if !strings.Contains(err.Error(), "validation failed") {
					t.Errorf("expected validation failure message, got %v", err)
				}
				_ = ve
			}
		})
	}
}

