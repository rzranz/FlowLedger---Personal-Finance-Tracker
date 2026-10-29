package sqlite

import (
	"context"
	"fmt"
	"time"

	"flowledger/internal/domain"
)

// CategoryRepository implements database operations for categories.
type CategoryRepository struct {
	db *DB
}

// NewCategoryRepository creates a new instance of CategoryRepository.
func NewCategoryRepository(db *DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create inserts a new Category into SQLite.
func (r *CategoryRepository) Create(ctx context.Context, cat *domain.Category) error {
	now := time.Now().UTC()
	query := `
	INSERT INTO categories (name, type, created_at)
	VALUES (?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query, cat.Name, cat.Type, now.Format(time.RFC3339))
	if err != nil {
		return MapSQLError(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve last insert id: %w", err)
	}

	cat.ID = id
	cat.CreatedAt = now
	return nil
}

// GetByID finds a Category by its ID.
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	query := `SELECT id, name, type, created_at FROM categories WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var cat domain.Category
	var createdAtStr string
	err := row.Scan(&cat.ID, &cat.Name, &cat.Type, &createdAtStr)
	if err != nil {
		return nil, MapSQLError(err)
	}

	t, _ := time.Parse(time.RFC3339, createdAtStr)
	cat.CreatedAt = t
	return &cat, nil
}

// List retrieves all categories, optionally filtered by type.
func (r *CategoryRepository) List(ctx context.Context, catType *domain.CategoryType) ([]domain.Category, error) {
	var query string
	var args []any

	if catType != nil && catType.IsValid() {
		query = `SELECT id, name, type, created_at FROM categories WHERE type = ? ORDER BY name ASC`
		args = append(args, *catType)
	} else {
		query = `SELECT id, name, type, created_at FROM categories ORDER BY type ASC, name ASC`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, MapSQLError(err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		var createdAtStr string
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Type, &createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		t, _ := time.Parse(time.RFC3339, createdAtStr)
		cat.CreatedAt = t
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if categories == nil {
		categories = []domain.Category{}
	}
	return categories, nil
}

// Update updates the category's name.
func (r *CategoryRepository) Update(ctx context.Context, cat *domain.Category) error {
	query := `UPDATE categories SET name = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, cat.Name, cat.ID)
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

// Delete removes a category if it is not referenced by any transactions or budgets.
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = ?`
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

// HasReferences checks whether a category is linked to transactions or budgets.
func (r *CategoryRepository) HasReferences(ctx context.Context, id int64) (bool, error) {
	var txCount, budgetCount int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE category_id = ?`, id).Scan(&txCount)
	if err != nil {
		return false, MapSQLError(err)
	}
	if txCount > 0 {
		return true, nil
	}

	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM budgets WHERE category_id = ?`, id).Scan(&budgetCount)
	if err != nil {
		return false, MapSQLError(err)
	}
	return budgetCount > 0, nil
}

