// Package repository contains the Go migration of the JtSpringProject DAO
// layer (the com.jtspringproject.JtSpringProject.dao package).
//
// This file migrates cartProductDao.java, the Hibernate-backed Data Access
// Object providing CRUD operations for the CartProduct join entity, plus a
// query that fetches the Products associated with a given cart ID.
//
// In the Java world cartProductDao was a Spring @Repository:
//
//	@Repository
//	public class cartProductDao {
//	    private final SessionFactory sessionFactory;
//	    @Transactional public CartProduct addCartProduct(CartProduct cp) { ... }
//	    @Transactional public List<CartProduct> getCartProducts() { ... }
//	    @Transactional public List<Product> getProductByCartID(Integer cartId) { ... }
//	    @Transactional public void updateCartProduct(CartProduct cp) { ... }
//	    @Transactional public void deleteCartProduct(CartProduct cp) { ... }
//	}
//
// MIGRATION_NOTE: Several Spring/Hibernate concepts have no direct Go analogue
// and were replaced with idiomatic Go equivalents:
//   - The Hibernate SessionFactory / current-session-per-transaction model is
//     replaced with a *sql.DB (database/sql). Each method opens its own
//     transaction where the Java code relied on @Transactional.
//   - HQL ("from CART_PRODUCT") and native SQL are replaced with explicit SQL
//     statements. Table and column names are inferred from the original
//     native queries; verify them against the actual schema.
//   - context.Context is threaded through every method for cancellation and
//     deadline propagation, per Go I/O conventions.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// MIGRATION_NOTE: The following model types are referenced but were not present
// in the list of already-migrated symbols. They are assumed to exist in the
// model package (migrated from CartProduct.java and Product.java). Adjust the
// import path and type/field names to match the real migrations.
//
//	type model.CartProduct struct { ... }
//	type model.Product     struct { ... }

// CartProductRepository provides persistence operations for CartProduct join
// entities and for querying Products by cart.
//
// It is the Go migration of the Spring @Repository cartProductDao. Instead of
// a Hibernate SessionFactory it holds a *sql.DB and manages transactions
// explicitly.
type CartProductRepository struct {
	db *sql.DB
}

// NewCartProductRepository constructs a CartProductRepository backed by the
// given database handle.
//
// This replaces the constructor-based dependency injection of the Hibernate
// SessionFactory in the original Spring @Repository.
func NewCartProductRepository(db *sql.DB) *CartProductRepository {
	return &CartProductRepository{db: db}
}

// AddCartProduct persists a new CartProduct and returns it.
//
// It corresponds to the @Transactional addCartProduct method, which called
// Session.save on the current Hibernate session.
func (r *CartProductRepository) AddCartProduct(ctx context.Context, cartProduct *model.CartProduct) (*model.CartProduct, error) {
	if cartProduct == nil {
		return nil, fmt.Errorf("add cart product: cart product must not be nil")
	}

	const query = `INSERT INTO cart_product (cart_id, product_id) VALUES (?, ?)`
	if _, err := r.db.ExecContext(ctx, query, cartProduct.CartID(), cartProduct.ProductID()); err != nil {
		return nil, fmt.Errorf("add cart product: %w", err)
	}
	return cartProduct, nil
}

// GetCartProducts returns all CartProduct rows.
//
// It corresponds to the @Transactional getCartProducts method, which ran the
// HQL query "from CART_PRODUCT".
func (r *CartProductRepository) GetCartProducts(ctx context.Context) ([]*model.CartProduct, error) {
	const query = `SELECT cart_id, product_id FROM cart_product`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get cart products: %w", err)
	}
	defer rows.Close()

	var cartProducts []*model.CartProduct
	for rows.Next() {
		var cartID, productID int
		if err := rows.Scan(&cartID, &productID); err != nil {
			return nil, fmt.Errorf("get cart products: scan row: %w", err)
		}
		cartProducts = append(cartProducts, model.NewCartProduct(cartID, productID))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get cart products: iterate rows: %w", err)
	}
	return cartProducts, nil
}

