package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

// TransactionRepository implements transaction database operations.
type TransactionRepository struct {
	db *DB
}

// NewTransactionRepository creates a new TransactionRepository.
func NewTransactionRepository(db *DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a new Transaction.
func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	now := time.Now().UTC()
	query := `
	INSERT INTO transactions (category_id, type, amount_idr, occurred_on, note, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query,
		tx.CategoryID,
		tx.Type,
		tx.AmountIDR.Int64(),
		tx.OccurredOn,
		tx.Note,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return MapSQLError(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve last insert id: %w", err)
	}

	tx.ID = id
	tx.CreatedAt = now
	tx.UpdatedAt = now
	return nil
}

// GetByID finds a Transaction with its Category by ID.
func (r *TransactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	query := `
	SELECT 
		t.id, t.category_id, t.type, t.amount_idr, t.occurred_on, t.note, t.created_at, t.updated_at,
		c.id, c.name, c.type, c.created_at
	FROM transactions t
	JOIN categories c ON t.category_id = c.id
	WHERE t.id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var tx domain.Transaction
	var cat domain.Category
	var txCreatedAt, txUpdatedAt, catCreatedAt string
	var rawAmount int64

	err := row.Scan(
		&tx.ID, &tx.CategoryID, &tx.Type, &rawAmount, &tx.OccurredOn, &tx.Note, &txCreatedAt, &txUpdatedAt,
		&cat.ID, &cat.Name, &cat.Type, &catCreatedAt,
	)
	if err != nil {
		return nil, MapSQLError(err)
	}

	tx.AmountIDR = domain.AmountIDR(rawAmount)
	tx.CreatedAt, _ = time.Parse(time.RFC3339, txCreatedAt)
	tx.UpdatedAt, _ = time.Parse(time.RFC3339, txUpdatedAt)
	cat.CreatedAt, _ = time.Parse(time.RFC3339, catCreatedAt)
	tx.Category = &cat

	return &tx, nil
}

// List queries transactions with filtering and pagination.
func (r *TransactionRepository) List(ctx context.Context, filter usecase.TransactionFilter) ([]domain.Transaction, int64, error) {
	whereClauses := []string{"1=1"}
	var args []any

	if filter.Month != "" {
		whereClauses = append(whereClauses, "substr(t.occurred_on, 1, 7) = ?")
		args = append(args, filter.Month)
	}
	if filter.Type != nil && filter.Type.IsValid() {
		whereClauses = append(whereClauses, "t.type = ?")
		args = append(args, *filter.Type)
	}
	if filter.CategoryID != nil && *filter.CategoryID > 0 {
		whereClauses = append(whereClauses, "t.category_id = ?")
		args = append(args, *filter.CategoryID)
	}
	if filter.Query != "" {
		whereClauses = append(whereClauses, "t.note LIKE ?")
		args = append(args, "%"+filter.Query+"%")
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Total count query
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM transactions t WHERE %s", whereSQL)
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, MapSQLError(err)
	}

	// Pagination defaults
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
	SELECT 
		t.id, t.category_id, t.type, t.amount_idr, t.occurred_on, t.note, t.created_at, t.updated_at,
		c.id, c.name, c.type, c.created_at
	FROM transactions t
	JOIN categories c ON t.category_id = c.id
	WHERE %s
	ORDER BY t.occurred_on DESC, t.id DESC
	LIMIT ? OFFSET ?
	`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, MapSQLError(err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		var cat domain.Category
		var txCreatedAt, txUpdatedAt, catCreatedAt string
		var rawAmount int64

		err := rows.Scan(
			&tx.ID, &tx.CategoryID, &tx.Type, &rawAmount, &tx.OccurredOn, &tx.Note, &txCreatedAt, &txUpdatedAt,
			&cat.ID, &cat.Name, &cat.Type, &catCreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}

		tx.AmountIDR = domain.AmountIDR(rawAmount)
		tx.CreatedAt, _ = time.Parse(time.RFC3339, txCreatedAt)
		tx.UpdatedAt, _ = time.Parse(time.RFC3339, txUpdatedAt)
		cat.CreatedAt, _ = time.Parse(time.RFC3339, catCreatedAt)
		tx.Category = &cat

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	if transactions == nil {
		transactions = []domain.Transaction{}
	}

	return transactions, total, nil
}

// Update updates an existing Transaction.
func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	now := time.Now().UTC()
	query := `
	UPDATE transactions
	SET category_id = ?, type = ?, amount_idr = ?, occurred_on = ?, note = ?, updated_at = ?
	WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		tx.CategoryID,
		tx.Type,
		tx.AmountIDR.Int64(),
		tx.OccurredOn,
		tx.Note,
		now.Format(time.RFC3339),
		tx.ID,
	)
	if err != nil {
		return MapSQLError(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	tx.UpdatedAt = now
	return nil
}

// Delete removes a Transaction by ID.
func (r *TransactionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM transactions WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return MapSQLError(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// StreamTransactions iterates through transactions matching the filter one by one,
// invoking fn for each row without buffering the entire result set in memory.
func (r *TransactionRepository) StreamTransactions(ctx context.Context, filter usecase.TransactionFilter, fn func(tx domain.Transaction) error) error {
	whereClauses := []string{"1=1"}
	var args []any

	if filter.Month != "" {
		whereClauses = append(whereClauses, "substr(t.occurred_on, 1, 7) = ?")
		args = append(args, filter.Month)
	}
	if filter.Type != nil && filter.Type.IsValid() {
		whereClauses = append(whereClauses, "t.type = ?")
		args = append(args, *filter.Type)
	}
	if filter.CategoryID != nil && *filter.CategoryID > 0 {
		whereClauses = append(whereClauses, "t.category_id = ?")
		args = append(args, *filter.CategoryID)
	}
	if filter.Query != "" {
		whereClauses = append(whereClauses, "t.note LIKE ?")
		args = append(args, "%"+filter.Query+"%")
	}

	whereSQL := strings.Join(whereClauses, " AND ")
	query := fmt.Sprintf(`
	SELECT 
		t.id, t.category_id, t.type, t.amount_idr, t.occurred_on, t.note, t.created_at, t.updated_at,
		c.id, c.name, c.type, c.created_at
	FROM transactions t
	JOIN categories c ON t.category_id = c.id
	WHERE %s
	ORDER BY t.occurred_on DESC, t.id DESC
	`, whereSQL)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return MapSQLError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var tx domain.Transaction
		var cat domain.Category
		var txCreatedAt, txUpdatedAt, catCreatedAt string
		var rawAmount int64

		err := rows.Scan(
			&tx.ID, &tx.CategoryID, &tx.Type, &rawAmount, &tx.OccurredOn, &tx.Note, &txCreatedAt, &txUpdatedAt,
			&cat.ID, &cat.Name, &cat.Type, &catCreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan streamed transaction: %w", err)
		}

		tx.AmountIDR = domain.AmountIDR(rawAmount)
		tx.CreatedAt, _ = time.Parse(time.RFC3339, txCreatedAt)
		tx.UpdatedAt, _ = time.Parse(time.RFC3339, txUpdatedAt)
		cat.CreatedAt, _ = time.Parse(time.RFC3339, catCreatedAt)
		tx.Category = &cat

		if err := fn(tx); err != nil {
			return err
		}
	}

	return rows.Err()
}

