// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates CartProductIdTest.java, the JUnit 5 unit test that
// verified the behaviour of the CartProductId composite-key value object.
//
// In the Java world CartProductId was a JPA composite primary key class
// (used via @IdClass / @EmbeddableId) carrying two int fields — cartId and
// productId — together with the value-object equals/hashCode contract that
// JPA requires of composite-key classes.
//
// MIGRATION_NOTE: The production CartProductId type does not yet appear in the
// list of already-migrated symbols. This test file assumes the presence of a
// migrated CartProductId type in this same package exposing:
//
//   - func NewCartProductId(cartID, productID int) *CartProductId
//   - func NewEmptyCartProductId() *CartProductId   (zero-value constructor)
//   - func (c *CartProductId) GetCartID() int
//   - func (c *CartProductId) SetCartID(id int)
//   - func (c *CartProductId) GetProductID() int
//   - func (c *CartProductId) SetProductID(id int)
//   - func (c *CartProductId) Equals(other any) bool
//   - func (c *CartProductId) HashCode() int
//
// If CartProductId is instead implemented as a plain comparable struct (the
// idiomatic Go approach for a composite key), the Equals/HashCode helpers can
// be replaced with direct == comparison and the tests adjusted accordingly.
// This file must be reviewed once CartProductId itself is migrated.
package model

import "testing"

// TestCartProductIdConstructorAndGetters verifies that the two-argument
// constructor stores the cart and product identifiers and that the getters
// return them unchanged. It migrates testConstructorAndGetters.
func TestCartProductIdConstructorAndGetters(t *testing.T) {
	id := NewCartProductId(1, 2)

	if got := id.GetCartID(); got != 1 {
		t.Errorf("GetCartID() = %d, want %d", got, 1)
	}
	if got := id.GetProductID(); got != 2 {
		t.Errorf("GetProductID() = %d, want %d", got, 2)
	}
}

// TestCartProductIdSetters verifies that the setters mutate the cart and
// product identifiers on a zero-value CartProductId. It migrates testSetters.
func TestCartProductIdSetters(t *testing.T) {
	id := NewEmptyCartProductId()
	id.SetCartID(5)
	id.SetProductID(10)

	if got := id.GetCartID(); got != 5 {
		t.Errorf("GetCartID() = %d, want %d", got, 5)
	}
	if got := id.GetProductID(); got != 10 {
		t.Errorf("GetProductID() = %d, want %d", got, 10)
	}
}

// TestCartProductIdEqualsAndHashCode verifies the value-object equality and
// hash-code contract: two identifiers with the same fields are equal and share
// a hash code, while identifiers with different fields are not equal and
// (expectedly) produce different hash codes. It migrates testEqualsAndHashCode.
func TestCartProductIdEqualsAndHashCode(t *testing.T) {
	id1 := NewCartProductId(1, 2)
	id2 := NewCartProductId(1, 2)
	id3 := NewCartProductId(2, 3)

	if !id1.Equals(id2) {
		t.Errorf("Equals(%v, %v) = false, want true", id1, id2)
	}
	if id1.Equals(id3) {
		t.Errorf("Equals(%v, %v) = true, want false", id1, id3)
	}

	if id1.HashCode() != id2.HashCode() {
		t.Errorf("HashCode mismatch: id1=%d id2=%d, want equal", id1.HashCode(), id2.HashCode())
	}
	if id1.HashCode() == id3.HashCode() {
		t.Errorf("HashCode collision: id1=%d id3=%d, want different", id1.HashCode(), id3.HashCode())
	}
}

// TestCartProductIdNotEqualsWithNilOrDifferentType verifies that a
// CartProductId is not considered equal to nil or to a value of an unrelated
// type. It migrates testNotEqualsWithNullOrDifferentType.
//
// MIGRATION_NOTE: Java's assertNotEquals(null, id) and
// assertNotEquals("string", id) rely on Object.equals accepting any argument.
// The Go analogue is Equals(other any), so we pass an untyped nil and an
// unrelated string value directly.
func TestCartProductIdNotEqualsWithNilOrDifferentType(t *testing.T) {
	id := NewCartProductId(1, 2)

	if id.Equals(nil) {
		t.Error("Equals(nil) = true, want false")
	}
	if id.Equals("string") {
		t.Error(`Equals("string") = true, want false`)
	}
}