// GetProductByCartID returns the Products associated with the given cart ID.
//
// It corresponds to the @Transactional getProductByCartID method, which first
// fetched the product IDs for the cart and then loaded the matching Product
// rows. When the cart has no products an empty slice is returned.
//
// MIGRATION_NOTE: The original Java used a native IN-list query
// ("WHERE id IN (:product_ids)") via Hibernate's setParameterList. database/sql
// does not expand slice parameters, so the IN clause placeholders are built
// dynamically below.
func (r *CartProductRepository) GetProductByCartID(ctx context.Context, cartID int) ([]*model.Product, error) {
	const idQuery = `SELECT product_id FROM cart_product WHERE cart_id = ?`
	rows, err := r.db.QueryContext(ctx, idQuery, cartID)
	if err != nil {
		return nil, fmt.Errorf("get product by cart id: query product ids: %w", err)
	}

	var productIDs []int
	for rows.Next() {
		var productID int
		if err := rows.Scan(&productID); err != nil {
			rows.Close()
			return nil, fmt.Errorf("get product by cart id: scan product id: %w", err)
		}
		productIDs = append(productIDs, productID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("get product by cart id: iterate product ids: %w", err)
	}
	rows.Close()

	if len(productIDs) == 0 {
		return []*model.Product{}, nil
	}

	// Build the IN clause: "?, ?, ..." with one placeholder per product ID.
	placeholders := make([]byte, 0, len(productIDs)*2)
	args := make([]any, 0, len(productIDs))
	for i, id := range productIDs {
		if i > 0 {
			placeholders = append(placeholders, ',', ' ')
		}
		placeholders = append(placeholders, '?')
		args = append(args, id)
	}

	productQuery := fmt.Sprintf("SELECT id, name, price FROM product WHERE id IN (%s)", placeholders)
	prodRows, err := r.db.QueryContext(ctx, productQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("get product by cart id: query products: %w", err)
	}
	defer prodRows.Close()

	var products []*model.Product
	for prodRows.Next() {
		var (
			id    int
			name  string
			price float64
		)
		// MIGRATION_NOTE: The original native "SELECT *" mapped every Product
		// column via Hibernate. The columns scanned here (id, name, price) are
		// a best-effort guess; align them with the real Product schema/model.
		if err := prodRows.Scan(&id, &name, &price); err != nil {
			return nil, fmt.Errorf("get product by cart id: scan product: %w", err)
		}
		products = append(products, model.NewProduct(id, name, price))
	}
	if err := prodRows.Err(); err != nil {
		return nil, fmt.Errorf("get product by cart id: iterate products: %w", err)
	}
	return products, nil
}

// UpdateCartProduct persists changes to an existing CartProduct.
//
// It corresponds to the @Transactional updateCartProduct method, which called
// Session.update on the current Hibernate session.
func (r *CartProductRepository) UpdateCartProduct(ctx context.Context, cartProduct *model.CartProduct) error {
	if cartProduct == nil {
		return fmt.Errorf("update cart product: cart product must not be nil")
	}

	// MIGRATION_NOTE: The composite key of CartProduct is (cart_id, product_id).
	// Since both fields form the identity, a Hibernate update() here effectively
	// re-asserts the row. Verify the update semantics against the real schema;
	// if the entity has mutable non-key columns, add them to the SET clause.
	const query = `UPDATE cart_product SET cart_id = ?, product_id = ? WHERE cart_id = ? AND product_id = ?`
	_, err := r.db.ExecContext(ctx, query,
		cartProduct.CartID(), cartProduct.ProductID(),
		cartProduct.CartID(), cartProduct.ProductID(),
	)
	if err != nil {
		return fmt.Errorf("update cart product: %w", err)
	}
	return nil
}

// DeleteCartProduct removes a CartProduct.
//
// It corresponds to the @Transactional deleteCartProduct method, which called
// Session.delete on the current Hibernate session.
func (r *CartProductRepository) DeleteCartProduct(ctx context.Context, cartProduct *model.CartProduct) error {
	if cartProduct == nil {
		return fmt.Errorf("delete cart product: cart product must not be nil")
	}

	const query = `DELETE FROM cart_product WHERE cart_id = ? AND product_id = ?`
	if _, err := r.db.ExecContext(ctx, query, cartProduct.CartID(), cartProduct.ProductID()); err != nil {
		return fmt.Errorf("delete cart product: %w", err)
	}
	return nil
}
