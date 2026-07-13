// Package repository contains the Go migration of the JtSpringProject DAO
// layer (the com.jtspringproject.JtSpringProject.dao package).
//
// This file migrates cartDao.java, the Hibernate-backed Data Access Object
// providing CRUD operations for the Cart entity.
//
// In the Java world cartDao was a Spring @Repository:
//
//	@Repository
//	public class cartDao {
//	    private final SessionFactory sessionFactory;
//	    @Transactional public Cart addCart(Cart cart) { ... }
//	    @Transactional public List<Cart> getCarts() { ... }
//	    @Transactional public void updateCart(Cart cart) { ... }
//	    @Transactional public void deleteCart(Cart cart) { ... }
//	}
//
// MIGRATION_NOTE: Several Spring/Hibernate concepts have no direct Go analogue
// and were replaced with idiomatic Go equivalents:
//
//   - @Repository (component scanning + exception translation): replaced with a
//     plain struct plus a NewCartRepository constructor for explicit DI.
//   - Constructor-injected SessionFactory: replaced with an injected *sql.DB.
//     Go has no Hibernate session; database/sql manages a connection pool.
//   - @Transactional (declarative AOP transaction management): replaced with
//     explicit transactions started inside each method via db.BeginTx. There is
//     no proxy/interceptor magic in Go — transaction boundaries are explicit.
//   - Hibernate getCurrentSession()/save/update/delete/HQL: replaced with raw
//     SQL executed through database/sql. The HQL "from CART" becomes a SELECT.
//   - context.Context is threaded through every method for cancellation and
//     deadline propagation, per Go I/O conventions.
//
// MIGRATION_NOTE: The Cart model type does not yet appear in the list of
// already-migrated symbols. This file assumes a migrated model.Cart type in
// internal/jtspringproject/jtspringproject/model exposing GetID/SetID accessors
// (mirroring the User model). The scan/insert column mapping below is a
// best-effort placeholder and REQUIRES MANUAL REVIEW once the Cart entity is
// migrated, because the original JPA @Entity field/column mapping is not
// available in this source file.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// CartRepository provides CRUD operations for the Cart entity.
//
// It is the Go migration of the Spring @Repository cartDao. Instead of a
// Hibernate SessionFactory it depends on a *sql.DB connection pool, and instead
// of declarative @Transactional it manages transactions explicitly.
type CartRepository struct {
	db *sql.DB
}

// NewCartRepository constructs a CartRepository backed by the given database
// connection pool.
//
// This replaces Spring's constructor-based dependency injection of the
// SessionFactory. It returns an error if db is nil so wiring mistakes surface
// immediately at construction time.
func NewCartRepository(db *sql.DB) (*CartRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("repository: db must not be nil")
	}
	return &CartRepository{db: db}, nil
}

// AddCart persists a new Cart and returns it.
//
// It is the migration of the @Transactional addCart method. The insert is run
// inside an explicit transaction; on success the generated identifier is set on
// the returned Cart.
//
// MIGRATION_NOTE: The column list and generated-key handling are placeholders
// and require review against the migrated Cart entity mapping.
func (r *CartRepository) AddCart(ctx context.Context, cart *model.Cart) (*model.Cart, error) {
	if cart == nil {
		return nil, fmt.Errorf("repository: cart must not be nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repository: begin transaction: %w", err)
	}

	result, err := tx.ExecContext(ctx, "INSERT INTO cart DEFAULT VALUES")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("repository: insert cart: %w (rollback failed: %v)", err, rbErr)
		}
		return nil, fmt.Errorf("repository: insert cart: %w", err)
	}

	if id, idErr := result.LastInsertId(); idErr == nil {
		cart.SetID(int(id))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: commit transaction: %w", err)
	}
	return cart, nil
}

// GetCarts returns all Cart records.
//
// It is the migration of the @Transactional getCarts method, whose HQL
// "from CART" becomes a SELECT over the cart table.
//
// MIGRATION_NOTE: The selected columns and row scan mapping are placeholders
// and require review against the migrated Cart entity mapping.
func (r *CartRepository) GetCarts(ctx context.Context) ([]*model.Cart, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("repository: begin transaction: %w", err)
	}

	rows, err := tx.QueryContext(ctx, "SELECT id FROM cart")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("repository: query carts: %w (rollback failed: %v)", err, rbErr)
		}
		return nil, fmt.Errorf("repository: query carts: %w", err)
	}
	defer rows.Close()

	var carts []*model.Cart
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("repository: scan cart row: %w", err)
		}
		cart := &model.Cart{}
		cart.SetID(id)
		carts = append(carts, cart)
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("repository: iterate cart rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: commit transaction: %w", err)
	}
	return carts, nil
}

// UpdateCart persists changes to an existing Cart.
//
// It is the migration of the @Transactional updateCart method.
//
// MIGRATION_NOTE: The update column list is a placeholder and requires review
// against the migrated Cart entity mapping.
func (r *CartRepository) UpdateCart(ctx context.Context, cart *model.Cart) error {
	if cart == nil {
		return fmt.Errorf("repository: cart must not be nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repository: begin transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "UPDATE cart SET id = id WHERE id = ?", cart.GetID()); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("repository: update cart: %w (rollback failed: %v)", err, rbErr)
		}
		return fmt.Errorf("repository: update cart: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: commit transaction: %w", err)
	}
	return nil
}

// DeleteCart removes an existing Cart.
//
// It is the migration of the @Transactional deleteCart method.
func (r *CartRepository) DeleteCart(ctx context.Context, cart *model.Cart) error {
	if cart == nil {
		return fmt.Errorf("repository: cart must not be nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repository: begin transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM cart WHERE id = ?", cart.GetID()); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("repository: delete cart: %w (rollback failed: %v)", err, rbErr)
		}
		return fmt.Errorf("repository: delete cart: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: commit transaction: %w", err)
	}
	return nil
}
