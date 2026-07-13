// Package service contains the Go migration of the JtSpringProject service
// layer (the com.jtspringproject.JtSpringProject.services package).
//
// This file migrates productService.java, a thin service-layer wrapper around
// the productDao that exposes CRUD operations for Product entities. In the Java
// world it was a Spring @Service that delegated every call directly to the DAO
// with no additional business logic (aside from setting the ID on update):
//
//	@Service
//	public class productService {
//	    private final productDao productDao;
//	    @Autowired public productService(productDao productDao) { ... }
//	    public List<Product> getProducts() { return productDao.getProducts(); }
//	    public Product addProduct(Product product) { return productDao.addProduct(product); }
//	    public Product getProduct(int id) { return productDao.getProduct(id); }
//	    public Product updateProduct(int id, Product product) { product.setId(id); return productDao.updateProduct(product); }
//	    public boolean deleteProduct(int id) { return productDao.deleteProduct(id); }
//	}
//
// MIGRATION_NOTE: The Spring @Service stereotype and @Autowired constructor
// injection are replaced by an explicit NewProductService constructor that
// accepts a ProductRepository interface. Bean wiring is done manually at
// application startup instead of via component scanning.
//
// MIGRATION_NOTE: All I/O-bound operations take a context.Context as their
// first parameter for cancellation/deadline propagation, and return an
// explicit error rather than relying on unchecked exceptions.
package service

import (
	"context"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/repository"
)

// Product is the entity managed by the product service. It mirrors the Java
// Product model. The concrete definition is expected to live in the repository
// (data access) layer, alongside productDao.
//
// MIGRATION_NOTE: The Product type is referenced here from the repository
// package. If the migrated Product model lives elsewhere, adjust the alias/
// import accordingly.
type Product = repository.Product

// ProductRepository abstracts the data-access operations the ProductService
// depends on. It corresponds to the Java productDao and lets the service be
// unit-tested with a mock implementation.
//
// MIGRATION_NOTE: Method names are inferred from the Java DAO. Verify they
// match the actual migrated productdao.go signatures during review.
type ProductRepository interface {
	// GetProducts returns all products.
	GetProducts(ctx context.Context) ([]Product, error)
	// AddProduct persists a new product and returns the stored entity.
	AddProduct(ctx context.Context, product Product) (Product, error)
	// GetProduct returns the product with the given id, or ErrProductNotFound.
	GetProduct(ctx context.Context, id int) (Product, error)
	// UpdateProduct persists changes to an existing product and returns it.
	UpdateProduct(ctx context.Context, product Product) (Product, error)
	// DeleteProduct removes the product with the given id.
	DeleteProduct(ctx context.Context, id int) error
}

// ProductService provides business-logic operations for Product entities by
// delegating to a ProductRepository. It is the Go equivalent of the Spring
// productService @Service bean.
type ProductService struct {
	repo ProductRepository
}

// NewProductService constructs a ProductService backed by the given
// ProductRepository. It replaces the Spring @Autowired constructor.
func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// GetProducts returns all products.
func (s *ProductService) GetProducts(ctx context.Context) ([]Product, error) {
	products, err := s.repo.GetProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}
	return products, nil
}

// AddProduct persists a new product and returns the stored entity.
func (s *ProductService) AddProduct(ctx context.Context, product Product) (Product, error) {
	saved, err := s.repo.AddProduct(ctx, product)
	if err != nil {
		return Product{}, fmt.Errorf("add product: %w", err)
	}
	return saved, nil
}

// GetProduct returns the product with the given id. It wraps
// repository.ErrProductNotFound when no matching product exists.
func (s *ProductService) GetProduct(ctx context.Context, id int) (Product, error) {
	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return Product{}, fmt.Errorf("get product %d: %w", id, err)
	}
	return product, nil
}

// UpdateProduct sets the id on the supplied product and persists the changes,
// returning the updated entity. This mirrors the Java behaviour of calling
// product.setId(id) before delegating to the DAO.
func (s *ProductService) UpdateProduct(ctx context.Context, id int, product Product) (Product, error) {
	product.SetID(id)
	updated, err := s.repo.UpdateProduct(ctx, product)
	if err != nil {
		return Product{}, fmt.Errorf("update product %d: %w", id, err)
	}
	return updated, nil
}

// DeleteProduct removes the product with the given id.
//
// MIGRATION_NOTE: The Java method returned a boolean indicating success. In Go
// we surface failure as an error instead; a nil error means the delete
// succeeded. Callers that specifically need to distinguish "not found" can use
// errors.Is(err, repository.ErrProductNotFound).
func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		return fmt.Errorf("delete product %d: %w", id, err)
	}
	return nil
}
