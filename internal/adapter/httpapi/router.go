package httpapi

import (
	"log/slog"
	"net/http"

	"flowledger/internal/adapter/sqlite"
	"flowledger/internal/usecase"
)

// RouterConfig contains dependencies for building the HTTP router.
type RouterConfig struct {
	Logger          *slog.Logger
	DB              *sqlite.DB
	CategoryUsecase *usecase.CategoryUsecase
	TxUsecase       *usecase.TransactionUsecase
	BudgetUsecase   *usecase.BudgetUsecase
	ReportUsecase   *usecase.ReportUsecase
	IdempotencyRepo *sqlite.IdempotencyRepository
}

// Middleware is a standard HTTP middleware type.
type Middleware func(http.Handler) http.Handler

// Chain compiles multiple middlewares into a single wrapped http.Handler.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// NewRouter constructs and configures the Go 1.22+ ServeMux with all routes and middlewares.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	probesHandler := NewProbesHandler(cfg.DB)
	categoryHandler := NewCategoryHandler(cfg.CategoryUsecase)
	txHandler := NewTransactionHandler(cfg.TxUsecase)
	budgetHandler := NewBudgetHandler(cfg.BudgetUsecase)
	reportHandler := NewReportHandler(cfg.ReportUsecase)
	exportHandler := NewExportHandler(cfg.TxUsecase)

	// Liveness & Readiness Probes
	mux.HandleFunc("GET /health", probesHandler.Health)
	mux.HandleFunc("GET /ready", probesHandler.Ready)

	// Category endpoints
	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("POST /api/v1/categories", categoryHandler.Create)
	mux.HandleFunc("PATCH /api/v1/categories/{id}", categoryHandler.Update)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", categoryHandler.Delete)

	// Transaction endpoints & CSV Export
	mux.HandleFunc("GET /api/v1/transactions", txHandler.List)
	mux.HandleFunc("POST /api/v1/transactions", txHandler.Create)
	mux.HandleFunc("GET /api/v1/transactions/export.csv", exportHandler.ExportCSV)
	mux.HandleFunc("GET /api/v1/transactions/{id}", txHandler.GetByID)
	mux.HandleFunc("PATCH /api/v1/transactions/{id}", txHandler.Update)
	mux.HandleFunc("DELETE /api/v1/transactions/{id}", txHandler.Delete)

	// Budget endpoints
	mux.HandleFunc("GET /api/v1/budgets", budgetHandler.List)
	mux.HandleFunc("PUT /api/v1/budgets/{category_id}/{month}", budgetHandler.Upsert)

	// Analytics report
	mux.HandleFunc("GET /api/v1/reports/monthly", reportHandler.Monthly)

	// Build global middleware chain
	globalMiddlewares := []Middleware{
		RequestID,
		Logger(cfg.Logger),
		Recoverer(cfg.Logger),
		CORS,
		IdempotencyMiddleware(cfg.IdempotencyRepo),
	}

	return Chain(mux, globalMiddlewares...)
}

