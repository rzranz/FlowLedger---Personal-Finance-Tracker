package sqlite_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"flowledger/internal/adapter/sqlite"
	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

func setupTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_flowledger.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestCategoryRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewCategoryRepository(db)
	ctx := context.Background()

	// 1. Create categories
	foodCat := &domain.Category{
		Name: "Food & Dining",
		Type: domain.CategoryTypeExpense,
	}
	if err := repo.Create(ctx, foodCat); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	if foodCat.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}

	salaryCat := &domain.Category{
		Name: "Salary",
		Type: domain.CategoryTypeIncome,
	}
	if err := repo.Create(ctx, salaryCat); err != nil {
		t.Fatalf("failed to create income category: %v", err)
	}

	// 2. Duplicate case-insensitive test
	dupFood := &domain.Category{
		Name: "food & dining", // lowercase duplicate
		Type: domain.CategoryTypeExpense,
	}
	err := repo.Create(ctx, dupFood)
	if err == nil {
		t.Fatalf("expected unique constraint conflict error on duplicate name, got nil")
	}
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected domain.ErrConflict, got: %v", err)
	}

	// 3. GetByID
	fetched, err := repo.GetByID(ctx, foodCat.ID)
	if err != nil {
		t.Fatalf("failed to get category by ID: %v", err)
	}
	if fetched.Name != foodCat.Name || fetched.Type != foodCat.Type {
		t.Fatalf("category mismatch: got %+v, want %+v", fetched, foodCat)
	}

	// 4. List with filter
	expenseType := domain.CategoryTypeExpense
	expenses, err := repo.List(ctx, &expenseType)
	if err != nil {
		t.Fatalf("failed to list categories: %v", err)
	}
	if len(expenses) != 1 || expenses[0].Name != "Food & Dining" {
		t.Fatalf("unexpected expense categories list: %+v", expenses)
	}

	// 5. Update name
	foodCat.Name = "Groceries & Food"
	if err := repo.Update(ctx, foodCat); err != nil {
		t.Fatalf("failed to update category: %v", err)
	}
	updated, err := repo.GetByID(ctx, foodCat.ID)
	if err != nil {
		t.Fatalf("failed to get updated category: %v", err)
	}
	if updated.Name != "Groceries & Food" {
		t.Fatalf("expected updated name 'Groceries & Food', got %q", updated.Name)
	}

	// 6. Delete unreferenced category
	if err := repo.Delete(ctx, salaryCat.ID); err != nil {
		t.Fatalf("failed to delete unreferenced category: %v", err)
	}
	_, err = repo.GetByID(ctx, salaryCat.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestTransactionRepository(t *testing.T) {
	db := setupTestDB(t)
	catRepo := sqlite.NewCategoryRepository(db)
	txRepo := sqlite.NewTransactionRepository(db)
	ctx := context.Background()

	cat := &domain.Category{
		Name: "Dining Out",
		Type: domain.CategoryTypeExpense,
	}
	if err := catRepo.Create(ctx, cat); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// 1. Create transaction
	tx := &domain.Transaction{
		CategoryID: cat.ID,
		Type:       domain.TransactionTypeExpense,
		AmountIDR:  domain.AmountIDR(75000),
		OccurredOn: "2026-10-15",
		Note:       "Dinner with friends",
	}
	if err := txRepo.Create(ctx, tx); err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}
	if tx.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}

	// 2. Referential integrity: cannot delete category with linked transactions
	err := catRepo.Delete(ctx, cat.ID)
	if err == nil {
		t.Fatalf("expected foreign key restriction error, got nil")
	}
	if !errors.Is(err, domain.ErrReferenced) {
		t.Fatalf("expected domain.ErrReferenced, got: %v", err)
	}

	// 3. GetByID with joined Category
	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if fetched.AmountIDR != 75000 || fetched.Category == nil || fetched.Category.Name != "Dining Out" {
		t.Fatalf("unexpected fetched transaction: %+v", fetched)
	}

	// 4. Update transaction
	tx.AmountIDR = domain.AmountIDR(85000)
	tx.Note = "Dinner with friends + dessert"
	if err := txRepo.Update(ctx, tx); err != nil {
		t.Fatalf("failed to update transaction: %v", err)
	}

	// 5. List with pagination and filter
	txs, total, err := txRepo.List(ctx, usecase.TransactionFilter{
		Month: "2026-10",
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("failed to list transactions: %v", err)
	}
	if total != 1 || len(txs) != 1 || txs[0].AmountIDR != 85000 {
		t.Fatalf("unexpected list result: total=%d, len=%d", total, len(txs))
	}

	// 6. Stream transactions
	var streamCount int
	err = txRepo.StreamTransactions(ctx, usecase.TransactionFilter{Month: "2026-10"}, func(tRow domain.Transaction) error {
		streamCount++
		if tRow.AmountIDR != 85000 {
			t.Errorf("unexpected streamed amount: %d", tRow.AmountIDR)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("streaming failed: %v", err)
	}
	if streamCount != 1 {
		t.Fatalf("expected 1 streamed row, got %d", streamCount)
	}

	// 7. Delete transaction
	if err := txRepo.Delete(ctx, tx.ID); err != nil {
		t.Fatalf("failed to delete transaction: %v", err)
	}
	_, err = txRepo.GetByID(ctx, tx.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Now category can be deleted
	if err := catRepo.Delete(ctx, cat.ID); err != nil {
		t.Fatalf("failed to delete category after transaction was deleted: %v", err)
	}
}

func TestBudgetRepository(t *testing.T) {
	db := setupTestDB(t)
	catRepo := sqlite.NewCategoryRepository(db)
	txRepo := sqlite.NewTransactionRepository(db)
	budgetRepo := sqlite.NewBudgetRepository(db)
	ctx := context.Background()

	cat := &domain.Category{
		Name: "Groceries",
		Type: domain.CategoryTypeExpense,
	}
	if err := catRepo.Create(ctx, cat); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// 1. Upsert initial budget
	budget := &domain.Budget{
		CategoryID: cat.ID,
		Month:      "2026-10",
		LimitIDR:   domain.AmountIDR(1000000), // Rp 1.000.000
	}
	if err := budgetRepo.Upsert(ctx, budget); err != nil {
		t.Fatalf("failed to upsert budget: %v", err)
	}

	// 2. Add an expense transaction
	tx := &domain.Transaction{
		CategoryID: cat.ID,
		Type:       domain.TransactionTypeExpense,
		AmountIDR:  domain.AmountIDR(850000), // Rp 850.000 (85% utilization -> WARNING)
		OccurredOn: "2026-10-05",
		Note:       "Monthly grocery shopping",
	}
	if err := txRepo.Create(ctx, tx); err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	// 3. ListWithUtilization
	items, err := budgetRepo.ListWithUtilization(ctx, "2026-10")
	if err != nil {
		t.Fatalf("failed to list budgets with utilization: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 budget item, got %d", len(items))
	}

	item := items[0]
	if item.CategoryName != "Groceries" {
		t.Errorf("expected category name 'Groceries', got %q", item.CategoryName)
	}
	if item.SpentIDR != 850000 {
		t.Errorf("expected spent 850000, got %d", item.SpentIDR)
	}
	if item.RemainingIDR != 150000 {
		t.Errorf("expected remaining 150000, got %d", item.RemainingIDR)
	}
	if item.UtilizationPct != 85 {
		t.Errorf("expected utilization 85%%, got %d%%", item.UtilizationPct)
	}
	if item.Status != domain.BudgetStatusWarning {
		t.Errorf("expected WARNING status, got %s", item.Status)
	}

	// 4. Upsert (update existing month budget)
	budget.LimitIDR = domain.AmountIDR(2000000) // Increase limit to 2.000.000 -> 42% (SAFE)
	if err := budgetRepo.Upsert(ctx, budget); err != nil {
		t.Fatalf("failed to update budget limit: %v", err)
	}

	items2, err := budgetRepo.ListWithUtilization(ctx, "2026-10")
	if err != nil {
		t.Fatalf("failed to re-fetch budgets: %v", err)
	}
	if items2[0].Status != domain.BudgetStatusSafe || items2[0].UtilizationPct != 42 {
		t.Errorf("expected SAFE and 42%%, got %s and %d%%", items2[0].Status, items2[0].UtilizationPct)
	}
}

func TestIdempotencyRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewIdempotencyRepository(db)
	ctx := context.Background()

	key := "test-key-123"
	record := &sqlite.IdempotencyRecord{
		Key:          key,
		RequestHash:  "sha256-hash-xyz",
		StatusCode:   201,
		ResponseBody: []byte(`{"id":1,"amount_idr":50000}`),
		CreatedAt:    time.Now().UTC(),
	}

	if err := repo.Save(ctx, record); err != nil {
		t.Fatalf("failed to save idempotency record: %v", err)
	}

	fetched, err := repo.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get idempotency record: %v", err)
	}
	if fetched.StatusCode != 201 || string(fetched.ResponseBody) != string(record.ResponseBody) {
		t.Fatalf("record mismatch: %+v", fetched)
	}

	// Duplicate key conflict
	err = repo.Save(ctx, record)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict on duplicate key insert, got %v", err)
	}
}

