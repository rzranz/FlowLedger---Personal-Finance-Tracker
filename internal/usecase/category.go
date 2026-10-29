package usecase

import (
	"context"
	"strings"

	"flowledger/internal/domain"
)

// CategoryStore defines the storage contract required by CategoryUsecase.
type CategoryStore interface {
	Create(ctx context.Context, cat *domain.Category) error
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
	List(ctx context.Context, catType *domain.CategoryType) ([]domain.Category, error)
	Update(ctx context.Context, cat *domain.Category) error
	Delete(ctx context.Context, id int64) error
	HasReferences(ctx context.Context, id int64) (bool, error)
}

// CategoryUsecase orchestrates category-related business logic.
type CategoryUsecase struct {
	store CategoryStore
}

// NewCategoryUsecase creates a new CategoryUsecase.
func NewCategoryUsecase(store CategoryStore) *CategoryUsecase {
	return &CategoryUsecase{store: store}
}

// Create validates and creates a new category.
func (uc *CategoryUsecase) Create(ctx context.Context, name string, catType domain.CategoryType) (*domain.Category, error) {
	cat := &domain.Category{
		Name: strings.TrimSpace(name),
		Type: catType,
	}

	if err := cat.Validate(); err != nil {
		return nil, err
	}

	if err := uc.store.Create(ctx, cat); err != nil {
		return nil, err
	}

	return cat, nil
}

// GetByID retrieves a category by its ID.
func (uc *CategoryUsecase) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	return uc.store.GetByID(ctx, id)
}

// List returns all categories, optionally filtered by type.
func (uc *CategoryUsecase) List(ctx context.Context, catType *domain.CategoryType) ([]domain.Category, error) {
	return uc.store.List(ctx, catType)
}

// UpdateName updates the name of an existing category.
func (uc *CategoryUsecase) UpdateName(ctx context.Context, id int64, newName string) (*domain.Category, error) {
	cat, err := uc.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	cat.Name = strings.TrimSpace(newName)
	if err := cat.Validate(); err != nil {
		return nil, err
	}

	if err := uc.store.Update(ctx, cat); err != nil {
		return nil, err
	}

	return cat, nil
}

// Delete removes a category if it is not referenced by any transactions or budgets.
func (uc *CategoryUsecase) Delete(ctx context.Context, id int64) error {
	hasRef, err := uc.store.HasReferences(ctx, id)
	if err != nil {
		return err
	}
	if hasRef {
		return domain.ErrReferenced
	}

	return uc.store.Delete(ctx, id)
}

