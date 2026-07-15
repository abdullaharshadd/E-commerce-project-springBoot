package service

import (
	"context"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// CategoryRepository is the subset of repository operations required by the
// CategoryService. Depending on an interface (rather than the concrete
// *repository.CategoryDAO) keeps the service testable and decoupled from the
// persistence implementation.
//
// MIGRATION_NOTE: The original Spring @Service injected the concrete categoryDao
// via constructor-based @Autowired. In idiomatic Go we accept an interface in
// the constructor, which both documents the required behaviour and allows tests
// to substitute a fake without a DI container.
type CategoryRepository interface {
	AddCategory(ctx context.Context, name string) (model.Category, error)
	GetCategories(ctx context.Context) ([]model.Category, error)
	DeleteCategory(ctx context.Context, id int) (bool, error)
	UpdateCategory(ctx context.Context, id int, name string) (model.Category, error)
	GetCategory(ctx context.Context, id int) (model.Category, error)
}

// CategoryService provides business-layer CRUD operations for Category
// entities by delegating to a CategoryRepository.
//
// MIGRATION_NOTE: This replaces the Spring @Service class of the same name. The
// service layer/DAO delegation pattern is preserved, but each method now takes
// a context.Context (propagated to the persistence layer) and returns an
// explicit error instead of relying on unchecked exceptions.
type CategoryService struct {
	repo CategoryRepository
}

// NewCategoryService constructs a CategoryService backed by the given
// CategoryRepository. It returns an error if the repository dependency is nil.
func NewCategoryService(repo CategoryRepository) (*CategoryService, error) {
	if repo == nil {
		return nil, fmt.Errorf("service: category repository must not be nil")
	}
	return &CategoryService{repo: repo}, nil
}

// AddCategory creates a new category with the given name and returns the
// persisted entity.
func (s *CategoryService) AddCategory(ctx context.Context, name string) (model.Category, error) {
	category, err := s.repo.AddCategory(ctx, name)
	if err != nil {
		return model.Category{}, fmt.Errorf("service: add category: %w", err)
	}
	return category, nil
}

// GetCategories returns all categories.
func (s *CategoryService) GetCategories(ctx context.Context) ([]model.Category, error) {
	categories, err := s.repo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: get categories: %w", err)
	}
	return categories, nil
}

// DeleteCategory removes the category identified by id and reports whether the
// deletion succeeded.
func (s *CategoryService) DeleteCategory(ctx context.Context, id int) (bool, error) {
	ok, err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return false, fmt.Errorf("service: delete category %d: %w", id, err)
	}
	return ok, nil
}

// UpdateCategory updates the name of the category identified by id and returns
// the updated entity.
func (s *CategoryService) UpdateCategory(ctx context.Context, id int, name string) (model.Category, error) {
	category, err := s.repo.UpdateCategory(ctx, id, name)
	if err != nil {
		return model.Category{}, fmt.Errorf("service: update category %d: %w", id, err)
	}
	return category, nil
}

// GetCategory returns the category identified by id.
func (s *CategoryService) GetCategory(ctx context.Context, id int) (model.Category, error) {
	category, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		return model.Category{}, fmt.Errorf("service: get category %d: %w", id, err)
	}
	return category, nil
}
