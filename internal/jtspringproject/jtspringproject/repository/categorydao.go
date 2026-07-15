package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// CategoryDAO provides CRUD operations for the Category entity against the
// database.
//
// MIGRATION_NOTE: The original was a Spring @Repository backed by Hibernate's
// SessionFactory. In Go there is no ORM session bound to a thread; instead we
// use the SQLTransactor abstraction (see hibernateconfiguration.go) which owns
// the *sql.DB and exposes WithinTransaction to replace Spring's declarative
// @Transactional AOP. Each method opens a transaction explicitly and takes a
// context.Context as its first parameter, propagating cancellation/deadlines
// down the call stack.
type CategoryDAO struct {
	tx jtspringproject.SQLTransactor
}

// NewCategoryDAO constructs a CategoryDAO backed by the given SQLTransactor.
//
// MIGRATION_NOTE: Replaces Spring constructor-based dependency injection of the
// Hibernate SessionFactory.
func NewCategoryDAO(tx jtspringproject.SQLTransactor) *CategoryDAO {
	return &CategoryDAO{tx: tx}
}

// AddCategory inserts a new category with the given name and returns the
// persisted Category (including its generated ID).
//
// MIGRATION_NOTE: The Java saveOrUpdate on a transient entity performs an
// INSERT. Here we execute an explicit INSERT within a transaction and read back
// the generated identifier.
func (d *CategoryDAO) AddCategory(ctx context.Context, name string) (*model.Category, error) {
	var category *model.Category
	err := d.tx.WithinTransaction(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "INSERT INTO category (name) VALUES (?)", name)
		if err != nil {
			return fmt.Errorf("inserting category: %w", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("retrieving inserted category id: %w", err)
		}
		category = model.NewCategory(int(id), name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return category, nil
}

// GetCategories returns all categories.
//
// MIGRATION_NOTE: Replaces the HQL "from CATEGORY" list query with an explicit
// SELECT. Returns an empty (non-nil) slice when there are no rows.
func (d *CategoryDAO) GetCategories(ctx context.Context) ([]*model.Category, error) {
	categories := make([]*model.Category, 0)
	err := d.tx.WithinTransaction(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, "SELECT id, name FROM category")
		if err != nil {
			return fmt.Errorf("querying categories: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id   int
				name string
			)
			if err := rows.Scan(&id, &name); err != nil {
				return fmt.Errorf("scanning category row: %w", err)
			}
			categories = append(categories, model.NewCategory(id, name))
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterating category rows: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// DeleteCategory deletes the category with the given ID. It returns true when a
// category was found and deleted, and false when no category with that ID
// exists.
//
// MIGRATION_NOTE: The Java code loaded the entity first to decide the boolean
// return value. Here we rely on the affected-row count from the DELETE, which
// is equivalent and avoids an extra round trip.
func (d *CategoryDAO) DeleteCategory(ctx context.Context, id int) (bool, error) {
	var deleted bool
	err := d.tx.WithinTransaction(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "DELETE FROM category WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("deleting category %d: %w", id, err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("checking affected rows for category %d: %w", id, err)
		}
		deleted = affected > 0
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}

// UpdateCategory updates the name of the category with the given ID and returns
// the updated Category. It returns (nil, nil) when no category with that ID
// exists.
//
// MIGRATION_NOTE: The Java method returned null when the entity was not found.
// We preserve that behaviour with a (nil, nil) return, detecting absence via
// the affected-row count of the UPDATE.
func (d *CategoryDAO) UpdateCategory(ctx context.Context, id int, name string) (*model.Category, error) {
	var category *model.Category
	err := d.tx.WithinTransaction(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE category SET name = ? WHERE id = ?", name, id)
		if err != nil {
			return fmt.Errorf("updating category %d: %w", id, err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("checking affected rows for category %d: %w", id, err)
		}
		if affected == 0 {
			return nil
		}
		category = model.NewCategory(id, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return category, nil
}

// GetCategory returns the category with the given ID. It returns (nil, nil)
// when no category with that ID exists.
//
// MIGRATION_NOTE: The Java session.get returned null for a missing entity. We
// translate sql.ErrNoRows into a (nil, nil) return to preserve that contract.
func (d *CategoryDAO) GetCategory(ctx context.Context, id int) (*model.Category, error) {
	var category *model.Category
	err := d.tx.WithinTransaction(ctx, func(tx *sql.Tx) error {
		var name string
		row := tx.QueryRowContext(ctx, "SELECT id, name FROM category WHERE id = ?", id)
		if err := row.Scan(&id, &name); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("querying category %d: %w", id, err)
		}
		category = model.NewCategory(id, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return category, nil
}
