package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// CartDAO provides CRUD operations for the Cart entity against the database.
//
// MIGRATION_NOTE: The original was a Spring @Repository backed by Hibernate's
// SessionFactory. In Go there is no ORM session bound to a thread; instead we
// use the SQLTransactor abstraction (see hibernateconfiguration.go) which owns
// the *sql.DB and exposes WithinTransaction to replace Spring's declarative
// @Transactional AOP. Each method opens a transaction explicitly and takes a
// context.Context as its first parameter, propagating cancellation/deadlines
// through the call stack.
type CartDAO struct {
	transactor *jtspringproject.SQLTransactor
}

// NewCartDAO constructs a CartDAO backed by the given SQLTransactor.
//
// MIGRATION_NOTE: This replaces Spring's constructor-based dependency injection
// of the SessionFactory. Wiring is performed explicitly in Run.
func NewCartDAO(transactor *jtspringproject.SQLTransactor) *CartDAO {
	return &CartDAO{transactor: transactor}
}

// AddCart persists the given cart and returns it.
//
// MIGRATION_NOTE: Mirrors the original session.save(cart). The Cart model
// (see model/cart.go) carries the product/user identifiers that make up the
// row. Adjust column names to match the actual schema during manual review.
func (d *CartDAO) AddCart(ctx context.Context, cart *model.Cart) (*model.Cart, error) {
	if cart == nil {
		return nil, fmt.Errorf("add cart: cart must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		const query = `INSERT INTO cart (user_id, product_id) VALUES (?, ?)`
		if _, err := tx.ExecContext(ctx, query, cart.UserID, cart.ProductID); err != nil {
			return fmt.Errorf("insert cart: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("add cart: %w", err)
	}
	return cart, nil
}

// GetCarts returns all carts.
//
// MIGRATION_NOTE: Replaces the HQL query "from CART". HQL operated on the
// mapped entity; here we issue explicit SQL against the cart table and scan
// rows into model.Cart values.
func (d *CartDAO) GetCarts(ctx context.Context) ([]*model.Cart, error) {
	var carts []*model.Cart

	err := d.transactor.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		const query = `SELECT id, user_id, product_id FROM cart`
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return fmt.Errorf("query carts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id        int
				userID    int
				productID int
			)
			if err := rows.Scan(&id, &userID, &productID); err != nil {
				return fmt.Errorf("scan cart row: %w", err)
			}
			cart := model.NewCart(userID, productID)
			cart.ID = id
			carts = append(carts, cart)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate cart rows: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get carts: %w", err)
	}
	return carts, nil
}

// UpdateCart updates the given cart.
//
// MIGRATION_NOTE: Mirrors the original session.update(cart).
func (d *CartDAO) UpdateCart(ctx context.Context, cart *model.Cart) error {
	if cart == nil {
		return fmt.Errorf("update cart: cart must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		const query = `UPDATE cart SET user_id = ?, product_id = ? WHERE id = ?`
		if _, err := tx.ExecContext(ctx, query, cart.UserID, cart.ProductID, cart.ID); err != nil {
			return fmt.Errorf("update cart: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update cart: %w", err)
	}
	return nil
}

// DeleteCart removes the given cart.
//
// MIGRATION_NOTE: Mirrors the original session.delete(cart).
func (d *CartDAO) DeleteCart(ctx context.Context, cart *model.Cart) error {
	if cart == nil {
		return fmt.Errorf("delete cart: cart must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		const query = `DELETE FROM cart WHERE id = ?`
		if _, err := tx.ExecContext(ctx, query, cart.ID); err != nil {
			return fmt.Errorf("delete cart: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete cart: %w", err)
	}
	return nil
}
