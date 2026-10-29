package domain

import (
	"fmt"
	"time"
)

// BudgetStatus defines the state of a monthly budget based on expenditure thresholds.
type BudgetStatus string

const (
	// BudgetStatusSafe indicates utilization is strictly under 80%.
	BudgetStatusSafe BudgetStatus = "SAFE"

	// BudgetStatusWarning indicates utilization is between 80% and 99% inclusive.
	BudgetStatusWarning BudgetStatus = "WARNING"

	// BudgetStatusExceeded indicates expenditure has reached or exceeded 100% of the limit.
	BudgetStatusExceeded BudgetStatus = "EXCEEDED"
)

// String returns the string representation of BudgetStatus.
func (bs BudgetStatus) String() string {
	return string(bs)
}

// Budget represents a monthly spending ceiling for a specific expense category.
type Budget struct {
	ID         int64     `json:"id"`
	CategoryID int64     `json:"category_id"`
	Month      string    `json:"month"` // Format: YYYY-MM
	LimitIDR   AmountIDR `json:"limit_idr"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate verifies domain invariants on a Budget struct.
func (b *Budget) Validate() error {
	var ve ValidationErrors

	if b.CategoryID <= 0 {
		ve.Add("category_id", "category_id must be a positive integer")
	}

	if b.Month == "" {
		ve.Add("month", "month is required (format: YYYY-MM)")
	} else if _, err := time.Parse("2006-01", b.Month); err != nil {
		ve.Add("month", "month must be a valid calendar month in YYYY-MM format")
	}

	if !b.LimitIDR.IsPositive() {
		ve.Add("limit_idr", "must be a positive integer greater than 0")
	}

	return ve.AsError()
}

// ValidateCategoryType ensures a budget is only associated with an expense category.
func (b *Budget) ValidateCategoryType(catType CategoryType) error {
	if catType != CategoryTypeExpense {
		return fmt.Errorf("%w: budgets can only be assigned to expense categories", ErrTypeMismatch)
	}
	return nil
}

// CalculateUtilization computes the remaining budget, utilization percentage, and status.
// Percentage is calculated as integer percent (e.g., 85 for 85%).
func CalculateUtilization(limitIDR, spentIDR AmountIDR) (remaining AmountIDR, pct int, status BudgetStatus) {
	remaining = limitIDR.Sub(spentIDR)

	if limitIDR <= 0 {
		return remaining, 0, BudgetStatusSafe
	}

	// Integer math: (spent * 100) / limit
	pct = int((spentIDR.Int64() * 100) / limitIDR.Int64())

	switch {
	case pct >= 100:
		status = BudgetStatusExceeded
	case pct >= 80:
		status = BudgetStatusWarning
	default:
		status = BudgetStatusSafe
	}

	return remaining, pct, status
}

// BudgetReportItem aggregates a budget with its real-time utilization details.
type BudgetReportItem struct {
	Budget         Budget       `json:"budget"`
	CategoryName   string       `json:"category_name"`
	SpentIDR       AmountIDR    `json:"spent_idr"`
	RemainingIDR   AmountIDR    `json:"remaining_idr"`
	UtilizationPct int          `json:"utilization_pct"`
	Status         BudgetStatus `json:"status"`
}

