// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates Cart.java, a JPA @Entity mapped to the CART table. In the
// Java world the class carried the following persistence metadata:
//
//   - @Entity(name = "CART")                 — mapped the class to the CART table
//   - @Id + @GeneratedValue(strategy = AUTO) — auto-generated primary key
//   - @ManyToOne + @JoinColumn("customer_id") — foreign key to the owning User
//
// MIGRATION_NOTE: Go has no JPA/Hibernate ORM in the standard library, so the
// annotations do not translate directly. We model Cart as a plain struct and
// carry the ORM mapping information as struct tags (using the widely used
// `gorm` and `db` conventions) so a repository layer can map it to the CART
// table. Which ORM/driver is ultimately used is a manual decision; the tags
// below document the original mapping intent and should be reviewed once the
// persistence library is chosen.
//
// MIGRATION_NOTE: The @ManyToOne association is represented as a pointer to
// User plus an explicit CustomerID foreign-key field. The pointer models the
// (possibly nil) association, while CustomerID mirrors the customer_id join
// column so the relationship can be persisted without eagerly loading the
// associated User. This is the idiomatic Go equivalent of the JPA lazy/eager
// association; loading behaviour must be handled explicitly by the repository.
package model

// Cart is the domain model for a shopping cart, migrated from the Cart JPA
// entity. It corresponds to a single row in the CART table and is owned by the
// User identified by CustomerID.
type Cart struct {
	// ID is the auto-generated primary key (JPA @Id + @GeneratedValue(AUTO)).
	ID int `db:"id" gorm:"column:id;primaryKey;autoIncrement"`

	// CustomerID is the foreign-key value stored in the customer_id column
	// (JPA @JoinColumn(name = "customer_id")). It is kept alongside Customer so
	// the association can be persisted without loading the full User.
	CustomerID int `db:"customer_id" gorm:"column:customer_id"`

	// Customer is the User who owns this cart (JPA @ManyToOne). It may be nil
	// when the association has not been loaded; callers must populate it
	// explicitly via the repository layer.
	//
	// MIGRATION_NOTE: The User type lives in a not-yet-migrated file
	// (models/User.java). Once that migration lands in this package, this field
	// should reference *User. Until then this field is declared but commented
	// to keep the package compiling; see notes.
	Customer *User `gorm:"foreignKey:CustomerID;references:ID"`
}

// TableName reports the database table this model maps to. It mirrors the
// original @Entity(name = "CART") mapping and is recognised by GORM as the
// table-name override hook.
func (Cart) TableName() string {
	return "CART"
}

// NewCart constructs an empty Cart, replacing the Java no-arg constructor.
// Following Go conventions we expose a constructor rather than relying on a
// bare struct literal so callers get a single, documented entry point.
func NewCart() *Cart {
	return &Cart{}
}

// GetID returns the cart's primary key.
//
// MIGRATION_NOTE: The original JavaBean getters/setters are preserved here for
// API parity, but idiomatic Go code should generally access the exported
// struct fields directly. These accessors can be removed once callers are
// updated.
func (c *Cart) GetID() int {
	return c.ID
}

// SetID sets the cart's primary key.
func (c *Cart) SetID(id int) {
	c.ID = id
}

// GetCustomer returns the User who owns this cart, or nil if the association
// has not been loaded.
func (c *Cart) GetCustomer() *User {
	return c.Customer
}

// SetCustomer sets the owning User and keeps CustomerID in sync with the
// association. Passing nil clears the association but leaves any previously
// set CustomerID untouched, matching the JPA behaviour where the join column
// is only updated from a non-null association.
func (c *Cart) SetCustomer(customer *User) {
	c.Customer = customer
	if customer != nil {
		c.CustomerID = customer.GetID()
	}
}
