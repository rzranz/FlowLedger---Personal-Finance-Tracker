package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"flowledger/internal/domain"
)

// TransactionFilter holds query parameters for filtering and paginating transactions.
type TransactionFilter struct {
	Month      string
	Type       *domain.TransactionType
	CategoryID *int64
	Query      string
	Page       int
	Limit      int
}

// TransactionStore defines data access methods needed by TransactionUsecase.
type TransactionStore interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	GetByID(ctx context.Context, id int64) (*domain.Transaction, error)
	List(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, int64, error)
	Update(ctx context.Context, tx *domain.Transaction) error
	Delete(ctx context.Context, id int64) error
	StreamTransactions(ctx context.Context, filter TransactionFilter, fn func(tx domain.Transaction) error) error
}

// TransactionUsecase handles transaction business logic and CSV exports.
type TransactionUsecase struct {
	txStore  TransactionStore
	catStore CategoryStore
}

// NewTransactionUsecase creates a new TransactionUsecase.
func NewTransactionUsecase(txStore TransactionStore, catStore CategoryStore) *TransactionUsecase {
	return &TransactionUsecase{
		txStore:  txStore,
		catStore: catStore,
	}
}

// CreateInput holds payload for creating a transaction.
type CreateTransactionInput struct {
	CategoryID int64                  `json:"category_id"`
	Type       domain.TransactionType `json:"type"`
	AmountIDR  domain.AmountIDR       `json:"amount_idr"`
	OccurredOn string                 `json:"occurred_on"`
	Note       string                 `json:"note"`
}

// Create validates and saves a new transaction.
func (uc *TransactionUsecase) Create(ctx context.Context, in CreateTransactionInput) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		CategoryID: in.CategoryID,
		Type:       in.Type,
		AmountIDR:  in.AmountIDR,
		OccurredOn: strings.TrimSpace(in.OccurredOn),
		Note:       strings.TrimSpace(in.Note),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	cat, err := uc.catStore.GetByID(ctx, tx.CategoryID)
	if err != nil {
		return nil, err
	}

	if err := tx.ValidateTypeMatch(cat.Type); err != nil {
		return nil, err
	}

	if err := uc.txStore.Create(ctx, tx); err != nil {
		return nil, err
	}

	tx.Category = cat
	return tx, nil
}

// UpdateTransactionInput holds payload for updating a transaction.
type UpdateTransactionInput struct {
	CategoryID *int64                  `json:"category_id"`
	Type       *domain.TransactionType `json:"type"`
	AmountIDR  *domain.AmountIDR       `json:"amount_idr"`
	OccurredOn *string                 `json:"occurred_on"`
	Note       *string                 `json:"note"`
}

// Update modifies an existing transaction while enforcing invariants.
func (uc *TransactionUsecase) Update(ctx context.Context, id int64, in UpdateTransactionInput) (*domain.Transaction, error) {
	existing, err := uc.txStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.CategoryID != nil {
		existing.CategoryID = *in.CategoryID
	}
	if in.Type != nil {
		existing.Type = *in.Type
	}
	if in.AmountIDR != nil {
		existing.AmountIDR = *in.AmountIDR
	}
	if in.OccurredOn != nil {
		existing.OccurredOn = strings.TrimSpace(*in.OccurredOn)
	}
	if in.Note != nil {
		existing.Note = strings.TrimSpace(*in.Note)
	}

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	cat, err := uc.catStore.GetByID(ctx, existing.CategoryID)
	if err != nil {
		return nil, err
	}

	if err := existing.ValidateTypeMatch(cat.Type); err != nil {
		return nil, err
	}

	if err := uc.txStore.Update(ctx, existing); err != nil {
		return nil, err
	}

	existing.Category = cat
	return existing, nil
}

// GetByID retrieves a transaction by ID.
func (uc *TransactionUsecase) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return uc.txStore.GetByID(ctx, id)
}

// List queries transactions with filtering and pagination.
func (uc *TransactionUsecase) List(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, int64, error) {
	return uc.txStore.List(ctx, filter)
}

// Delete removes a transaction.
func (uc *TransactionUsecase) Delete(ctx context.Context, id int64) error {
	return uc.txStore.Delete(ctx, id)
}

// SanitizeCSVField prevents CSV formula injection.
func SanitizeCSVField(val string) string {
	if len(val) > 0 {
		switch val[0] {
		case '=', '+', '-', '@':
			return "'" + val
		}
	}
	return val
}

// StreamCSV writes transactions as RFC 4180 CSV directly to the given io.Writer.
func (uc *TransactionUsecase) StreamCSV(ctx context.Context, filter TransactionFilter, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write CSV Header
	header := []string{"ID", "Date", "Type", "Category", "Amount (IDR)", "Note"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	csvWriter.Flush()

	// Stream rows
	err := uc.txStore.StreamTransactions(ctx, filter, func(tx domain.Transaction) error {
		catName := ""
		if tx.Category != nil {
			catName = tx.Category.Name
		}
		row := []string{
			strconv.FormatInt(tx.ID, 10),
			tx.OccurredOn,
			string(tx.Type),
			SanitizeCSVField(catName),
			strconv.FormatInt(tx.AmountIDR.Int64(), 10),
			SanitizeCSVField(tx.Note),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
		csvWriter.Flush()
		return nil
	})

	if err != nil {
		return err
	}

	return csvWriter.Error()
}

