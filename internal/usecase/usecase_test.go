package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"flowledger/internal/domain"
	"flowledger/internal/usecase"
)

// MockCategoryStore implements usecase.CategoryStore for testing.
type MockCategoryStore struct {
	categories map[int64]*domain.Category
	referenced map[int64]bool
	nextID     int64
}

func newMockCategoryStore() *MockCategoryStore {
	return &MockCategoryStore{
		categories: make(map[int64]*domain.Category),
		referenced: make(map[int64]bool),
		nextID:     1,
	}
}

func (m *MockCategoryStore) Create(ctx context.Context, cat *domain.Category) error {
	for _, c := range m.categories {
		if strings.EqualFold(c.Name, cat.Name) && c.Type == cat.Type {
			return domain.ErrConflict
		}
	}
	cat.ID = m.nextID
	m.nextID++
	copied := *cat
	m.categories[cat.ID] = &copied
	return nil
}

func (m *MockCategoryStore) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	cat, ok := m.categories[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *cat
	return &copied, nil
}

func (m *MockCategoryStore) List(ctx context.Context, catType *domain.CategoryType) ([]domain.Category, error) {
	var res []domain.Category
	for _, cat := range m.categories {
		if catType == nil || cat.Type == *catType {
			res = append(res, *cat)
		}
	}
	return res, nil
}

func (m *MockCategoryStore) Update(ctx context.Context, cat *domain.Category) error {
	if _, ok := m.categories[cat.ID]; !ok {
		return domain.ErrNotFound
	}
	copied := *cat
	m.categories[cat.ID] = &copied
	return nil
}

func (m *MockCategoryStore) Delete(ctx context.Context, id int64) error {
	if _, ok := m.categories[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.categories, id)
	return nil
}

func (m *MockCategoryStore) HasReferences(ctx context.Context, id int64) (bool, error) {
	return m.referenced[id], nil
}

// MockTransactionStore implements usecase.TransactionStore.
type MockTransactionStore struct {
	transactions map[int64]*domain.Transaction
	nextID       int64
}

func newMockTransactionStore() *MockTransactionStore {
	return &MockTransactionStore{
		transactions: make(map[int64]*domain.Transaction),
		nextID:       1,
	}
}

func (m *MockTransactionStore) Create(ctx context.Context, tx *domain.Transaction) error {
	tx.ID = m.nextID
	m.nextID++
	copied := *tx
	m.transactions[tx.ID] = &copied
	return nil
}

func (m *MockTransactionStore) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	tx, ok := m.transactions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *tx
	return &copied, nil
}

func (m *MockTransactionStore) List(ctx context.Context, filter usecase.TransactionFilter) ([]domain.Transaction, int64, error) {
	var list []domain.Transaction
	for _, tx := range m.transactions {
		list = append(list, *tx)
	}
	return list, int64(len(list)), nil
}

func (m *MockTransactionStore) Update(ctx context.Context, tx *domain.Transaction) error {
	if _, ok := m.transactions[tx.ID]; !ok {
		return domain.ErrNotFound
	}
	copied := *tx
	m.transactions[tx.ID] = &copied
	return nil
}

func (m *MockTransactionStore) Delete(ctx context.Context, id int64) error {
	if _, ok := m.transactions[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.transactions, id)
	return nil
}

func (m *MockTransactionStore) StreamTransactions(ctx context.Context, filter usecase.TransactionFilter, fn func(tx domain.Transaction) error) error {
	for _, tx := range m.transactions {
		if err := fn(*tx); err != nil {
			return err
		}
	}
	return nil
}

// MockBudgetStore implements usecase.BudgetStore.
type MockBudgetStore struct {
	budgets map[string]*domain.Budget // key: categoryID:month
	items   []domain.BudgetReportItem
}

func newMockBudgetStore() *MockBudgetStore {
	return &MockBudgetStore{
		budgets: make(map[string]*domain.Budget),
	}
}

func (m *MockBudgetStore) Upsert(ctx context.Context, b *domain.Budget) error {
	key := strings.Join([]string{string(rune(b.CategoryID)), b.Month}, ":")
	copied := *b
	m.budgets[key] = &copied
	return nil
}

