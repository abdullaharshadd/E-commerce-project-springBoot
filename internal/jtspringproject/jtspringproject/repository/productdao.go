// Package repository contains the Go migration of the JtSpringProject DAO
// layer (the com.jtspringproject.JtSpringProject.dao package).
//
// This file migrates productDao.java, the Hibernate-backed Data Access Object
// providing CRUD operations for the Product entity.
//
// In the Java world productDao was a Spring @Repository:
//
//	@Repository
//	public class productDao {
//	    private final SessionFactory sessionFactory;
//	    @Transactional public List<Product> getProducts() { ... }
//	    @Transactional public Product addProduct(Product product) { ... }
//	    @Transactional public Product getProduct(int id) { ... }
//	    @Transactional public Product updateProduct(Product product) { ... }
//	    @Transactional public Boolean deleteProduct(int id) { ... }
//	}
//
// MIGRATION_NOTE: Several Spring/Hibernate concepts have no direct Go analogue
// and were replaced with idiomatic Go equivalents:
//
//   - @Repository / constructor injection of a Hibernate SessionFactory becomes
//     an exported struct (ProductRepository) holding a *sql.DB, created via a
//     NewProductRepository constructor.
//   - @Transactional declarative transaction management has no Go equivalent.
//     Each method uses the standard database/sql API. Callers that need
//     multi-statement atomicity should wrap operations in a *sql.Tx. The
//     context.Context is threaded through every I/O operation so cancellation
//     and timeouts propagate correctly.
//   - Hibernate's HQL "from PRODUCT" and get/save/update/delete auto-mapping
//     become explicit SQL statements. The table/column names below are best
//     guesses derived from the Product model and MUST be verified against the
//     actual schema (see MIGRATION_NOTE on the SQL constants).
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrProductNotFound is returned when a Product lookup by ID yields no rows.
//
// MIGRATION_NOTE: In Java, getProduct returned null and deleteProduct returned
// false when the entity was absent. In idiomatic Go we surface a sentinel error
// from lookups so callers can distinguish "not found" from other failures using
// errors.Is.
var ErrProductNotFound = errors.New("product not found")

// Product is the persistence representation of a product row.
//
// MIGRATION_NOTE: The Java Product model class was not part of this migration
// unit. The fields below are inferred from typical usage and the SQL mapping;
// verify names/types against the actual models.Product definition and adjust
// the struct tags / scan targets accordingly.
type Product struct {
	ID       int
	Name     string
	Category string
	Price    int
	Weight   int
	Quantity int
	Image    string
	Description string
}

// SQL statements backing the repository.
//
// MIGRATION_NOTE: These statements were reconstructed from Hibernate's implicit
// mapping ("from PRODUCT", save/get/update/delete on Product). The table name
// ("product") and column list are assumptions and REQUIRE manual verification
// against the real database schema and the models.Product annotations.
const (
	selectProductsSQL = `SELECT id, name, category, price, weight, quantity, image, description FROM product`
	selectProductSQL  = `SELECT id, name, category, price, weight, quantity, image, description FROM product WHERE id = ?`
	insertProductSQL  = `INSERT INTO product (name, category, price, weight, quantity, image, description) VALUES (?, ?, ?, ?, ?, ?, ?)`
	updateProductSQL  = `UPDATE product SET name = ?, category = ?, price = ?, weight = ?, quantity = ?, image = ?, description = ? WHERE id = ?`
	deleteProductSQL  = `DELETE FROM product WHERE id = ?`
)

// ProductRepository provides CRUD operations for Product entities.
//
// It is the Go equivalent of the Spring @Repository productDao. Rather than a
// Hibernate SessionFactory it depends on a *sql.DB, injected via
// NewProductRepository.
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository constructs a ProductRepository backed by the given
// *sql.DB. It returns an error if db is nil so misconfiguration is caught at
// startup rather than on first query.
func NewProductRepository(db *sql.DB) (*ProductRepository, error) {
	if db == nil {
		return nil, errors.New("repository: db must not be nil")
	}
	return &ProductRepository{db: db}, nil
}

// GetProducts returns all products.
//
// It is the Go equivalent of the Java getProducts() HQL query "from PRODUCT".
func (r *ProductRepository) GetProducts(ctx context.Context) ([]Product, error) {
	rows, err := r.db.QueryContext(ctx, selectProductsSQL)
	if err != nil {
		return nil, fmt.Errorf("repository: query products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Category,
			&p.Price,
			&p.Weight,
			&p.Quantity,
			&p.Image,
			&p.Description,
		); err != nil {
			return nil, fmt.Errorf("repository: scan product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate products: %w", err)
	}
	return products, nil
}

// AddProduct inserts the given product and returns it with its generated ID
// populated.
//
// It is the Go equivalent of the Java addProduct(Product), which relied on
// Hibernate's save() to assign the identifier.
func (r *ProductRepository) AddProduct(ctx context.Context, product Product) (Product, error) {
	res, err := r.db.ExecContext(ctx, insertProductSQL,
		product.Name,
		product.Category,
		product.Price,
		product.Weight,
		product.Quantity,
		product.Image,
		product.Description,
	)
	if err != nil {
		return Product{}, fmt.Errorf("repository: insert product: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		// MIGRATION_NOTE: Not all drivers support LastInsertId (e.g. PostgreSQL).
		// If you switch to such a driver, use INSERT ... RETURNING id with
		// QueryRowContext instead.
		return Product{}, fmt.Errorf("repository: read generated product id: %w", err)
	}
	product.ID = int(id)
	return product, nil
}

// GetProduct returns the product with the given ID. It returns
// ErrProductNotFound if no matching row exists.
//
// It is the Go equivalent of the Java getProduct(int), which returned null when
// the entity was absent.
func (r *ProductRepository) GetProduct(ctx context.Context, id int) (Product, error) {
	var p Product
	err := r.db.QueryRowContext(ctx, selectProductSQL, id).Scan(
		&p.ID,
		&p.Name,
		&p.Category,
		&p.Price,
		&p.Weight,
		&p.Quantity,
		&p.Image,
		&p.Description,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrProductNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("repository: get product %d: %w", id, err)
	}
	return p, nil
}

// UpdateProduct persists changes to the given product and returns it.
//
// It returns ErrProductNotFound if no row with the product's ID exists, which
// is stricter than the Java updateProduct(Product) that delegated to
// Hibernate's update() without reporting missing rows.
func (r *ProductRepository) UpdateProduct(ctx context.Context, product Product) (Product, error) {
	res, err := r.db.ExecContext(ctx, updateProductSQL,
		product.Name,
		product.Category,
		product.Price,
		product.Weight,
		product.Quantity,
		product.Image,
		product.Description,
		product.ID,
	)
	if err != nil {
		return Product{}, fmt.Errorf("repository: update product %d: %w", product.ID, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return Product{}, fmt.Errorf("repository: rows affected updating product %d: %w", product.ID, err)
	}
	if affected == 0 {
		return Product{}, ErrProductNotFound
	}
	return product, nil
}

// DeleteProduct removes the product with the given ID. It reports whether a
// product was actually deleted, mirroring the Java deleteProduct(int) which
// returned true only when the entity existed.
func (r *ProductRepository) DeleteProduct(ctx context.Context, id int) (bool, error) {
	res, err := r.db.ExecContext(ctx, deleteProductSQL, id)
	if err != nil {
		return false, fmt.Errorf("repository: delete product %d: %w", id, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("repository: rows affected deleting product %d: %w", id, err)
	}
	return affected > 0, nil
}
