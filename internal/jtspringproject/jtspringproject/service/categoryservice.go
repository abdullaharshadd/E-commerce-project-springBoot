// Package service contains the Go migration of the JtSpringProject service
// layer (the com.jtspringproject.JtSpringProject.services package).
//
// This file migrates categoryService.java, a thin service-layer wrapper around
// the categoryDao that exposes CRUD operations for product Category entities.
// In the Java world it was a Spring @Service that delegated every call directly
// to the DAO with no additional business logic:
//
//	@Service
//	public class categoryService {
//	    private final categoryDao categoryDao;
//	    @Autowired public categoryService(categoryDao categoryDao) { ... }
//	    public Category addCategory(String name) { return categoryDao.addCategory(name); }
//	    public List<Category> getCategories() { return categoryDao.getCategories(); }
//	    public Boolean deleteCategory(int id) { return categoryDao.deleteCategory(id); }
//	    public Category updateCategory(int id, String name) { return categoryDao.updateCategory(id, name); }
//	    public Category getCategory(int id) { return categoryDao.getCategory(id); }
//	}
//
// MIGRATION_NOTE: The DAO (categoryDao.java) had not yet been migrated at the
// time this file was produced. The CategoryRepository interface below describes
// the contract this service depends on. The concrete repository implementation
// must live in the repository package (categorydao.go) and satisfy this
// interface. Because the repository package does not yet export a Category
// type, this service defines the CategoryRepository dependency in terms of an
// interface it declares locally; the repository's concrete type is expected to
// implement it. Adjust the import and Category type once the DAO is migrated.
package service

import (
	"context"
	"fmt"
)

// Category is the domain model for a product category.
//
// MIGRATION_NOTE: In the Java code this was com.jtspringproject.JtSpringProject
// .models.Category. The models package had not been migrated at the time this
// file was produced, so a minimal representation is declared here. When the
// models package is migrated, replace this type with an import of the canonical
// Category and remove this declaration.
type Category struct {
	// ID is the unique identifier of the category.
	ID int
	// Name is the display name of the category.
	Name string
}

// CategoryRepository is the data-access contract required by CategoryService.
//
// It mirrors the operations that categoryDao provided in the Java code. The
// concrete implementation lives in the repository package. All methods take a
// context.Context so that request-scoped values, cancellation and deadlines
// propagate through to the database layer (replacing Spring's implicit
// @Transactional session management).
type CategoryRepository interface {
	// AddCategory persists a new category with the given name.
	AddCategory(ctx context.Context, name string) (Category, error)
	// GetCategories returns all categories.
	GetCategories(ctx context.Context) ([]Category, error)
	// DeleteCategory removes the category with the given id. It reports
	// whether a category was deleted.
	DeleteCategory(ctx context.Context, id int) (bool, error)
	// UpdateCategory updates the name of the category with the given id and
	// returns the updated category.
	UpdateCategory(ctx context.Context, id int, name string) (Category, error)
	// GetCategory returns the category with the given id.
	GetCategory(ctx context.Context, id int) (Category, error)
}

// CategoryService provides business operations for managing product
// categories. It is a thin pass-through over a CategoryRepository, matching the
// original Spring @Service which contained no additional business logic.
type CategoryService struct {
	repo CategoryRepository
}

// NewCategoryService constructs a CategoryService backed by the given
// repository. This replaces the Spring constructor injection (@Autowired) with
// explicit dependency injection. It returns an error if repo is nil so that
// misconfiguration is caught at wiring time rather than at first use.
func NewCategoryService(repo CategoryRepository) (*CategoryService, error) {
	if repo == nil {
		return nil, fmt.Errorf("service: CategoryRepository must not be nil")
	}
	return &CategoryService{repo: repo}, nil
}

// AddCategory creates a new category with the given name.
func (s *CategoryService) AddCategory(ctx context.Context, name string) (Category, error) {
	category, err := s.repo.AddCategory(ctx, name)
	if err != nil {
		return Category{}, fmt.Errorf("service: add category %q: %w", name, err)
	}
	return category, nil
}

// GetCategories returns all categories.
func (s *CategoryService) GetCategories(ctx context.Context) ([]Category, error) {
	categories, err := s.repo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: get categories: %w", err)
	}
	return categories, nil
}

// DeleteCategory removes the category with the given id. It reports whether a
// category was deleted.
func (s *CategoryService) DeleteCategory(ctx context.Context, id int) (bool, error) {
	deleted, err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return false, fmt.Errorf("service: delete category %d: %w", id, err)
	}
	return deleted, nil
}

// UpdateCategory updates the name of the category with the given id and returns
// the updated category.
func (s *CategoryService) UpdateCategory(ctx context.Context, id int, name string) (Category, error) {
	category, err := s.repo.UpdateCategory(ctx, id, name)
	if err != nil {
		return Category{}, fmt.Errorf("service: update category %d: %w", id, err)
	}
	return category, nil
}

// GetCategory returns the category with the given id.
func (s *CategoryService) GetCategory(ctx context.Context, id int) (Category, error) {
	category, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		return Category{}, fmt.Errorf("service: get category %d: %w", id, err)
	}
	return category, nil
}
