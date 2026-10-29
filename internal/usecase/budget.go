package usecase

import (
	"context"
	"strings"

	"flowledger/internal/domain"
)

// BudgetStore defines the persistence contract required for budgets.
type BudgetStore interface {
	Upsert(ctx context.Context, b *domain.Budget) error
	GetByCategoryAndMonth(ctx context.Context, categoryID int64, month string) (*domain.Budget, error)
	ListWithUtilization(ctx context.Context, month string) ([]domain.BudgetReportItem, error)
}

// BudgetUsecase coordinates budget management and validation.
type BudgetUsecase struct {
	budgetStore BudgetStore
	catStore    CategoryStore
}

// NewBudgetUsecase creates a new BudgetUsecase.
func NewBudgetUsecase(budgetStore BudgetStore, catStore CategoryStore) *BudgetUsecase {
	return &BudgetUsecase{
		budgetStore: budgetStore,
		catStore:    catStore,
	}
}

// UpsertBudgetInput holds the payload for creating or updating a budget.
type UpsertBudgetInput struct {
	CategoryID int64            `json:"category_id"`
	Month      string           `json:"month"`
	LimitIDR   domain.AmountIDR `json:"limit_idr"`
}

// Upsert creates or updates a budget for an expense category.
func (uc *BudgetUsecase) Upsert(ctx context.Context, in UpsertBudgetInput) (*domain.Budget, error) {
	budget := &domain.Budget{
		CategoryID: in.CategoryID,
		Month:      strings.TrimSpace(in.Month),
		LimitIDR:   in.LimitIDR,
	}

	if err := budget.Validate(); err != nil {
		return nil, err
	}

	cat, err := uc.catStore.GetByID(ctx, budget.CategoryID)
	if err != nil {
		return nil, err
	}

	if err := budget.ValidateCategoryType(cat.Type); err != nil {
		return nil, err
	}

	if err := uc.budgetStore.Upsert(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

// ListByMonth retrieves all budgets with utilization data for a specific month.
func (uc *BudgetUsecase) ListByMonth(ctx context.Context, month string) ([]domain.BudgetReportItem, error) {
	month = strings.TrimSpace(month)
	// Validate month format
	dummy := domain.Budget{Month: month, CategoryID: 1, LimitIDR: 1000}
	if err := dummy.Validate(); err != nil {
		return nil, err
	}

	return uc.budgetStore.ListWithUtilization(ctx, month)
}

// GetByCategoryAndMonth returns the budget for the specified category and month.
func (uc *BudgetUsecase) GetByCategoryAndMonth(ctx context.Context, categoryID int64, month string) (*domain.Budget, error) {
	return uc.budgetStore.GetByCategoryAndMonth(ctx, categoryID, strings.TrimSpace(month))
}

