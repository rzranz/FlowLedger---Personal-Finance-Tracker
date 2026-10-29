package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

// CategoryType defines the permissible financial classifications for categories.
type CategoryType string

const (
	// CategoryTypeIncome designates categories for inward cash flow.
	CategoryTypeIncome CategoryType = "income"

	// CategoryTypeExpense designates categories for outward cash flow.
	CategoryTypeExpense CategoryType = "expense"
)

// IsValid checks if the CategoryType is either income or expense.
func (ct CategoryType) IsValid() bool {
	return ct == CategoryTypeIncome || ct == CategoryTypeExpense
}

// String returns the string representation of CategoryType.
func (ct CategoryType) String() string {
	return string(ct)
}

// Category represents a grouping for financial transactions and monthly budgets.
type Category struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Type      CategoryType `json:"type"`
	CreatedAt time.Time    `json:"created_at"`
}

// Validate ensures that category invariants are upheld.
func (c *Category) Validate() error {
	var ve ValidationErrors

	trimmedName := strings.TrimSpace(c.Name)
	if trimmedName == "" {
		ve.Add("name", "name is required and cannot be empty")
	} else if utf8.RuneCountInString(trimmedName) > 100 {
		ve.Add("name", "name must not exceed 100 characters")
	}

	if !c.Type.IsValid() {
		ve.Add("type", "type must be either 'income' or 'expense'")
	}

	return ve.AsError()
}

