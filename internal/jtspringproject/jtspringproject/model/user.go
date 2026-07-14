// Package model contains the migrated domain entities for the original
// Spring/JPA application.
//
// This file migrates User.java. The original class was a JPA @Entity mapped to
// the "CUSTOMER" table with an auto-generated integer primary key ("id"
// column), a unique username, and scalar profile/credential attributes (email,
// password, role, address).
//
// MIGRATION_NOTE: Go has no ORM, JPA annotations or JavaBean getter/setter
// conventions. The idiomatic replacement is a plain struct describing the row
// shape with exported fields accessed directly. Persistence concerns are
// handled explicitly by the repository/data-access layer using database/sql:
//
//   - @Entity(name = "CUSTOMER") -> the repository targets the "CUSTOMER" table.
//   - @Id + @GeneratedValue(IDENTITY) -> ID is populated by the database on
//     insert; the repository reads it back (e.g. via LastInsertId).
//   - @Column(unique = true) on username -> enforced by a UNIQUE constraint in
//     the database schema, not by the struct. Callers should handle the
//     resulting duplicate-key error at insert time.
//   - JavaBean getters/setters -> exported struct fields.
package model

// User represents an application user (the CUSTOMER table row) with credentials,
// role, and profile information.
type User struct {
	// ID is the auto-generated primary key. It is zero for a User that has not
	// yet been persisted and is populated by the database on insert.
	ID int

	// Username is the unique login name. Uniqueness is enforced by a database
	// constraint, not by this struct.
	Username string

	// Email is the user's email address.
	Email string

	// Password is the user's (typically hashed) credential.
	Password string

	// Role is the authorization role assigned to the user.
	Role string

	// Address is the user's profile address.
	Address string
}

// NewUser constructs a User with the given profile and credential fields.
//
// The ID is intentionally omitted: it is assigned by the database when the User
// is persisted. Use the repository layer to insert the User and read the
// generated ID back.
func NewUser(username, email, password, role, address string) *User {
	return &User{
		Username: username,
		Email:    email,
		Password: password,
		Role:     role,
		Address:  address,
	}
}
