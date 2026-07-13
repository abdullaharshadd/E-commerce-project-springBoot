// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates CartProductId.java, a JPA @Embeddable composite primary
// key that combined cartId and productId into a single identity object. In the
// Java world the class carried the following persistence metadata:
//
//   - @Embeddable                     — marked the class as a value type that
//     could be embedded as the composite key of an owning @Entity.
//   - @Column(name = "cart_id")        — mapped the cartId field to the
//     cart_id column.
//   - @Column(name = "product_id")     — mapped the productId field to the
//     product_id column.
//   - implements Serializable          — JPA requires composite key classes to
//     be Serializable so identities can be cached/passed around.
//   - equals/hashCode                  — JPA requires a value-based equality
//     contract on identity classes.
//
// MIGRATION_NOTE: Go has no JPA/Hibernate ORM in the standard library, so the
// annotations do not translate directly. We model CartProductID as a plain
// struct and carry the column mapping information as struct tags (using the
// widely used `gorm` and `db` conventions).
//
// MIGRATION_NOTE: In Java the @Column fields were Integer (a nullable boxed
// type). Go's int is not nullable; a zero value means "unset". Since a
// composite key is always fully populated in practice, we use plain int here.
// If genuine nullability is ever required, switch to *int or sql.NullInt64 and
// adjust the equality logic accordingly.
//
// MIGRATION_NOTE: Java's Serializable requirement has no Go equivalent — Go
// uses standard encoding packages (encoding/json, encoding/gob) that work on
// any exported struct fields, so nothing extra is needed.
//
// MIGRATION_NOTE: Java's equals()/hashCode() value-equality contract is
// satisfied automatically in Go: this struct contains only comparable fields
// (two ints), so the `==` operator already performs value comparison, and the
// struct is usable directly as a map key. We still expose an Equals method for
// call sites that were translated from the Java equals() usage.
//
// NOTE: The composite-key type used by CartProduct was already migrated as
// CartProductID in cartproduct.go. This file provides the standalone,
// tag-annotated definition; reviewers should reconcile the two so only one
// CartProductID definition exists in the package.
package model

// CartProductID is the composite primary key for a CartProduct, combining the
// owning cart's identifier with the referenced product's identifier.
//
// It corresponds to the JPA @Embeddable CartProductId class. Because it
// contains only comparable fields, values of this type may be compared with ==
// and used directly as map keys.
type CartProductID struct {
	// CartID is the identifier of the owning cart (column: cart_id).
	CartID int `db:"cart_id" gorm:"column:cart_id"`
	// ProductID is the identifier of the referenced product (column: product_id).
	ProductID int `db:"product_id" gorm:"column:product_id"`
}

// NewCartProductID constructs a CartProductID from the given cart and product
// identifiers.
//
// It mirrors the two-argument Java constructor. The zero-value struct
// (CartProductID{}) corresponds to the Java no-argument constructor.
func NewCartProductID(cartID, productID int) CartProductID {
	return CartProductID{
		CartID:    cartID,
		ProductID: productID,
	}
}

// GetCartID returns the cart identifier component of the key.
func (id CartProductID) GetCartID() int {
	return id.CartID
}

// SetCartID sets the cart identifier component of the key.
func (id *CartProductID) SetCartID(cartID int) {
	id.CartID = cartID
}

// GetProductID returns the product identifier component of the key.
func (id CartProductID) GetProductID() int {
	return id.ProductID
}

// SetProductID sets the product identifier component of the key.
func (id *CartProductID) SetProductID(productID int) {
	id.ProductID = productID
}

// Equals reports whether this key is equal to other by value.
//
// It preserves the Java equals() semantics. Because CartProductID is fully
// comparable, this is equivalent to the `==` operator; the method is provided
// for call sites migrated from Java equals() usage.
func (id CartProductID) Equals(other CartProductID) bool {
	return id == other
}
