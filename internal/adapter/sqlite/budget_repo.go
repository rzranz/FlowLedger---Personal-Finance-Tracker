package sqlite

import (
	"context"
	"fmt"
	"time"

	"flowledger/internal/domain"
)

// BudgetRepository implements SQLite operations for monthly category budgets.
type BudgetRepository struct {
	db *DB
}

// NewBudgetRepository creates a new BudgetRepository.
func NewBudgetRepository(db *DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

// Upsert inserts or updates a monthly budget for a category.
func (r *BudgetRepository) Upsert(ctx context.Context, b *domain.Budget) error {
	now := time.Now().UTC()
	query := `
	INSERT INTO budgets (category_id, month, limit_idr, updated_at)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(category_id, month) DO UPDATE SET
		limit_idr = excluded.limit_idr,
		updated_at = excluded.updated_at
	`
	res, err := r.db.ExecContext(ctx, query,
		b.CategoryID,
		b.Month,
		b.LimitIDR.Int64(),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return MapSQLError(err)
	}

	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		b.ID = id
	}
	b.UpdatedAt = now
	return nil
}

// GetByCategoryAndMonth fetches the budget for a specific category and month.
func (r *BudgetRepository) GetByCategoryAndMonth(ctx context.Context, categoryID int64, month string) (*domain.Budget, error) {
	query := `
	SELECT id, category_id, month, limit_idr, updated_at
	FROM budgets
	WHERE category_id = ? AND month = ?
	`
	row := r.db.QueryRowContext(ctx, query, categoryID, month)

	var b domain.Budget
	var rawLimit int64
	var updatedAtStr string

	if err := row.Scan(&b.ID, &b.CategoryID, &b.Month, &rawLimit, &updatedAtStr); err != nil {
		return nil, MapSQLError(err)
	}

	b.LimitIDR = domain.AmountIDR(rawLimit)
	b.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &b, nil
}

// ListWithUtilization retrieves all budgets for the given month with calculated actual spending.
func (r *BudgetRepository) ListWithUtilization(ctx context.Context, month string) ([]domain.BudgetReportItem, error) {
	query := `
	SELECT 
		b.id, b.category_id, b.month, b.limit_idr, b.updated_at,
		c.name,
		COALESCE(SUM(t.amount_idr), 0) AS spent_idr
	FROM budgets b
	JOIN categories c ON b.category_id = c.id
	LEFT JOIN transactions t ON t.category_id = b.category_id 
		AND substr(t.occurred_on, 1, 7) = b.month 
		AND t.type = 'expense'
	WHERE b.month = ?
	GROUP BY b.id, b.category_id, b.month, b.limit_idr, b.updated_at, c.name
	ORDER BY c.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, month)
	if err != nil {
		return nil, MapSQLError(err)
	}
	defer rows.Close()

	var items []domain.BudgetReportItem
	for rows.Next() {
		var item domain.BudgetReportItem
		var rawLimit, rawSpent int64
		var updatedAtStr string

		err := rows.Scan(
			&item.Budget.ID,
			&item.Budget.CategoryID,
			&item.Budget.Month,
			&rawLimit,
			&updatedAtStr,
			&item.CategoryName,
			&rawSpent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget report item: %w", err)
		}

		item.Budget.LimitIDR = domain.AmountIDR(rawLimit)
		item.Budget.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		item.SpentIDR = domain.AmountIDR(rawSpent)

		remaining, pct, status := domain.CalculateUtilization(item.Budget.LimitIDR, item.SpentIDR)
		item.RemainingIDR = remaining
		item.UtilizationPct = pct
		item.Status = status

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if items == nil {
		items = []domain.BudgetReportItem{}
	}
	return items, nil
}

