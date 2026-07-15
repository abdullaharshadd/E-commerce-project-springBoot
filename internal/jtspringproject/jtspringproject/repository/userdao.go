package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
)

// UserDAO provides CRUD and lookup operations for the User entity against the
// database.
//
// MIGRATION_NOTE: The original was a Spring @Repository backed by Hibernate's
// SessionFactory. In Go there is no ORM session bound to a thread; instead we
// use the SQLTransactor abstraction (see hibernateconfiguration.go) which owns
// the *sql.DB and exposes WithinTransaction to replace Spring's declarative
// @Transactional AOP. Each method opens a transaction explicitly and takes a
// context.Context as its first parameter, propagating cancellation.
//
// MIGRATION_NOTE: The HQL entity name "CUSTOMER" refers to the User entity
// mapped to the customer table. We query the concrete table directly here.
type UserDAO struct {
	tx jtspringproject.SQLTransactor
}

// NewUserDAO constructs a UserDAO backed by the given SQLTransactor.
//
// MIGRATION_NOTE: Replaces Spring's constructor-based dependency injection of
// the SessionFactory. Wire this up explicitly at application startup.
func NewUserDAO(tx jtspringproject.SQLTransactor) *UserDAO {
	return &UserDAO{tx: tx}
}

// scanUser reads a single User row from the given *sql.Rows or *sql.Row-like
// scanner into a *model.User.
//
// MIGRATION_NOTE: The column list must match the customer table schema. Adjust
// the columns to reflect the actual mapping of the User entity. Human review
// recommended to confirm column ordering.
func scanUser(scan func(dest ...any) error) (*model.User, error) {
	var (
		id       int
		username string
		password string
		email    string
		address  string
		role     string
	)
	if err := scan(&id, &username, &password, &email, &address, &role); err != nil {
		return nil, err
	}
	return model.NewUser(id, username, password, email, address, role), nil
}

const userColumns = "id, username, password, email, address, role"

// GetAllUser returns every User stored in the database.
func (d *UserDAO) GetAllUser(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	err := d.tx.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, "SELECT "+userColumns+" FROM customer")
		if err != nil {
			return fmt.Errorf("query all users: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			user, err := scanUser(rows.Scan)
			if err != nil {
				return fmt.Errorf("scan user: %w", err)
			}
			users = append(users, user)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate users: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

// SaveUser inserts a new User or updates an existing one (upsert), then returns
// the persisted User.
//
// MIGRATION_NOTE: Hibernate's saveOrUpdate decides insert vs. update based on
// the entity identifier. Here we treat a zero ID as an insert and a non-zero ID
// as an update. The generated ID is written back to the returned User on insert.
func (d *UserDAO) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, errors.New("save user: user must not be nil")
	}

	err := d.tx.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if user.ID == 0 {
			res, err := tx.ExecContext(ctx,
				"INSERT INTO customer (username, password, email, address, role) VALUES (?, ?, ?, ?, ?)",
				user.Username, user.Password, user.Email, user.Address, user.Role)
			if err != nil {
				return fmt.Errorf("insert user: %w", err)
			}
			id, err := res.LastInsertId()
			if err != nil {
				return fmt.Errorf("read inserted user id: %w", err)
			}
			user.ID = int(id)
			return nil
		}

		_, err := tx.ExecContext(ctx,
			"UPDATE customer SET username = ?, password = ?, email = ?, address = ?, role = ? WHERE id = ?",
			user.Username, user.Password, user.Email, user.Address, user.Role, user.ID)
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserExists reports whether a User with the given username exists.
func (d *UserDAO) UserExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := d.tx.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM customer WHERE username = ?)", username)
		if err := row.Scan(&exists); err != nil {
			return fmt.Errorf("check user exists: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetUserByUsername returns the User with the given username. The boolean
// result is false when no matching user exists.
//
// MIGRATION_NOTE: The original returned null on NoResultException. Idiomatic Go
// uses the (T, bool) pattern to signal absence instead of a nil sentinel.
func (d *UserDAO) GetUserByUsername(ctx context.Context, username string) (*model.User, bool, error) {
	var user *model.User
	err := d.tx.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT "+userColumns+" FROM customer WHERE username = ?", username)
		u, err := scanUser(row.Scan)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("get user by username: %w", err)
		}
		user = u
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, nil
	}
	return user, true, nil
}

// GetUserByID returns the User with the given id. The boolean result is false
// when no matching user exists.
//
// MIGRATION_NOTE: Hibernate's Session.get returns null when the entity is not
// found; we surface that as a (nil, false) result rather than a nil sentinel.
func (d *UserDAO) GetUserByID(ctx context.Context, id int) (*model.User, bool, error) {
	var user *model.User
	err := d.tx.WithinTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			"SELECT "+userColumns+" FROM customer WHERE id = ?", id)
		u, err := scanUser(row.Scan)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("get user by id: %w", err)
		}
		user = u
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, nil
	}
	return user, true, nil
}