func TestReportRepository(t *testing.T) {
	db := setupTestDB(t)
	catRepo := sqlite.NewCategoryRepository(db)
	txRepo := sqlite.NewTransactionRepository(db)
	reportRepo := sqlite.NewReportRepository(db)
	ctx := context.Background()

	salaryCat := &domain.Category{Name: "Salary", Type: domain.CategoryTypeIncome}
	foodCat := &domain.Category{Name: "Food", Type: domain.CategoryTypeExpense}
	_ = catRepo.Create(ctx, salaryCat)
	_ = catRepo.Create(ctx, foodCat)

	// Add income
	_ = txRepo.Create(ctx, &domain.Transaction{
		CategoryID: salaryCat.ID,
		Type:       domain.TransactionTypeIncome,
		AmountIDR:  domain.AmountIDR(10000000),
		OccurredOn: "2026-10-01",
	})
	// Add expense
	_ = txRepo.Create(ctx, &domain.Transaction{
		CategoryID: foodCat.ID,
		Type:       domain.TransactionTypeExpense,
		AmountIDR:  domain.AmountIDR(2500000),
		OccurredOn: "2026-10-05",
	})

	// 1. GetMonthlyTotals
	income, expense, net, err := reportRepo.GetMonthlyTotals(ctx, "2026-10")
	if err != nil {
		t.Fatalf("failed to get monthly totals: %v", err)
	}
	if income != 10000000 || expense != 2500000 || net != 7500000 {
		t.Errorf("unexpected totals: income=%d, expense=%d, net=%d", income, expense, net)
	}

	// 2. GetCategoryBreakdown
	breakdown, err := reportRepo.GetCategoryBreakdown(ctx, "2026-10", expense)
	if err != nil {
		t.Fatalf("failed to get category breakdown: %v", err)
	}
	if len(breakdown) != 1 || breakdown[0].AmountIDR != 2500000 || breakdown[0].Percentage != 100 {
		t.Errorf("unexpected breakdown: %+v", breakdown)
	}

	// 3. GetPreviousMonth helper
	prev, err := sqlite.GetPreviousMonth("2026-10")
	if err != nil || prev != "2026-09" {
		t.Errorf("expected 2026-09, got %q (err=%v)", prev, err)
	}
	_, err = sqlite.GetPreviousMonth("bad-date")
	if err == nil {
		t.Errorf("expected error on malformed month")
	}

	// 4. Test MapSQLError branches
	if sqlite.MapSQLError(nil) != nil {
		t.Errorf("expected nil error")
	}
	if !errors.Is(sqlite.MapSQLError(errors.New("CHECK constraint failed: amount_idr > 0")), domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput")
	}
}

func TestSeedRealisticIndonesianData(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// First run seeds data
	if err := sqlite.SeedRealisticIndonesianData(ctx, db); err != nil {
		t.Fatalf("seeding failed: %v", err)
	}

	// Verify categories were inserted
	catRepo := sqlite.NewCategoryRepository(db)
	cats, err := catRepo.List(ctx, nil)
	if err != nil || len(cats) != 9 {
		t.Fatalf("expected 9 categories seeded, got %d (err: %v)", len(cats), err)
	}

	// Second run is idempotent (does not duplicate)
	if err := sqlite.SeedRealisticIndonesianData(ctx, db); err != nil {
		t.Fatalf("second seeding run failed: %v", err)
	}
	catsAfter, _ := catRepo.List(ctx, nil)
	if len(catsAfter) != 9 {
		t.Fatalf("expected category count to remain 9 after second run, got %d", len(catsAfter))
	}
}


