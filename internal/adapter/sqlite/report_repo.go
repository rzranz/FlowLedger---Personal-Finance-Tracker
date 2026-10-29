package sqlite

import (
	"context"
	"fmt"
	"time"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

// ReportRepository handles analytical SQL queries.
type ReportRepository struct {
	db *DB
}

// NewReportRepository creates a new ReportRepository.
func NewReportRepository(db *DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// GetMonthlyTotals calculates total income, expense, and net cash flow for a month.
func (r *ReportRepository) GetMonthlyTotals(ctx context.Context, month string) (income, expense, net domain.AmountIDR, err error) {
	query := `
	SELECT 
		COALESCE(SUM(CASE WHEN type = 'income' THEN amount_idr ELSE 0 END), 0) AS total_income,
		COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_idr ELSE 0 END), 0) AS total_expense
	FROM transactions
	WHERE substr(occurred_on, 1, 7) = ?
	`
	var incomeRaw, expenseRaw int64
	err = r.db.QueryRowContext(ctx, query, month).Scan(&incomeRaw, &expenseRaw)
	if err != nil {
		return 0, 0, 0, MapSQLError(err)
	}

	income = domain.AmountIDR(incomeRaw)
	expense = domain.AmountIDR(expenseRaw)
	net = income.Sub(expense)

	return income, expense, net, nil
}

// GetCategoryBreakdown returns expenses grouped by category for a month, ordered by highest spend.
func (r *ReportRepository) GetCategoryBreakdown(ctx context.Context, month string, totalExpense domain.AmountIDR) ([]usecase.CategorySpendItem, error) {
	query := `
	SELECT 
		c.id, c.name, COALESCE(SUM(t.amount_idr), 0) AS spent
	FROM categories c
	JOIN transactions t ON t.category_id = c.id
	WHERE substr(t.occurred_on, 1, 7) = ? AND t.type = 'expense'
	GROUP BY c.id, c.name
	ORDER BY spent DESC
	`
	rows, err := r.db.QueryContext(ctx, query, month)
	if err != nil {
		return nil, MapSQLError(err)
	}
	defer rows.Close()

	var items []usecase.CategorySpendItem
	for rows.Next() {
		var item usecase.CategorySpendItem
		var rawSpent int64
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &rawSpent); err != nil {
			return nil, fmt.Errorf("failed to scan category breakdown: %w", err)
		}
		item.AmountIDR = domain.AmountIDR(rawSpent)
		if totalExpense.Int64() > 0 {
			item.Percentage = int((item.AmountIDR.Int64() * 100) / totalExpense.Int64())
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if items == nil {
		items = []usecase.CategorySpendItem{}
	}
	return items, nil
}

// GetPreviousMonth returns the YYYY-MM string of the preceding calendar month.
func GetPreviousMonth(month string) (string, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "", fmt.Errorf("invalid month format: %w", err)
	}
	prev := t.AddDate(0, -1, 0)
	return prev.Format("2006-01"), nil
}

