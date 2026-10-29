package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"flowledger/internal/adapter/httpapi"
	"flowledger/internal/adapter/sqlite"
	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

func setupTestServer(t *testing.T) (http.Handler, *sqlite.DB) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "flowledger_test.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	catRepo := sqlite.NewCategoryRepository(db)
	txRepo := sqlite.NewTransactionRepository(db)
	budgetRepo := sqlite.NewBudgetRepository(db)
	idempotencyRepo := sqlite.NewIdempotencyRepository(db)
	reportRepo := sqlite.NewReportRepository(db)

	catUc := usecase.NewCategoryUsecase(catRepo)
	txUc := usecase.NewTransactionUsecase(txRepo, catRepo)
	budgetUc := usecase.NewBudgetUsecase(budgetRepo, catRepo)
	reportUc := usecase.NewReportUsecase(reportRepo, budgetRepo)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := httpapi.NewRouter(httpapi.RouterConfig{
		Logger:          logger,
		DB:              db,
		CategoryUsecase: catUc,
		TxUsecase:       txUc,
		BudgetUsecase:   budgetUc,
		ReportUsecase:   reportUc,
		IdempotencyRepo: idempotencyRepo,
	})

	return handler, db
}

func TestProbes(t *testing.T) {
	h, _ := setupTestServer(t)

	// Liveness
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /health, got %d", rec.Code)
	}

	// Readiness
	req = httptest.NewRequest("GET", "/ready", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /ready, got %d", rec.Code)
	}
}

