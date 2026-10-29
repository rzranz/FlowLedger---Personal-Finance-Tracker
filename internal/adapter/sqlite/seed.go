package sqlite

import (
	"context"
	"fmt"
	"time"

	"flowledger/internal/domain"
)

// SeedRealisticIndonesianData inserts realistic IDR categories, transactions, and budgets
// designed for portfolio demonstration.
func SeedRealisticIndonesianData(ctx context.Context, db *DB) error {
	catRepo := NewCategoryRepository(db)
	txRepo := NewTransactionRepository(db)
	budgetRepo := NewBudgetRepository(db)

	// Check if already seeded
	existing, err := catRepo.List(ctx, nil)
	if err == nil && len(existing) > 0 {
		return nil // already seeded
	}

	// 1. Categories
	incomeCats := []string{
		"Gaji & Kompensasi",
		"Freelance & Proyek Sampingan",
		"Investasi & Dividen",
	}
	expenseCats := []string{
		"Makanan & Minuman",
		"Tempat Tinggal & Kos",
		"Transportasi",
		"Utilitas & Internet",
		"Hiburan & Rekreasi",
		"Kesehatan & Obat",
	}

	catMap := make(map[string]int64)

	for _, name := range incomeCats {
		cat := &domain.Category{Name: name, Type: domain.CategoryTypeIncome}
		if err := catRepo.Create(ctx, cat); err != nil {
			return fmt.Errorf("seeding income cat %s: %w", name, err)
		}
		catMap[name] = cat.ID
	}

	for _, name := range expenseCats {
		cat := &domain.Category{Name: name, Type: domain.CategoryTypeExpense}
		if err := catRepo.Create(ctx, cat); err != nil {
			return fmt.Errorf("seeding expense cat %s: %w", name, err)
		}
		catMap[name] = cat.ID
	}

	now := time.Now()
	currMonth := now.Format("2006-01")
	prevMonth := now.AddDate(0, -1, 0).Format("2006-01")

	// 2. Sample Transactions
	sampleTxs := []struct {
		catName string
		txType  domain.TransactionType
		amount  int64
		date    string
		note    string
	}{
		// Previous Month Income & Expenses
		{"Gaji & Kompensasi", domain.TransactionTypeIncome, 12000000, fmt.Sprintf("%s-01", prevMonth), "Gaji bulanan reguler"},
		{"Freelance & Proyek Sampingan", domain.TransactionTypeIncome, 3500000, fmt.Sprintf("%s-14", prevMonth), "Proyek backend API Go"},
		{"Tempat Tinggal & Kos", domain.TransactionTypeExpense, 2500000, fmt.Sprintf("%s-02", prevMonth), "Sewa kos bulanan"},
		{"Makanan & Minuman", domain.TransactionTypeExpense, 2800000, fmt.Sprintf("%s-10", prevMonth), "Makan siang & belanja mingguan"},
		{"Transportasi", domain.TransactionTypeExpense, 600000, fmt.Sprintf("%s-15", prevMonth), "Bensin & e-toll"},
		{"Utilitas & Internet", domain.TransactionTypeExpense, 450000, fmt.Sprintf("%s-18", prevMonth), "Tagihan Indihome & token listrik"},

		// Current Month Income
		{"Gaji & Kompensasi", domain.TransactionTypeIncome, 12500000, fmt.Sprintf("%s-01", currMonth), "Gaji bulanan + bonus performa"},
		{"Freelance & Proyek Sampingan", domain.TransactionTypeIncome, 2000000, fmt.Sprintf("%s-12", currMonth), "Konsultasi software architecture"},

		// Current Month Expenses
		{"Tempat Tinggal & Kos", domain.TransactionTypeExpense, 2500000, fmt.Sprintf("%s-02", currMonth), "Sewa kos bulanan"},
		{"Makanan & Minuman", domain.TransactionTypeExpense, 1750000, fmt.Sprintf("%s-05", currMonth), "Belanja bulanan supermarket"},
		{"Transportasi", domain.TransactionTypeExpense, 650000, fmt.Sprintf("%s-08", currMonth), "Servis motor & transport kantor"},
		{"Utilitas & Internet", domain.TransactionTypeExpense, 480000, fmt.Sprintf("%s-10", currMonth), "Listrik PLN & internet fiber"},
		{"Hiburan & Rekreasi", domain.TransactionTypeExpense, 350000, fmt.Sprintf("%s-15", currMonth), "Nonton bioskop & kafe kopi"},
		{"Kesehatan & Obat", domain.TransactionTypeExpense, 180000, fmt.Sprintf("%s-20", currMonth), "Vitamin & suplemen"},
	}

	for _, s := range sampleTxs {
		catID, ok := catMap[s.catName]
		if !ok {
			continue
		}
		tx := &domain.Transaction{
			CategoryID: catID,
			Type:       s.txType,
			AmountIDR:  domain.AmountIDR(s.amount),
			OccurredOn: s.date,
			Note:       s.note,
		}
		if err := txRepo.Create(ctx, tx); err != nil {
			return fmt.Errorf("seeding transaction (%s): %w", s.note, err)
		}
	}

	// 3. Budgets demonstrating SAFE, WARNING, and EXCEEDED
	sampleBudgets := []struct {
		catName string
		limit   int64
	}{
		// Limit 2.500.000, spent 1.750.000 -> 70% SAFE (<80%)
		{"Makanan & Minuman", 2500000},
		// Limit 750.000, spent 650.000 -> 86% WARNING (80% - 99%)
		{"Transportasi", 750000},
		// Limit 300.000, spent 350.000 -> 116% EXCEEDED (>=100%)
		{"Hiburan & Rekreasi", 300000},
		// Limit 3.000.000, spent 2.500.000 -> 83% WARNING
		{"Tempat Tinggal & Kos", 3000000},
	}

	for _, b := range sampleBudgets {
		catID, ok := catMap[b.catName]
		if !ok {
			continue
		}
		budget := &domain.Budget{
			CategoryID: catID,
			Month:      currMonth,
			LimitIDR:   domain.AmountIDR(b.limit),
		}
		if err := budgetRepo.Upsert(ctx, budget); err != nil {
			return fmt.Errorf("seeding budget (%s): %w", b.catName, err)
		}
	}

	return nil
}

