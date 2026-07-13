// Package repository contains the Go migration of the JtSpringProject DAO
// layer (the com.jtspringproject.JtSpringProject.dao package).
//
// This file migrates userDao.java, the Hibernate-backed Data Access Object
// providing CRUD and query operations for the User entity.
//
// In the Java world userDao was a Spring @Repository:
//
//	@Repository
//	public class userDao {
//	    private final SessionFactory sessionFactory;
//	    @Transactional public List<User> getAllUser() { ... }
//	    @Transactional public User saveUser(User user) { ... }
//	    @Transactional public boolean userExists(String username) { ... }
//	    @Transactional public User getUserByUsername(String username) { ... }
//	    @Transactional public User getUserById(int id) { ... }
//	}
//
// MIGRATION_NOTE: Hibernate's SessionFactory / getCurrentSession() and Spring's
// declarative @Transactional are replaced here by an explicit *sql.DB handle
// and manual context propagation. Transaction management is left to the caller
// (or a service layer) rather than being woven in via annotations; each method
// accepts a context.Context as its first parameter for cancellation and
// deadline propagation. The Java HQL entity name "CUSTOMER" maps to the
// underlying customer table.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrUserNotFound is returned when a user lookup yields no matching row.
//
// This mirrors the Java code's handling of NoResultException (which returned
// null); Go callers should check for this sentinel using errors.Is.
var ErrUserNotFound = errors.New("user not found")

// User is the Go representation of the com.jtspringproject.JtSpringProject.models.User
// entity as persisted by this repository.
//
// MIGRATION_NOTE: The original User model lived in the models package and was
// mapped by Hibernate. Only the columns exercised by userDao are modelled here;
// adjust the fields and column mapping to match the real schema during review.
type User struct {
	id       int
	username string
	password string
	email    string
	address  string
	role     string
}

// NewUser constructs a User with all persisted fields populated.
func NewUser(id int, username, password, email, address, role string) *User {
	return &User{
		id:       id,
		username: username,
		password: password,
		email:    email,
		address:  address,
		role:     role,
	}
}

// GetID returns the user's primary key.
func (u *User) GetID() int { return u.id }

// SetID sets the user's primary key.
func (u *User) SetID(id int) { u.id = id }

// GetUsername returns the user's username.
func (u *User) GetUsername() string { return u.username }

// SetUsername sets the user's username.
func (u *User) SetUsername(username string) { u.username = username }

// GetPassword returns the user's password.
func (u *User) GetPassword() string { return u.password }

// SetPassword sets the user's password.
func (u *User) SetPassword(password string) { u.password = password }

// GetEmail returns the user's email address.
func (u *User) GetEmail() string { return u.email }

// SetEmail sets the user's email address.
func (u *User) SetEmail(email string) { u.email = email }

// GetAddress returns the user's address.
func (u *User) GetAddress() string { return u.address }

// SetAddress sets the user's address.
func (u *User) SetAddress(address string) { u.address = address }

// GetRole returns the user's role.
func (u *User) GetRole() string { return u.role }

// SetRole sets the user's role.
func (u *User) SetRole(role string) { u.role = role }

// UserRepository provides CRUD and query operations for User entities.
//
// It replaces the Java userDao @Repository. The Hibernate SessionFactory is
// replaced by a *sql.DB, injected via NewUserRepository (constructor injection,
// matching the original constructor-based DI).
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository constructs a UserRepository backed by the given database
// handle. It returns an error if db is nil.
func NewUserRepository(db *sql.DB) (*UserRepository, error) {
	if db == nil {
		return nil, errors.New("repository: db must not be nil")
	}
	return &UserRepository{db: db}, nil
}

// scanUser reads a single User row from the given scanner.
func scanUser(scan func(dest ...any) error) (*User, error) {
	var u User
	if err := scan(&u.id, &u.username, &u.password, &u.email, &u.address, &u.role); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAllUser returns every user in the store.
//
// This migrates getAllUser(): "from CUSTOMER".
func (r *UserRepository) GetAllUser(ctx context.Context) ([]*User, error) {
	const query = `SELECT id, username, password, email, address, role FROM customer`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: query all users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := scanUser(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("repository: scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate users: %w", err)
	}
	return users, nil
}

// SaveUser persists the given user, inserting a new row or updating an existing
// one depending on whether its ID is set.
//
// This migrates saveUser(), which used Hibernate's saveOrUpdate upsert idiom.
//
// MIGRATION_NOTE: Hibernate's saveOrUpdate decided insert-vs-update based on
// the entity's identifier and persistence state. Here we approximate that by
// treating a zero ID as an insert and a non-zero ID as an update. Review this
// against the real primary-key/generation strategy; a database-native UPSERT
// (e.g. INSERT ... ON DUPLICATE KEY UPDATE) may be more faithful.
func (r *UserRepository) SaveUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, errors.New("repository: user must not be nil")
	}

	if user.id == 0 {
		const insert = `INSERT INTO customer (username, password, email, address, role) VALUES (?, ?, ?, ?, ?)`
		res, err := r.db.ExecContext(ctx, insert, user.username, user.password, user.email, user.address, user.role)
		if err != nil {
			return nil, fmt.Errorf("repository: insert user: %w", err)
		}
		if newID, err := res.LastInsertId(); err == nil {
			user.id = int(newID)
		}
		return user, nil
	}

	const update = `UPDATE customer SET username = ?, password = ?, email = ?, address = ?, role = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, update, user.username, user.password, user.email, user.address, user.role, user.id); err != nil {
		return nil, fmt.Errorf("repository: update user: %w", err)
	}
	return user, nil
}

// UserExists reports whether a user with the given username exists.
//
// This migrates userExists(), returning (bool, error) so callers can
// distinguish a genuine "does not exist" from a query failure.
func (r *UserRepository) UserExists(ctx context.Context, username string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM customer WHERE username = ?)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, username).Scan(&exists); err != nil {
		return false, fmt.Errorf("repository: check user exists: %w", err)
	}
	return exists, nil
}

// GetUserByUsername returns the user with the given username.
//
// This migrates getUserByUsername(). The Java version caught NoResultException
// and returned null; here we return ErrUserNotFound so callers can check with
// errors.Is.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	const query = `SELECT id, username, password, email, address, role FROM customer WHERE username = ?`

	u, err := scanUser(r.db.QueryRowContext(ctx, query, username).Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: get user by username: %w", err)
	}
	return u, nil
}

// GetUserByID returns the user with the given primary key.
//
// This migrates getUserById(). Hibernate's Session.get returned null when the
// entity was absent; here we return ErrUserNotFound in that case.
func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	const query = `SELECT id, username, password, email, address, role FROM customer WHERE id = ?`

	u, err := scanUser(r.db.QueryRowContext(ctx, query, id).Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: get user by id: %w", err)
	}
	return u, nil
}