func TestCategoryEndpoints(t *testing.T) {
	h, _ := setupTestServer(t)

	// 1. Create Category
	body := `{"name":"Groceries","type":"expense"}`
	req := httptest.NewRequest("POST", "/api/v1/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var created domain.Category
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if created.ID == 0 || created.Name != "Groceries" {
		t.Fatalf("unexpected category: %+v", created)
	}

	// 2. Reject duplicate category with 409 Conflict
	dupBody := `{"name":"groceries","type":"expense"}`
	req = httptest.NewRequest("POST", "/api/v1/categories", strings.NewReader(dupBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate category, got %d: %s", rec.Code, rec.Body.String())
	}

	// 3. Reject unknown JSON fields with 400 Bad Request
	unknownFieldBody := `{"name":"Salary","type":"income","unknown_prop":"hack"}`
	req = httptest.NewRequest("POST", "/api/v1/categories", strings.NewReader(unknownFieldBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on unknown JSON fields, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Update Category
	updateBody := `{"name":"Supermarket"}`
	req = httptest.NewRequest("PATCH", "/api/v1/categories/1", strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK updating category, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. List with type filter
	req = httptest.NewRequest("GET", "/api/v1/categories?type=expense", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing categories, got %d", rec.Code)
	}

	// 6. Delete Category
	req = httptest.NewRequest("DELETE", "/api/v1/categories/1", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content deleting category, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTransactionEndpoints_And_Idempotency(t *testing.T) {
	h, _ := setupTestServer(t)

	// Setup Category
	catBody := `{"name":"Transportation","type":"expense"}`
	req := httptest.NewRequest("POST", "/api/v1/categories", strings.NewReader(catBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var cat domain.Category
	_ = json.Unmarshal(rec.Body.Bytes(), &cat)

	// 1. Create Transaction with Idempotency-Key
	idemKey := "tx-client-uuid-001"
	txBody := `{"category_id":1,"type":"expense","amount_idr":45000,"occurred_on":"2026-10-20","note":"Train ticket"}`

	req = httptest.NewRequest("POST", "/api/v1/transactions", strings.NewReader(txBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idemKey)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Re-send identical request with same Idempotency-Key -> should replay cached 201
	req = httptest.NewRequest("POST", "/api/v1/transactions", strings.NewReader(txBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idemKey)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created from idempotent cache, got %d", rec.Code)
	}
	if rec.Header().Get("X-Cache-Lookup") != "HIT" {
		t.Errorf("expected X-Cache-Lookup: HIT header")
	}

	// 3. Re-send with same Idempotency-Key but different payload -> 409 Conflict
	tamperedBody := `{"category_id":1,"type":"expense","amount_idr":99000,"occurred_on":"2026-10-20","note":"Different"}`
	req = httptest.NewRequest("POST", "/api/v1/transactions", strings.NewReader(tamperedBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idemKey)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on payload mismatch for same key, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Verify only 1 transaction actually exists in DB
	req = httptest.NewRequest("GET", "/api/v1/transactions", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var listResp struct {
		Items      []domain.Transaction `json:"items"`
		Pagination struct {
			Total int64 `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to parse list transactions: %v", err)
	}
	if listResp.Pagination.Total != 1 {
		t.Fatalf("expected exactly 1 transaction in DB, found %d", listResp.Pagination.Total)
	}
}

func TestBudgetAndReportEndpoints(t *testing.T) {
	h, _ := setupTestServer(t)

	// Create Expense Category
	catBody := `{"name":"Dining","type":"expense"}`
	req := httptest.NewRequest("POST", "/api/v1/categories", strings.NewReader(catBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// Create Budget for Dining in 2026-10 (limit: 500.000)
	budgetBody := `{"limit_idr":500000}`
	req = httptest.NewRequest("PUT", "/api/v1/budgets/1/2026-10", strings.NewReader(budgetBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK upserting budget, got %d: %s", rec.Code, rec.Body.String())
	}

	// Add an expense transaction of 450.000 (90% utilization -> WARNING)
	txBody := `{"category_id":1,"type":"expense","amount_idr":450000,"occurred_on":"2026-10-10","note":"Team lunch"}`
	req = httptest.NewRequest("POST", "/api/v1/transactions", strings.NewReader(txBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create tx: %d: %s", rec.Code, rec.Body.String())
	}

	// Check Budget List Endpoint
	req = httptest.NewRequest("GET", "/api/v1/budgets?month=2026-10", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/v1/budgets, got %d", rec.Code)
	}

	var budgetListResp struct {
		Items []domain.BudgetReportItem `json:"items"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &budgetListResp)
	if len(budgetListResp.Items) != 1 || budgetListResp.Items[0].UtilizationPct != 90 {
		t.Fatalf("unexpected budget list items: %+v", budgetListResp.Items)
	}

	// Check Monthly Report Endpoint
	req = httptest.NewRequest("GET", "/api/v1/reports/monthly?month=2026-10", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/v1/reports/monthly, got %d: %s", rec.Code, rec.Body.String())
	}

	var report usecase.MonthlyReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to parse monthly report: %v", err)
	}

	if report.TotalExpenseIDR != 450000 {
		t.Errorf("expected total expense 450000, got %d", report.TotalExpenseIDR)
	}

	// Test CSV Export
	req = httptest.NewRequest("GET", "/api/v1/transactions/export.csv", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from CSV export, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Errorf("expected Content-Type text/csv, got %s", rec.Header().Get("Content-Type"))
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Team lunch")) {
		t.Errorf("expected CSV body to contain transaction note")
	}

	// 5. GET /api/v1/transactions/{id}
	req = httptest.NewRequest("GET", "/api/v1/transactions/1", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from GET tx, got %d", rec.Code)
	}

	// 6. PATCH /api/v1/transactions/{id}
	patchBody := `{"amount_idr":480000}`
	req = httptest.NewRequest("PATCH", "/api/v1/transactions/1", strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from PATCH tx, got %d", rec.Code)
	}

	// 7. DELETE /api/v1/transactions/{id}
	req = httptest.NewRequest("DELETE", "/api/v1/transactions/1", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content from DELETE tx, got %d", rec.Code)
	}

	// 8. Test CORS preflight
	req = httptest.NewRequest("OPTIONS", "/api/v1/categories", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content from CORS OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *")
	}
}
