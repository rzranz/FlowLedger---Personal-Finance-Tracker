package sqlite

import (
	"context"
	"time"
)

// IdempotencyRecord stores cached HTTP response data for an idempotency key.
type IdempotencyRecord struct {
	Key          string
	RequestHash  string
	StatusCode   int
	ResponseBody []byte
	CreatedAt    time.Time
}

// IdempotencyRepository manages idempotency key persistence.
type IdempotencyRepository struct {
	db *DB
}

// NewIdempotencyRepository creates a new IdempotencyRepository.
func NewIdempotencyRepository(db *DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

// Get finds an idempotency record by key.
func (r *IdempotencyRepository) Get(ctx context.Context, key string) (*IdempotencyRecord, error) {
	query := `
	SELECT key, request_hash, status_code, response_body, created_at
	FROM idempotency_keys
	WHERE key = ?
	`
	row := r.db.QueryRowContext(ctx, query, key)

	var rec IdempotencyRecord
	var createdAtStr string
	err := row.Scan(&rec.Key, &rec.RequestHash, &rec.StatusCode, &rec.ResponseBody, &createdAtStr)
	if err != nil {
		return nil, MapSQLError(err)
	}

	rec.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return &rec, nil
}

// Save stores the idempotency record.
func (r *IdempotencyRepository) Save(ctx context.Context, rec *IdempotencyRecord) error {
	now := time.Now().UTC()
	query := `
	INSERT INTO idempotency_keys (key, request_hash, status_code, response_body, created_at)
	VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		rec.Key,
		rec.RequestHash,
		rec.StatusCode,
		rec.ResponseBody,
		now.Format(time.RFC3339),
	)
	if err != nil {
		return MapSQLError(err)
	}

	rec.CreatedAt = now
	return nil
}

