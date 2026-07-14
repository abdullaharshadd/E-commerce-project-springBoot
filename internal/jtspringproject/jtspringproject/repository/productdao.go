package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// ProductDAO provides CRUD operations for the Product entity against the
// database.
//
// MIGRATION_NOTE: The original was a Spring @Repository backed by Hibernate's
// SessionFactory. In Go there is no ORM session bound to a thread; instead we
// use the SQLTransactor abstraction (see hibernateconfiguration.go) which owns
// the *sql.DB and exposes WithinTransaction to replace Spring's declarative
// @Transactional AOP. Each method opens a transaction explicitly and takes a
// context.Context as its first parameter, propagating cancellation/deadlines
// through the call stack.
type ProductDAO struct {
	transactor jtspringproject.SQLTransactor
}

// NewProductDAO constructs a ProductDAO backed by the given SQLTransactor.
//
// MIGRATION_NOTE: Replaces Spring's constructor-based dependency injection of
// the Hibernate SessionFactory.
func NewProductDAO(transactor jtspringproject.SQLTransactor) *ProductDAO {
	return &ProductDAO{transactor: transactor}
}

// GetProducts returns all products stored in the database.
//
// MIGRATION_NOTE: Equivalent to the Hibernate HQL query "from PRODUCT".
func (d *ProductDAO) GetProducts(ctx context.Context) ([]*model.Product, error) {
	var products []*model.Product

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `SELECT id, name, category_id, price, weight, quantity, description, image FROM product`

		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return fmt.Errorf("query products: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			p, scanErr := scanProduct(rows)
			if scanErr != nil {
				return scanErr
			}
			products = append(products, p)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate products: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return products, nil
}

// AddProduct persists a new product and returns it.
//
// MIGRATION_NOTE: Equivalent to Hibernate session.save(product). After the
// INSERT we read back the generated identifier and assign it to the returned
// product, mirroring Hibernate populating the entity's id.
func (d *ProductDAO) AddProduct(ctx context.Context, product *model.Product) (*model.Product, error) {
	if product == nil {
		return nil, errors.New("addProduct: product must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `INSERT INTO product (name, category_id, price, weight, quantity, description, image) VALUES (?, ?, ?, ?, ?, ?, ?)`

		res, err := tx.ExecContext(ctx, query,
			product.Name(),
			product.CategoryID(),
			product.Price(),
			product.Weight(),
			product.Quantity(),
			product.Description(),
			product.Image(),
		)
		if err != nil {
			return fmt.Errorf("insert product: %w", err)
		}

		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("read generated product id: %w", err)
		}
		product.SetID(int(id))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}

// GetProduct returns the product with the given id.
//
// MIGRATION_NOTE: Equivalent to Hibernate session.get(Product.class, id).
// Hibernate returns null when no row matches; here we return a nil product and
// a nil error so callers can distinguish "not found" from an actual failure.
func (d *ProductDAO) GetProduct(ctx context.Context, id int) (*model.Product, error) {
	var product *model.Product

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `SELECT id, name, category_id, price, weight, quantity, description, image FROM product WHERE id = ?`

		row := tx.QueryRowContext(ctx, query, id)
		p, scanErr := scanProduct(row)
		if errors.Is(scanErr, sql.ErrNoRows) {
			product = nil
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		product = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}

// UpdateProduct updates an existing product and returns it.
//
// MIGRATION_NOTE: Equivalent to Hibernate session.update(product).
func (d *ProductDAO) UpdateProduct(ctx context.Context, product *model.Product) (*model.Product, error) {
	if product == nil {
		return nil, errors.New("updateProduct: product must not be nil")
	}

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `UPDATE product SET name = ?, category_id = ?, price = ?, weight = ?, quantity = ?, description = ?, image = ? WHERE id = ?`

		_, err := tx.ExecContext(ctx, query,
			product.Name(),
			product.CategoryID(),
			product.Price(),
			product.Weight(),
			product.Quantity(),
			product.Description(),
			product.Image(),
			product.ID(),
		)
		if err != nil {
			return fmt.Errorf("update product: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}

// DeleteProduct removes the product with the given id, returning true when a
// row was deleted and false when no product with that id existed.
//
// MIGRATION_NOTE: The original loaded the entity with session.get and only
// deleted (returning true) when it was non-null. We preserve that semantics by
// inspecting the number of rows affected by the DELETE instead of performing a
// separate SELECT, which is both correct and cheaper.
func (d *ProductDAO) DeleteProduct(ctx context.Context, id int) (bool, error) {
	var deleted bool

	err := d.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		const query = `DELETE FROM product WHERE id = ?`

		res, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("delete product: %w", err)
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("read affected rows: %w", err)
		}
		deleted = affected > 0
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows, allowing scanProduct
// to be reused for single-row and multi-row queries.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanProduct reads a single product row into a *model.Product.
//
// MIGRATION_NOTE: The exact column set and the Product constructor signature
// depend on the migrated model.Product definition; verify field order against
// model/product.go and the actual database schema.
func scanProduct(s rowScanner) (*model.Product, error) {
	var (
		id          int
		name        string
		categoryID  int
		price       int
		weight      int
		quantity    int
		description string
		image       string
	)

	if err := s.Scan(&id, &name, &categoryID, &price, &weight, &quantity, &description, &image); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("scan product: %w", err)
	}

	product := model.NewProduct(name, categoryID, price, weight, quantity, description, image)
	product.SetID(id)
	return product, nil
}
