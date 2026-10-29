package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"flowledger/internal/domain"
	"golang.org/x/sync/errgroup"
)

// CategorySpendItem represents category spend in a given month.
type CategorySpendItem struct {
	CategoryID   int64            `json:"category_id"`
	CategoryName string           `json:"category_name"`
	AmountIDR    domain.AmountIDR `json:"amount_idr"`
	Percentage   int              `json:"percentage"`
}

// MoMDelta represents Month-over-Month percentage changes.
type MoMDelta struct {
	PrevIncomeIDR   domain.AmountIDR `json:"prev_income_idr"`
	PrevExpenseIDR  domain.AmountIDR `json:"prev_expense_idr"`
	IncomeDeltaPct  int              `json:"income_delta_pct"`
	ExpenseDeltaPct int              `json:"expense_delta_pct"`
}

// MonthlyReport is the unified response containing financial metrics for a month.
type MonthlyReport struct {
	Month             string                    `json:"month"`
	TotalIncomeIDR    domain.AmountIDR          `json:"total_income_idr"`
	TotalExpenseIDR   domain.AmountIDR          `json:"total_expense_idr"`
	NetCashFlowIDR    domain.AmountIDR          `json:"net_cash_flow_idr"`
	SavingsRatePct    int                       `json:"savings_rate_pct"`
	MoMComparison     MoMDelta                  `json:"mom_comparison"`
	CategoryBreakdown []CategorySpendItem       `json:"category_breakdown"`
	BudgetWatchlist   []domain.BudgetReportItem `json:"budget_watchlist"`
}

// ReportStore defines analytical query contracts.
type ReportStore interface {
	GetMonthlyTotals(ctx context.Context, month string) (totalIncome, totalExpense, net domain.AmountIDR, err error)
	GetCategoryBreakdown(ctx context.Context, month string, totalExpense domain.AmountIDR) ([]CategorySpendItem, error)
}

// ReportUsecase orchestrates concurrent report generation using errgroup.
type ReportUsecase struct {
	reportStore ReportStore
	budgetStore BudgetStore
}

// NewReportUsecase creates a new ReportUsecase.
func NewReportUsecase(reportStore ReportStore, budgetStore BudgetStore) *ReportUsecase {
	return &ReportUsecase{
		reportStore: reportStore,
		budgetStore: budgetStore,
	}
}

// GetMonthlyReport executes concurrent parallel queries via errgroup to build a comprehensive monthly report.
func (uc *ReportUsecase) GetMonthlyReport(ctx context.Context, month string) (*MonthlyReport, error) {
	month = strings.TrimSpace(month)
	parsedDate, err := time.Parse("2006-01", month)
	if err != nil {
		var ve domain.ValidationErrors
		ve.Add("month", "month must be in YYYY-MM format")
		return nil, ve
	}

	prevMonth := parsedDate.AddDate(0, -1, 0).Format("2006-01")

	var (
		currIncome, currExpense, currNet domain.AmountIDR
		prevIncome, prevExpense          domain.AmountIDR
		breakdown                        []CategorySpendItem
		budgetWatchlist                  []domain.BudgetReportItem
	)

	g, gCtx := errgroup.WithContext(ctx)

	// Query 1: Current month totals
	g.Go(func() error {
		var err error
		currIncome, currExpense, currNet, err = uc.reportStore.GetMonthlyTotals(gCtx, month)
		if err != nil {
			return fmt.Errorf("failed fetching current monthly totals: %w", err)
		}
		return nil
	})

	// Query 2: Previous month totals for MoM comparison
	g.Go(func() error {
		var err error
		prevIncome, prevExpense, _, err = uc.reportStore.GetMonthlyTotals(gCtx, prevMonth)
		if err != nil {
			return fmt.Errorf("failed fetching previous monthly totals: %w", err)
		}
		return nil
	})

	// Query 3: Budget watchlist for the month
	g.Go(func() error {
		budgets, err := uc.budgetStore.ListWithUtilization(gCtx, month)
		if err != nil {
			return fmt.Errorf("failed fetching budget utilization: %w", err)
		}
		// Watchlist includes any budgets (or filter towards WARNING/EXCEEDED first)
		budgetWatchlist = budgets
		return nil
	})

	// Wait for the totals query to finish before running breakdown, or run breakdown in parallel
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Query 4: Category breakdown with the actual total expense
	breakdown, err = uc.reportStore.GetCategoryBreakdown(ctx, month, currExpense)
	if err != nil {
		return nil, fmt.Errorf("failed fetching category breakdown: %w", err)
	}

	// Compute Savings Rate: ((Income - Expense) * 100) / Income
	var savingsRatePct int
	if currIncome.Int64() > 0 {
		savingsRatePct = int(((currIncome.Int64() - currExpense.Int64()) * 100) / currIncome.Int64())
	}

	// Compute MoM Delta
	mom := MoMDelta{
		PrevIncomeIDR:  prevIncome,
		PrevExpenseIDR: prevExpense,
	}
	if prevIncome.Int64() > 0 {
		mom.IncomeDeltaPct = int(((currIncome.Int64() - prevIncome.Int64()) * 100) / prevIncome.Int64())
	} else if currIncome.Int64() > 0 {
		mom.IncomeDeltaPct = 100
	}

	if prevExpense.Int64() > 0 {
		mom.ExpenseDeltaPct = int(((currExpense.Int64() - prevExpense.Int64()) * 100) / prevExpense.Int64())
	} else if currExpense.Int64() > 0 {
		mom.ExpenseDeltaPct = 100
	}

	return &MonthlyReport{
		Month:             month,
		TotalIncomeIDR:    currIncome,
		TotalExpenseIDR:   currExpense,
		NetCashFlowIDR:    currNet,
		SavingsRatePct:    savingsRatePct,
		MoMComparison:     mom,
		CategoryBreakdown: breakdown,
		BudgetWatchlist:   budgetWatchlist,
	}, nil
}