func (m *MockBudgetStore) GetByCategoryAndMonth(ctx context.Context, categoryID int64, month string) (*domain.Budget, error) {
	key := strings.Join([]string{string(rune(categoryID)), month}, ":")
	b, ok := m.budgets[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copied := *b
	return &copied, nil
}

func (m *MockBudgetStore) ListWithUtilization(ctx context.Context, month string) ([]domain.BudgetReportItem, error) {
	return m.items, nil
}

// MockReportStore implements usecase.ReportStore.
type MockReportStore struct {
	monthlyTotals map[string][3]domain.AmountIDR // month -> [income, expense, net]
	breakdown     map[string][]usecase.CategorySpendItem
}

func (m *MockReportStore) GetMonthlyTotals(ctx context.Context, month string) (domain.AmountIDR, domain.AmountIDR, domain.AmountIDR, error) {
	vals := m.monthlyTotals[month]
	return vals[0], vals[1], vals[2], nil
}

func (m *MockReportStore) GetCategoryBreakdown(ctx context.Context, month string, totalExpense domain.AmountIDR) ([]usecase.CategorySpendItem, error) {
	return m.breakdown[month], nil
}

func TestCategoryUsecase(t *testing.T) {
	store := newMockCategoryStore()
	uc := usecase.NewCategoryUsecase(store)
	ctx := context.Background()

	// 1. Create category
	cat, err := uc.Create(ctx, "Food & Dining", domain.CategoryTypeExpense)
	if err != nil {
		t.Fatalf("unexpected error creating category: %v", err)
	}
	if cat.ID != 1 {
		t.Errorf("expected ID 1, got %d", cat.ID)
	}

	// 2. Duplicate category error
	_, err = uc.Create(ctx, "food & dining", domain.CategoryTypeExpense)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict on duplicate category name, got %v", err)
	}

	// 3. Validation failure (empty name)
	_, err = uc.Create(ctx, "   ", domain.CategoryTypeExpense)
	if err == nil {
		t.Fatalf("expected validation error on empty category name")
	}

	// 4. GetByID
	fetched, err := uc.GetByID(ctx, cat.ID)
	if err != nil || fetched.Name != "Food & Dining" {
		t.Fatalf("failed to get category: %v", err)
	}

	// 5. List categories
	list, err := uc.List(ctx, nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("unexpected list: %v, count=%d", err, len(list))
	}

	// 6. Update category name
	updated, err := uc.UpdateName(ctx, cat.ID, "Dining")
	if err != nil {
		t.Fatalf("failed to update category name: %v", err)
	}
	if updated.Name != "Dining" {
		t.Errorf("expected updated name 'Dining', got %q", updated.Name)
	}

	// 7. Update non-existent category
	_, err = uc.UpdateName(ctx, 999, "Non-existent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// 8. Delete referenced category
	store.referenced[cat.ID] = true
	err = uc.Delete(ctx, cat.ID)
	if !errors.Is(err, domain.ErrReferenced) {
		t.Fatalf("expected ErrReferenced, got %v", err)
	}

	// 9. Delete unreferenced
	store.referenced[cat.ID] = false
	if err := uc.Delete(ctx, cat.ID); err != nil {
		t.Fatalf("failed to delete unreferenced category: %v", err)
	}
}

func TestTransactionUsecase(t *testing.T) {
	catStore := newMockCategoryStore()
	txStore := newMockTransactionStore()
	uc := usecase.NewTransactionUsecase(txStore, catStore)
	ctx := context.Background()

	expenseCat := &domain.Category{ID: 1, Name: "Utilities", Type: domain.CategoryTypeExpense}
	incomeCat := &domain.Category{ID: 2, Name: "Salary", Type: domain.CategoryTypeIncome}
	_ = catStore.Create(ctx, expenseCat)
	_ = catStore.Create(ctx, incomeCat)

	// 1. Type mismatch validation (expense category with income tx)
	_, err := uc.Create(ctx, usecase.CreateTransactionInput{
		CategoryID: expenseCat.ID,
		Type:       domain.TransactionTypeIncome,
		AmountIDR:  100000,
		OccurredOn: "2026-10-15",
		Note:       "Electricity refund",
	})
	if !errors.Is(err, domain.ErrTypeMismatch) {
		t.Fatalf("expected ErrTypeMismatch, got %v", err)
	}

	// 2. Category not found
	_, err = uc.Create(ctx, usecase.CreateTransactionInput{
		CategoryID: 999,
		Type:       domain.TransactionTypeExpense,
		AmountIDR:  100000,
		OccurredOn: "2026-10-15",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-existing category, got %v", err)
	}

	// 3. Invariant validation error (amount <= 0)
	_, err = uc.Create(ctx, usecase.CreateTransactionInput{
		CategoryID: expenseCat.ID,
		Type:       domain.TransactionTypeExpense,
		AmountIDR:  0,
		OccurredOn: "2026-10-15",
	})
	if err == nil {
		t.Fatalf("expected validation error for amount 0")
	}

	// 4. Valid transaction creation
	tx, err := uc.Create(ctx, usecase.CreateTransactionInput{
		CategoryID: expenseCat.ID,
		Type:       domain.TransactionTypeExpense,
		AmountIDR:  250000,
		OccurredOn: "2026-10-15",
		Note:       "=SUM(A1:B2) malicious note",
	})
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}
	if tx.ID != 1 {
		t.Errorf("expected ID 1, got %d", tx.ID)
	}

	// 5. GetByID
	fetched, err := uc.GetByID(ctx, tx.ID)
	if err != nil || fetched.AmountIDR != 250000 {
		t.Fatalf("failed to get transaction: %v", err)
	}

	// 6. List transactions
	txList, count, err := uc.List(ctx, usecase.TransactionFilter{})
	if err != nil || count != 1 || len(txList) != 1 {
		t.Fatalf("failed to list transactions: %v", err)
	}

	// 7. Update transaction
	newAmount := domain.AmountIDR(300000)
	newNote := "Updated note"
	updated, err := uc.Update(ctx, tx.ID, usecase.UpdateTransactionInput{
		AmountIDR: &newAmount,
		Note:      &newNote,
	})
	if err != nil {
		t.Fatalf("failed to update transaction: %v", err)
	}
	if updated.AmountIDR != 300000 || updated.Note != "Updated note" {
		t.Errorf("unexpected updated tx: %+v", updated)
	}

	// 8. Update non-existent transaction
	_, err = uc.Update(ctx, 888, usecase.UpdateTransactionInput{
		AmountIDR: &newAmount,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound updating non-existent tx, got %v", err)
	}

	// 9. CSV Streaming with formula sanitization
	var buf bytes.Buffer
	err = uc.StreamCSV(ctx, usecase.TransactionFilter{}, &buf)
	if err != nil {
		t.Fatalf("failed to stream CSV: %v", err)
	}

	csvOutput := buf.String()
	if !strings.Contains(csvOutput, "Updated note") {
		t.Fatalf("expected CSV body to contain updated note, got:\n%s", csvOutput)
	}

	// 10. Delete transaction
	if err := uc.Delete(ctx, tx.ID); err != nil {
		t.Fatalf("failed to delete transaction: %v", err)
	}
}

func TestBudgetUsecase(t *testing.T) {
	catStore := newMockCategoryStore()
	budgetStore := newMockBudgetStore()
	uc := usecase.NewBudgetUsecase(budgetStore, catStore)
	ctx := context.Background()

	incomeCat := &domain.Category{ID: 1, Name: "Dividends", Type: domain.CategoryTypeIncome}
	expenseCat := &domain.Category{ID: 2, Name: "Groceries", Type: domain.CategoryTypeExpense}
	_ = catStore.Create(ctx, incomeCat)
	_ = catStore.Create(ctx, expenseCat)

	// 1. Budget on income category must fail with ErrTypeMismatch
	_, err := uc.Upsert(ctx, usecase.UpsertBudgetInput{
		CategoryID: incomeCat.ID,
		Month:      "2026-10",
		LimitIDR:   1000000,
	})
	if !errors.Is(err, domain.ErrTypeMismatch) {
		t.Fatalf("expected ErrTypeMismatch when setting budget on income category, got %v", err)
	}

	// 2. Budget on non-existent category must fail with ErrNotFound
	_, err = uc.Upsert(ctx, usecase.UpsertBudgetInput{
		CategoryID: 999,
		Month:      "2026-10",
		LimitIDR:   1000000,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-existing category, got %v", err)
	}

	// 3. Validation error on invalid month
	_, err = uc.Upsert(ctx, usecase.UpsertBudgetInput{
		CategoryID: expenseCat.ID,
		Month:      "invalid-month",
		LimitIDR:   1000000,
	})
	if err == nil {
		t.Fatalf("expected validation error on malformed month")
	}

	// 4. Budget on expense category succeeds
	b, err := uc.Upsert(ctx, usecase.UpsertBudgetInput{
		CategoryID: expenseCat.ID,
		Month:      "2026-10",
		LimitIDR:   2000000,
	})
	if err != nil {
		t.Fatalf("failed to upsert budget: %v", err)
	}
	if b.LimitIDR != 2000000 {
		t.Errorf("expected limit 2000000, got %d", b.LimitIDR)
	}

	// 5. GetByCategoryAndMonth
	fetched, err := uc.GetByCategoryAndMonth(ctx, expenseCat.ID, "2026-10")
	if err != nil || fetched.LimitIDR != 2000000 {
		t.Fatalf("failed to get budget: %v", err)
	}

	// 6. ListByMonth
	list, err := uc.ListByMonth(ctx, "2026-10")
	if err != nil {
		t.Fatalf("failed to list budgets by month: %v", err)
	}
	_ = list

	// 7. ListByMonth invalid format
	_, err = uc.ListByMonth(ctx, "bad-month")
	if err == nil {
		t.Fatalf("expected error on invalid month format")
	}
}

func TestReportUsecase(t *testing.T) {
	reportStore := &MockReportStore{
		monthlyTotals: map[string][3]domain.AmountIDR{
			"2026-10": {domain.AmountIDR(10000000), domain.AmountIDR(4000000), domain.AmountIDR(6000000)},
			"2026-09": {domain.AmountIDR(8000000), domain.AmountIDR(5000000), domain.AmountIDR(3000000)},
		},
		breakdown: map[string][]usecase.CategorySpendItem{
			"2026-10": {
				{CategoryID: 1, CategoryName: "Housing", AmountIDR: 2500000, Percentage: 62},
				{CategoryID: 2, CategoryName: "Groceries", AmountIDR: 1500000, Percentage: 37},
			},
		},
	}
	budgetStore := newMockBudgetStore()
	budgetStore.items = []domain.BudgetReportItem{
		{
			Budget:         domain.Budget{ID: 1, CategoryID: 1, Month: "2026-10", LimitIDR: 3000000},
			CategoryName:   "Housing",
			SpentIDR:       2500000,
			RemainingIDR:   500000,
			UtilizationPct: 83,
			Status:         domain.BudgetStatusWarning,
		},
	}

	uc := usecase.NewReportUsecase(reportStore, budgetStore)
	ctx := context.Background()

	// 1. Invalid month format
	_, err := uc.GetMonthlyReport(ctx, "invalid")
	if err == nil {
		t.Fatalf("expected error on invalid month in report")
	}

	// 2. Valid monthly report
	report, err := uc.GetMonthlyReport(ctx, "2026-10")
	if err != nil {
		t.Fatalf("failed to get monthly report: %v", err)
	}

	// Assert calculations:
	// Total Income = 10.000.000, Expense = 4.000.000, Net = 6.000.000
	// Savings rate = (6.000.000 / 10.000.000) * 100 = 60%
	if report.SavingsRatePct != 60 {
		t.Errorf("expected savings rate 60%%, got %d%%", report.SavingsRatePct)
	}
	// MoM Income Delta: (10.000.000 - 8.000.000) / 8.000.000 * 100 = +25%
	if report.MoMComparison.IncomeDeltaPct != 25 {
		t.Errorf("expected MoM income delta 25%%, got %d%%", report.MoMComparison.IncomeDeltaPct)
	}
	// MoM Expense Delta: (4.000.000 - 5.000.000) / 5.000.000 * 100 = -20%
	if report.MoMComparison.ExpenseDeltaPct != -20 {
		t.Errorf("expected MoM expense delta -20%%, got %d%%", report.MoMComparison.ExpenseDeltaPct)
	}
	if len(report.BudgetWatchlist) != 1 || report.BudgetWatchlist[0].Status != domain.BudgetStatusWarning {
		t.Errorf("unexpected budget watchlist: %+v", report.BudgetWatchlist)
	}
}
