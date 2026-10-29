package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// TransactionType defines whether money is entering or leaving the system.
type TransactionType string

const (
	// TransactionTypeIncome denotes cash entering.
	TransactionTypeIncome TransactionType = "income"

	// TransactionTypeExpense denotes cash leaving.
	TransactionTypeExpense TransactionType = "expense"
)

// IsValid checks if the transaction type is supported.
func (tt TransactionType) IsValid() bool {
	return tt == TransactionTypeIncome || tt == TransactionTypeExpense
}

// String returns the string representation of TransactionType.
func (tt TransactionType) String() string {
	return string(tt)
}

// Transaction represents an individual cash flow entry in whole Indonesian Rupiah.
type Transaction struct {
	ID         int64           `json:"id"`
	CategoryID int64           `json:"category_id"`
	Category   *Category       `json:"category,omitempty"`
	Type       TransactionType `json:"type"`
	AmountIDR  AmountIDR       `json:"amount_idr"`
	OccurredOn string          `json:"occurred_on"` // Format: YYYY-MM-DD
	Note       string          `json:"note"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// Validate verifies domain invariants on a Transaction struct.
func (t *Transaction) Validate() error {
	var ve ValidationErrors

	if t.CategoryID <= 0 {
		ve.Add("category_id", "category_id must be a positive integer")
	}

	if !t.Type.IsValid() {
		ve.Add("type", "type must be either 'income' or 'expense'")
	}

	if !t.AmountIDR.IsPositive() {
		ve.Add("amount_idr", "must be a positive integer greater than 0")
	}

	if t.OccurredOn == "" {
		ve.Add("occurred_on", "occurred_on date is required (format: YYYY-MM-DD)")
	} else if _, err := time.Parse(time.DateOnly, t.OccurredOn); err != nil {
		ve.Add("occurred_on", "occurred_on must be a valid calendar date in YYYY-MM-DD format")
	}

	trimmedNote := strings.TrimSpace(t.Note)
	if utf8.RuneCountInString(trimmedNote) > 500 {
		ve.Add("note", "note must not exceed 500 characters")
	}

	return ve.AsError()
}

// ValidateTypeMatch checks that the transaction type matches the category type.
func (t *Transaction) ValidateTypeMatch(catType CategoryType) error {
	if string(t.Type) != string(catType) {
		return fmt.Errorf("%w: transaction of type %q cannot reference category of type %q",
			ErrTypeMismatch, t.Type, catType)
	}
	return nil
}

