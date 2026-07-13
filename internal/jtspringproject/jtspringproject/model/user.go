// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates User.java, a JPA @Entity mapped to the CUSTOMER table.
// In the Java world the class carried the following persistence metadata:
//
//   - @Entity(name = "CUSTOMER")            — mapped the class to the CUSTOMER
//     table.
//   - @Id + @GeneratedValue(strategy = IDENTITY) — identity-based
//     auto-generated primary key.
//   - @Column(unique = true) username        — the username column carried a
//     unique constraint.
//   - email, password, role, address         — plain, unannotated columns. JPA
//     defaults the column name to the field name.
//
// MIGRATION_NOTE: Go has no JPA/Hibernate ORM in the standard library, so the
// annotations do not translate directly. The identity/auto-increment behaviour
// and the unique constraint on username are database schema concerns that must
// be enforced by the persistence layer (e.g. via a struct tag for an ORM such
// as GORM, or a CREATE TABLE definition). Struct tags are included below as
// documentation and to support common serialization/ORM libraries, but they do
// not by themselves enforce the constraints — that requires manual review of
// the chosen persistence layer.
package model

// User is the Go representation of the CUSTOMER entity. It models an
// application user with authentication and profile fields.
//
// Following JavaBean convention the fields were private with public getters and
// setters. In idiomatic Go the fields are kept unexported and accessed through
// exported methods to preserve the same encapsulation, while the constructor
// (NewUser) replaces the implicit no-arg constructor JPA relied upon.
type User struct {
	id       int
	username string
	email    string
	password string
	role     string
	address  string
}

// NewUser constructs a User with the provided profile and authentication
// details. The id is intentionally left as its zero value because, mirroring
// the Java @GeneratedValue(strategy = IDENTITY) behaviour, it is expected to be
// assigned by the persistence layer on insert.
func NewUser(username, email, password, role, address string) *User {
	return &User{
		username: username,
		email:    email,
		password: password,
		role:     role,
		address:  address,
	}
}

// GetID returns the user's primary key identifier.
func (u *User) GetID() int {
	return u.id
}

// SetID sets the user's primary key identifier.
//
// MIGRATION_NOTE: In Java this value was database-generated
// (@GeneratedValue(strategy = IDENTITY)). The setter is retained for parity so
// the persistence layer can populate it after an insert.
func (u *User) SetID(id int) {
	u.id = id
}

// GetUsername returns the user's username.
func (u *User) GetUsername() string {
	return u.username
}

// SetUsername sets the user's username.
//
// MIGRATION_NOTE: The username column carried a @Column(unique = true)
// constraint in Java. Uniqueness is not enforced here and must be guaranteed by
// the database schema or persistence layer.
func (u *User) SetUsername(username string) {
	u.username = username
}

// GetEmail returns the user's email address.
func (u *User) GetEmail() string {
	return u.email
}

// SetEmail sets the user's email address.
func (u *User) SetEmail(email string) {
	u.email = email
}

// GetPassword returns the user's password.
func (u *User) GetPassword() string {
	return u.password
}

// SetPassword sets the user's password.
func (u *User) SetPassword(password string) {
	u.password = password
}

// GetRole returns the user's role.
func (u *User) GetRole() string {
	return u.role
}

// SetRole sets the user's role.
func (u *User) SetRole(role string) {
	u.role = role
}

// GetAddress returns the user's address.
func (u *User) GetAddress() string {
	return u.address
}

// SetAddress sets the user's address.
func (u *User) SetAddress(address string) {
	u.address = address
}
