package domain_test

import (
	"errors"
	"testing"
	"time"

	"flowledger/internal/domain"
)

func TestBudgetStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status  domain.BudgetStatus
		wantStr string
	}{
		{status: domain.BudgetStatusSafe, wantStr: "SAFE"},
		{status: domain.BudgetStatusWarning, wantStr: "WARNING"},
		{status: domain.BudgetStatusExceeded, wantStr: "EXCEEDED"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.wantStr, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestBudget_Validate(t *testing.T) {
	t.Parallel()

	validBudget := domain.Budget{
		CategoryID: 1,
		Month:      "2026-10",
		LimitIDR:   1000000,
		UpdatedAt:  time.Now(),
	}

	tests := []struct {
		name    string
		modify  func(b *domain.Budget)
		wantErr bool
	}{
		{
			name:    "valid budget",
			modify:  func(b *domain.Budget) {},
			wantErr: false,
		},
		{
			name: "invalid zero category_id",
			modify: func(b *domain.Budget) {
				b.CategoryID = 0
			},
			wantErr: true,
		},
		{
			name: "invalid negative category_id",
			modify: func(b *domain.Budget) {
				b.CategoryID = -1
			},
			wantErr: true,
		},
		{
			name: "missing month",
			modify: func(b *domain.Budget) {
				b.Month = ""
			},
			wantErr: true,
		},
		{
			name: "malformed month format",
			modify: func(b *domain.Budget) {
				b.Month = "10-2026"
			},
			wantErr: true,
		},
		{
			name: "invalid month number",
			modify: func(b *domain.Budget) {
				b.Month = "2026-13"
			},
			wantErr: true,
		},
		{
			name: "invalid zero limit",
			modify: func(b *domain.Budget) {
				b.LimitIDR = 0
			},
			wantErr: true,
		},
		{
			name: "invalid negative limit",
			modify: func(b *domain.Budget) {
				b.LimitIDR = -500000
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := validBudget
			tt.modify(&b)

			err := b.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBudget_ValidateCategoryType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		catType   domain.CategoryType
		wantErrIs error
	}{
		{
			name:      "valid expense category",
			catType:   domain.CategoryTypeExpense,
			wantErrIs: nil,
		},
		{
			name:      "invalid income category",
			catType:   domain.CategoryTypeIncome,
			wantErrIs: domain.ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := domain.Budget{}
			err := b.ValidateCategoryType(tt.catType)

			if tt.wantErrIs == nil {
				if err != nil {
					t.Errorf("ValidateCategoryType() = %v, want nil", err)
				}
			} else {
				if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("ValidateCategoryType() = %v, want %v", err, tt.wantErrIs)
				}
			}
		})
	}
}

func TestCalculateUtilization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		limitIDR      domain.AmountIDR
		spentIDR      domain.AmountIDR
		wantRemaining domain.AmountIDR
		wantPct       int
		wantStatus    domain.BudgetStatus
	}{
		{
			name:          "0% utilization (zero spent)",
			limitIDR:      1000000,
			spentIDR:      0,
			wantRemaining: 1000000,
			wantPct:       0,
			wantStatus:    domain.BudgetStatusSafe,
		},
		{
			name:          "50% utilization (safe)",
			limitIDR:      1000000,
			spentIDR:      500000,
			wantRemaining: 500000,
			wantPct:       50,
			wantStatus:    domain.BudgetStatusSafe,
		},
		{
			name:          "79% utilization (safe threshold)",
			limitIDR:      1000000,
			spentIDR:      790000,
			wantRemaining: 210000,
			wantPct:       79,
			wantStatus:    domain.BudgetStatusSafe,
		},
		{
			name:          "80% utilization (warning trigger)",
			limitIDR:      1000000,
			spentIDR:      800000,
			wantRemaining: 200000,
			wantPct:       80,
			wantStatus:    domain.BudgetStatusWarning,
		},
		{
			name:          "99% utilization (warning limit)",
			limitIDR:      1000000,
			spentIDR:      990000,
			wantRemaining: 10000,
			wantPct:       99,
			wantStatus:    domain.BudgetStatusWarning,
		},
		{
			name:          "100% utilization (exceeded trigger)",
			limitIDR:      1000000,
			spentIDR:      1000000,
			wantRemaining: 0,
			wantPct:       100,
			wantStatus:    domain.BudgetStatusExceeded,
		},
		{
			name:          "125% utilization (over-budget with negative remaining)",
			limitIDR:      1000000,
			spentIDR:      1250000,
			wantRemaining: -250000,
			wantPct:       125,
			wantStatus:    domain.BudgetStatusExceeded,
		},
		{
			name:          "zero limit guard",
			limitIDR:      0,
			spentIDR:      50000,
			wantRemaining: -50000,
			wantPct:       0,
			wantStatus:    domain.BudgetStatusSafe,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rem, pct, status := domain.CalculateUtilization(tt.limitIDR, tt.spentIDR)

			if rem != tt.wantRemaining {
				t.Errorf("Remaining = %d, want %d", rem, tt.wantRemaining)
			}
			if pct != tt.wantPct {
				t.Errorf("UtilizationPct = %d, want %d", pct, tt.wantPct)
			}
			if status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", status, tt.wantStatus)
			}
		})
	}
}

