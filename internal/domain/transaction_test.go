package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"flowledger/internal/domain"
)

func TestTransactionType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		txType    domain.TransactionType
		wantValid bool
		wantStr   string
	}{
		{
			name:      "income",
			txType:    domain.TransactionTypeIncome,
			wantValid: true,
			wantStr:   "income",
		},
		{
			name:      "expense",
			txType:    domain.TransactionTypeExpense,
			wantValid: true,
			wantStr:   "expense",
		},
		{
			name:      "invalid empty",
			txType:    domain.TransactionType(""),
			wantValid: false,
			wantStr:   "",
		},
		{
			name:      "invalid transfer",
			txType:    domain.TransactionType("transfer"),
			wantValid: false,
			wantStr:   "transfer",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.txType.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
			}
			if got := tt.txType.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestTransaction_Validate(t *testing.T) {
	t.Parallel()

	validTx := domain.Transaction{
		CategoryID: 1,
		Type:       domain.TransactionTypeExpense,
		AmountIDR:  50000,
		OccurredOn: "2026-10-30",
		Note:       "Makan siang ayam geprek",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	tests := []struct {
		name    string
		modify  func(tx *domain.Transaction)
		wantErr bool
	}{
		{
			name:    "valid expense transaction",
			modify:  func(tx *domain.Transaction) {},
			wantErr: false,
		},
		{
			name: "valid income transaction",
			modify: func(tx *domain.Transaction) {
				tx.Type = domain.TransactionTypeIncome
				tx.AmountIDR = 5000000
				tx.Note = "Gaji bulanan"
			},
			wantErr: false,
		},
		{
			name: "invalid zero category_id",
			modify: func(tx *domain.Transaction) {
				tx.CategoryID = 0
			},
			wantErr: true,
		},
		{
			name: "invalid negative category_id",
			modify: func(tx *domain.Transaction) {
				tx.CategoryID = -5
			},
			wantErr: true,
		},
		{
			name: "invalid transaction type",
			modify: func(tx *domain.Transaction) {
				tx.Type = domain.TransactionType("crypto")
			},
			wantErr: true,
		},
		{
			name: "invalid zero amount",
			modify: func(tx *domain.Transaction) {
				tx.AmountIDR = 0
			},
			wantErr: true,
		},
		{
			name: "invalid negative amount",
			modify: func(tx *domain.Transaction) {
				tx.AmountIDR = -25000
			},
			wantErr: true,
		},
		{
			name: "missing occurred_on",
			modify: func(tx *domain.Transaction) {
				tx.OccurredOn = ""
			},
			wantErr: true,
		},
		{
			name: "malformed occurred_on date",
			modify: func(tx *domain.Transaction) {
				tx.OccurredOn = "30-10-2026"
			},
			wantErr: true,
		},
		{
			name: "impossible calendar date",
			modify: func(tx *domain.Transaction) {
				tx.OccurredOn = "2026-02-30"
			},
			wantErr: true,
		},
		{
			name: "note exceeding 500 runes",
			modify: func(tx *domain.Transaction) {
				tx.Note = strings.Repeat("x", 501)
			},
			wantErr: true,
		},
		{
			name: "multiple invalid fields",
			modify: func(tx *domain.Transaction) {
				tx.CategoryID = 0
				tx.Type = "unknown"
				tx.AmountIDR = -100
				tx.OccurredOn = "invalid"
				tx.Note = strings.Repeat("x", 501)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := validTx
			tt.modify(&tx)

			err := tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransaction_ValidateTypeMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		txType    domain.TransactionType
		catType   domain.CategoryType
		wantErrIs error
	}{
		{
			name:      "matching expense",
			txType:    domain.TransactionTypeExpense,
			catType:   domain.CategoryTypeExpense,
			wantErrIs: nil,
		},
		{
			name:      "matching income",
			txType:    domain.TransactionTypeIncome,
			catType:   domain.CategoryTypeIncome,
			wantErrIs: nil,
		},
		{
			name:      "mismatch expense tx with income category",
			txType:    domain.TransactionTypeExpense,
			catType:   domain.CategoryTypeIncome,
			wantErrIs: domain.ErrTypeMismatch,
		},
		{
			name:      "mismatch income tx with expense category",
			txType:    domain.TransactionTypeIncome,
			catType:   domain.CategoryTypeExpense,
			wantErrIs: domain.ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := domain.Transaction{Type: tt.txType}
			err := tx.ValidateTypeMatch(tt.catType)

			if tt.wantErrIs == nil {
				if err != nil {
					t.Errorf("ValidateTypeMatch() = %v, want nil", err)
				}
			} else {
				if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("ValidateTypeMatch() = %v, want %v", err, tt.wantErrIs)
				}
			}
		})
	}
}
