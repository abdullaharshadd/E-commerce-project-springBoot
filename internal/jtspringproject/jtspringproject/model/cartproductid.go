// Package model contains the migrated domain entities for the original
// Spring/JPA application.
//
// This file migrates CartProductId.java. The original class was a JPA
// @Embeddable composite primary key combining the cart ID and product ID for
// the CART_PRODUCT join entity. It implemented Serializable (a JPA requirement
// for embedded IDs) and overrode equals/hashCode to satisfy the identity
// contract used by the persistence provider.
//
// MIGRATION_NOTE: Go has no @Embeddable, @Column, Serializable or the
// equals/hashCode contract. The idiomatic replacement is a small, comparable
// value type. Because CartProductID contains only comparable fields (two
// integers), Go's built-in == operator provides value equality for free, and
// the struct is directly usable as a map key — this removes the need for the
// hand-written equals/hashCode boilerplate the Java version required.
//
//   - @Column(name = "cart_id")    -> CartID field; column mapping is handled
//     explicitly by the repository/data-access layer.
//   - @Column(name = "product_id") -> ProductID field; column mapping is handled
//     explicitly by the repository/data-access layer.
//   - equals / hashCode            -> covered by Go value semantics (==) and map
//     key usability for a struct of comparable fields.
//   - Serializable                 -> not needed; JPA key serialization has no
//     Go analogue.
//
// The Java Integer type was nullable; the components of a composite primary key
// are, in practice, always non-null, so plain int is used here. If a genuine
// "unset" distinction is ever required, switch to *int or a (int, bool) pair.
package model

import "fmt"

// CartProductID is the composite primary key of the CART_PRODUCT join entity,
// combining the owning cart's ID with the referenced product's ID.
//
// It is a comparable value type: two CartProductID values are equal when both
// components match, and it may be used directly as a map key.
type CartProductID struct {
	// CartID is the identifier of the owning cart (originally the "cart_id"
	// column).
	CartID int
	// ProductID is the identifier of the referenced product (originally the
	// "product_id" column).
	ProductID int
}

// NewCartProductID constructs a CartProductID from the given cart and product
// identifiers.
func NewCartProductID(cartID, productID int) CartProductID {
	return CartProductID{CartID: cartID, ProductID: productID}
}

// Equal reports whether two CartProductID values refer to the same cart and
// product. It mirrors the original equals method; for most call sites the
// built-in == operator is equivalent and preferable.
func (id CartProductID) Equal(other CartProductID) bool {
	return id == other
}

// String returns a human-readable representation of the composite key.
func (id CartProductID) String() string {
	return fmt.Sprintf("CartProductID{cartID=%d, productID=%d}", id.CartID, id.ProductID)
}
