// Package repository contains the Go migration of the JtSpringProject DAO
// layer (the com.jtspringproject.JtSpringProject.dao package).
//
// This file migrates categoryDao.java, the Hibernate-backed Data Access Object
// providing CRUD operations for the Category entity.
//
// In the Java world categoryDao was a Spring @Repository:
//
//	@Repository
//	public class categoryDao {
//	    private final SessionFactory sessionFactory;
//	    @Transactional public Category addCategory(String name) { ... }
//	    @Transactional public List<Category> getCategories() { ... }
//	    @Transactional public Boolean deleteCategory(int id) { ... }
//	    @Transactional public Category updateCategory(int id, String name) { ... }
//	    @Transactional public Category getCategory(int id) { ... }
//	}
//
// MIGRATION_NOTE: Several Spring/Hibernate concepts have no direct Go analogue
// and were replaced with idiomatic Go equivalents:
//
//   - Hibernate's SessionFactory / getCurrentSession() (a thread-bound session
//     tied to the transaction scope) is replaced by an injected *sql.DB. Each
//     method uses context.Context for cancellation/deadline propagation.
//   - Spring's declarative @Transactional is replaced by explicit transaction
//     management using sql.Tx where a write must be atomic (delete/update reads
//     then mutates). Simple single-statement operations run directly on the DB.
//   - The HQL query "from CATEGORY" becomes an explicit SQL SELECT. The concrete
//     table/column names must be verified against the actual schema — see the
//     MIGRATION_NOTE constants below.
//   - Java's constructor-based dependency injection is replaced by a
//     NewCategoryRepository constructor function.
//   - Nullable returns (Java returning null / Boolean) are replaced with the
//     idiomatic Go (T, error) convention, returning ErrCategoryNotFound where
//     the entity does not exist.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrCategoryNotFound is returned when a Category lookup finds no matching row.
//
// MIGRATION_NOTE: In Java, getCategory/updateCategory returned null and
// deleteCategory returned false when the entity did not exist. Go callers
// should use errors.Is(err, ErrCategoryNotFound) to distinguish this case.
var ErrCategoryNotFound = errors.New("category not found")

// MIGRATION_NOTE: The Hibernate entity Category (mapped to the "CATEGORY"
// entity in HQL) does not yet appear in the list of already-migrated symbols.
// The Category type below is a minimal local model reflecting the Java entity's
// id/name fields. Replace it with the canonical model.Category type once that
// entity is migrated, and update the field mappings accordingly.
type Category struct {
	id   int
	name string
}

// NewCategory constructs a Category with the given id and name.
func NewCategory(id int, name string) *Category {
	return &Category{id: id, name: name}
}

// GetID returns the category identifier.
func (c *Category) GetID() int { return c.id }

// SetID sets the category identifier.
func (c *Category) SetID(id int) { c.id = id }

// GetName returns the category name.
func (c *Category) GetName() string { return c.name }

// SetName sets the category name.
func (c *Category) SetName(name string) { c.name = name }

// MIGRATION_NOTE: These identifiers are best guesses derived from the Java
// entity name ("CATEGORY") and its fields. Verify them against the actual
// database schema and JPA @Table/@Column annotations on the Category entity.
const (
	categoryTable      = "category"
	categoryColID      = "id"
	categoryColName    = "name"
	categorySelectCols = categoryColID + ", " + categoryColName
)

// CategoryRepository provides CRUD access to Category records.
//
// It replaces the Java categoryDao @Repository. The injected *sql.DB stands in
// for Hibernate's SessionFactory.
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository constructs a CategoryRepository backed by the given
// database handle. It returns an error if db is nil.
//
// This replaces Spring's constructor-based dependency injection.
func NewCategoryRepository(db *sql.DB) (*CategoryRepository, error) {
	if db == nil {
		return nil, errors.New("repository: db must not be nil")
	}
	return &CategoryRepository{db: db}, nil
}

// AddCategory creates a new Category with the given name and returns it,
// populated with its generated identifier.
//
// This migrates categoryDao.addCategory, which built a Category, set its name,
// and called saveOrUpdate.
func (r *CategoryRepository) AddCategory(ctx context.Context, name string) (*Category, error) {
	const query = "INSERT INTO " + categoryTable + " (" + categoryColName + ") VALUES (?)"

	res, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("add category: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		// MIGRATION_NOTE: Not all drivers support LastInsertId (e.g. PostgreSQL
		// needs INSERT ... RETURNING id). If your driver does not, replace this
		// with a RETURNING clause and QueryRowContext.
		return nil, fmt.Errorf("add category: retrieve generated id: %w", err)
	}

	return NewCategory(int(id), name), nil
}

// GetCategories returns all Category records.
//
// This migrates categoryDao.getCategories (HQL "from CATEGORY").
func (r *CategoryRepository) GetCategories(ctx context.Context) ([]*Category, error) {
	const query = "SELECT " + categorySelectCols + " FROM " + categoryTable

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("get categories: scan row: %w", err)
		}
		categories = append(categories, NewCategory(id, name))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get categories: iterate rows: %w", err)
	}

	return categories, nil
}

// DeleteCategory removes the Category with the given id.
//
// It returns ErrCategoryNotFound if no row matched. This migrates
// categoryDao.deleteCategory, which returned true/false depending on whether
// the entity existed.
func (r *CategoryRepository) DeleteCategory(ctx context.Context, id int) error {
	const query = "DELETE FROM " + categoryTable + " WHERE " + categoryColID + " = ?"

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete category %d: %w", id, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete category %d: rows affected: %w", id, err)
	}
	if affected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// UpdateCategory updates the name of the Category with the given id and returns
// the updated record.
//
// It returns ErrCategoryNotFound if no row matched. This migrates
// categoryDao.updateCategory, which fetched the entity, returned null if
// missing, then updated its name.
//
// MIGRATION_NOTE: The Java code performed a read (get) followed by an update
// within a single @Transactional method. This is preserved here using an
// explicit sql.Tx so the existence check and the update are atomic.
func (r *CategoryRepository) UpdateCategory(ctx context.Context, id int, name string) (*Category, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("update category %d: begin tx: %w", id, err)
	}
	defer func() { _ = tx.Rollback() }() // no-op after a successful Commit

	const selectQuery = "SELECT " + categoryColID + " FROM " + categoryTable + " WHERE " + categoryColID + " = ?"
	var existingID int
	if err := tx.QueryRowContext(ctx, selectQuery, id).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("update category %d: lookup: %w", id, err)
	}

	const updateQuery = "UPDATE " + categoryTable + " SET " + categoryColName + " = ? WHERE " + categoryColID + " = ?"
	if _, err := tx.ExecContext(ctx, updateQuery, name, id); err != nil {
		return nil, fmt.Errorf("update category %d: %w", id, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("update category %d: commit: %w", id, err)
	}

	return NewCategory(id, name), nil
}

// GetCategory returns the Category with the given id.
//
// It returns ErrCategoryNotFound if no row matched. This migrates
// categoryDao.getCategory, which returned null when the entity did not exist.
func (r *CategoryRepository) GetCategory(ctx context.Context, id int) (*Category, error) {
	const query = "SELECT " + categorySelectCols + " FROM " + categoryTable + " WHERE " + categoryColID + " = ?"

	var (
		foundID int
		name    string
	)
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&foundID, &name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category %d: %w", id, err)
	}

	return NewCategory(foundID, name), nil
}
