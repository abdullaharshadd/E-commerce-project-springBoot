package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// CartProductDAO provides CRUD operations for the CartProduct entity and a
// helper to fetch the Products associated with a given cart.
//
// MIGRATION_NOTE: The original was a Spring @Repository backed by Hibernate's
// SessionFactory. In Go there is no ORM session bound to a thread; instead we
// use the SQLTransactor abstraction (see hibernateconfiguration.go) which owns
// the *sql.DB and exposes WithinTransaction to replace Spring's declarative
// @Transactional AOP. Each method opens a transaction explicitly and takes a
// context.Context as its first parameter, propagating cancellation/deadlines
// through the call stack.
type CartProductDAO struct {
	transactor jtspringproject.SQLTransactor
}

// NewCartProductDAO constructs a CartProductDAO backed by the given transactor.
//
// MIGRATION_NOTE: Replaces Spring's constructor-based dependency injection of
// the Hibernate SessionFactory.
func NewCartProductDAO(transactor jtspringproject.SQLTransactor) *CartProductDAO {
	return &CartProductDAO{transactor: transactor}
}

// AddCartProduct persists a new CartProduct and returns it.
//
// MIGRATION_NOTE: Mirrors Hibernate session.save on the CART_PRODUCT table.
// The composite key (cart_id, product_id) is written explicitly.
func (d *CartProductDAO) AddCartProduct(ctx context.Context, cartProduct *model.CartProduct) (*model.CartProduct, error) {
	if cartProduct == nil {
		return nil, fmt.Errorf("add cart product: cartProduct must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `INSERT INTO cart_product (cart_id, product_id, quantity) VALUES (?, ?, ?)`
		if _, execErr := tx.ExecContext(ctx, query,
			cartProduct.ID.CartID,
			cartProduct.ID.ProductID,
			cartProduct.Quantity,
		); execErr != nil {
			return fmt.Errorf("insert cart product: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("add cart product: %w", err)
	}
	return cartProduct, nil
}

// GetCartProducts returns all CartProduct rows.
//
// MIGRATION_NOTE: Replaces the HQL query "from CART_PRODUCT" with an explicit
// SELECT against the cart_product table.
func (d *CartProductDAO) GetCartProducts(ctx context.Context) ([]*model.CartProduct, error) {
	var cartProducts []*model.CartProduct

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `SELECT cart_id, product_id, quantity FROM cart_product`
		rows, queryErr := tx.QueryContext(ctx, query)
		if queryErr != nil {
			return fmt.Errorf("query cart products: %w", queryErr)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				cartID    int
				productID int
				quantity  int
			)
			if scanErr := rows.Scan(&cartID, &productID, &quantity); scanErr != nil {
				return fmt.Errorf("scan cart product: %w", scanErr)
			}
			id := model.NewCartProductID(cartID, productID)
			cp := &model.CartProduct{ID: id, Quantity: quantity}
			cartProducts = append(cartProducts, cp)
		}
		if rowsErr := rows.Err(); rowsErr != nil {
			return fmt.Errorf("iterate cart products: %w", rowsErr)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get cart products: %w", err)
	}
	return cartProducts, nil
}

// GetProductByCartID returns the Products associated with the given cart ID.
//
// MIGRATION_NOTE: The original used two native queries: first to collect the
// product_id list for a cart, then a second query with an IN clause. The IN
// clause is built dynamically here with the correct number of placeholders,
// since database/sql does not expand slice parameters. When no product IDs are
// found, an empty slice is returned without issuing the second query.
func (d *CartProductDAO) GetProductByCartID(ctx context.Context, cartID int) ([]*model.Product, error) {
	var products []*model.Product

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const idQuery = `SELECT product_id FROM cart_product WHERE cart_id = ?`
		rows, queryErr := tx.QueryContext(ctx, idQuery, cartID)
		if queryErr != nil {
			return fmt.Errorf("query product ids for cart %d: %w", cartID, queryErr)
		}

		var productIDs []int
		for rows.Next() {
			var productID int
			if scanErr := rows.Scan(&productID); scanErr != nil {
				rows.Close()
				return fmt.Errorf("scan product id: %w", scanErr)
			}
			productIDs = append(productIDs, productID)
		}
		if rowsErr := rows.Err(); rowsErr != nil {
			rows.Close()
			return fmt.Errorf("iterate product ids: %w", rowsErr)
		}
		rows.Close()

		if len(productIDs) == 0 {
			return nil
		}

		placeholders := make([]string, len(productIDs))
		args := make([]interface{}, len(productIDs))
		for i, id := range productIDs {
			placeholders[i] = "?"
			args[i] = id
		}
		productQuery := fmt.Sprintf(
			`SELECT id, name, price, weight, description, category_id, image FROM product WHERE id IN (%s)`,
			strings.Join(placeholders, ", "),
		)

		productRows, prodErr := tx.QueryContext(ctx, productQuery, args...)
		if prodErr != nil {
			return fmt.Errorf("query products by ids: %w", prodErr)
		}
		defer productRows.Close()

		for productRows.Next() {
			var (
				id          int
				name        string
				price       int
				weight      int
				description string
				categoryID  int
				image       string
			)
			if scanErr := productRows.Scan(&id, &name, &price, &weight, &description, &categoryID, &image); scanErr != nil {
				return fmt.Errorf("scan product: %w", scanErr)
			}
			// MIGRATION_NOTE: Only the product's category id is available from
			// this row; the associated Category is left nil here. A join or a
			// follow-up lookup would be required to hydrate it fully.
			product := model.NewProduct(id, name, image, price, weight, description, nil)
			products = append(products, product)
		}
		if rowsErr := productRows.Err(); rowsErr != nil {
			return fmt.Errorf("iterate products: %w", rowsErr)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get products by cart id: %w", err)
	}
	return products, nil
}

// UpdateCartProduct updates an existing CartProduct identified by its composite key.
//
// MIGRATION_NOTE: Mirrors Hibernate session.update. The composite key
// (cart_id, product_id) is immutable, so only the mutable quantity column is
// updated.
func (d *CartProductDAO) UpdateCartProduct(ctx context.Context, cartProduct *model.CartProduct) error {
	if cartProduct == nil {
		return fmt.Errorf("update cart product: cartProduct must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `UPDATE cart_product SET quantity = ? WHERE cart_id = ? AND product_id = ?`
		if _, execErr := tx.ExecContext(ctx, query,
			cartProduct.Quantity,
			cartProduct.ID.CartID,
			cartProduct.ID.ProductID,
		); execErr != nil {
			return fmt.Errorf("update cart product: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update cart product: %w", err)
	}
	return nil
}

// DeleteCartProduct removes the CartProduct identified by its composite key.
//
// MIGRATION_NOTE: Mirrors Hibernate session.delete.
func (d *CartProductDAO) DeleteCartProduct(ctx context.Context, cartProduct *model.CartProduct) error {
	if cartProduct == nil {
		return fmt.Errorf("delete cart product: cartProduct must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `DELETE FROM cart_product WHERE cart_id = ? AND product_id = ?`
		if _, execErr := tx.ExecContext(ctx, query,
			cartProduct.ID.CartID,
			cartProduct.ID.ProductID,
		); execErr != nil {
			return fmt.Errorf("delete cart product: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete cart product: %w", err)
	}
	return nil
}
