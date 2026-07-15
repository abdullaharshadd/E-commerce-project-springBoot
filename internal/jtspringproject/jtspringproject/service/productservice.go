package service

import (
	"context"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// ProductRepository is the subset of repository operations required by the
// ProductService. Depending on an interface (rather than the concrete
// *repository.ProductDAO) keeps the service testable and decoupled from the
// persistence implementation.
//
// MIGRATION_NOTE: The original Spring @Service injected the concrete productDao
// via constructor-based @Autowired. In idiomatic Go we accept an interface in
// the constructor, which both documents the required behaviour and allows tests
// to substitute a fake without a DI container.
type ProductRepository interface {
	GetProducts(ctx context.Context) ([]model.Product, error)
	AddProduct(ctx context.Context, product model.Product) (model.Product, error)
	GetProduct(ctx context.Context, id int) (model.Product, error)
	UpdateProduct(ctx context.Context, product model.Product) (model.Product, error)
	DeleteProduct(ctx context.Context, id int) (bool, error)
}

// ProductService provides product-related business operations, delegating
// persistence to a ProductRepository.
//
// MIGRATION_NOTE: The original was a Spring @Service acting as a pass-through
// layer between controllers and the productDao. The delegation semantics are
// preserved; the only added logic is explicit error propagation, since Go does
// not have unchecked exceptions.
type ProductService struct {
	repo ProductRepository
}

// NewProductService constructs a ProductService backed by the given
// ProductRepository.
func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// GetProducts returns all products.
func (s *ProductService) GetProducts(ctx context.Context) ([]model.Product, error) {
	products, err := s.repo.GetProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}
	return products, nil
}

// AddProduct persists a new product and returns the stored representation.
func (s *ProductService) AddProduct(ctx context.Context, product model.Product) (model.Product, error) {
	saved, err := s.repo.AddProduct(ctx, product)
	if err != nil {
		return model.Product{}, fmt.Errorf("add product: %w", err)
	}
	return saved, nil
}

// GetProduct returns the product identified by id.
func (s *ProductService) GetProduct(ctx context.Context, id int) (model.Product, error) {
	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return model.Product{}, fmt.Errorf("get product %d: %w", id, err)
	}
	return product, nil
}

// UpdateProduct updates the product identified by id with the provided fields
// and returns the stored representation.
//
// MIGRATION_NOTE: The original set product.setId(id) before delegating to the
// DAO. We reproduce that here by overwriting the ID on the incoming value so
// the repository operates on the correct entity.
func (s *ProductService) UpdateProduct(ctx context.Context, id int, product model.Product) (model.Product, error) {
	product.ID = id
	updated, err := s.repo.UpdateProduct(ctx, product)
	if err != nil {
		return model.Product{}, fmt.Errorf("update product %d: %w", id, err)
	}
	return updated, nil
}

// DeleteProduct removes the product identified by id and reports whether the
// deletion succeeded.
func (s *ProductService) DeleteProduct(ctx context.Context, id int) (bool, error) {
	ok, err := s.repo.DeleteProduct(ctx, id)
	if err != nil {
		return false, fmt.Errorf("delete product %d: %w", id, err)
	}
	return ok, nil
}
