// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates CartProductTest.java, the JUnit 5 unit test that verified
// the constructor, getters, and setters of the CartProduct JPA entity,
// including its composite key (CartProductId) and its associations with the
// Cart and Product entities.
//
// In the Java world CartProduct was a JPA join entity:
//
//	@Entity
//	class CartProduct {
//	    @EmbeddedId CartProductId id;
//	    @ManyToOne Cart cart;
//	    @ManyToOne Product product;
//	}
//
// MIGRATION_NOTE: The production CartProduct, Cart, and CartProductId types do
// not yet appear in the list of already-migrated symbols. This test assumes
// the presence of migrated types in this same package exposing the following
// idiomatic Go API (constructor + getters/setters), which mirror the original
// JPA POJO surface:
//
//   - type Cart with NewCart() *Cart, GetID() int, SetID(int)
//   - type Product with NewProduct() *Product, GetID() int, SetID(int)
//   - type CartProductId with NewCartProductId(cartID, productID int) *CartProductId,
//     GetCartID() int, GetProductID() int
//   - type CartProduct with:
//       NewCartProduct(cart *Cart, product *Product) *CartProduct
//       NewEmptyCartProduct() *CartProduct  (zero-value constructor equivalent to new CartProduct())
//       GetID() *CartProductId, SetID(*CartProductId)
//       GetCart() *Cart, SetCart(*Cart)
//       GetProduct() *Product, SetProduct(*Product)
//
// If the migrated production names differ, update the references below during
// manual review.
package model

import "testing"

// TestCartProductConstructorAndGetters verifies that constructing a CartProduct
// from a Cart and a Product derives the correct composite identifier and
// preserves the entity associations. It mirrors the Java testConstructorAndGetters.
func TestCartProductConstructorAndGetters(t *testing.T) {
	cart := NewCart()
	cart.SetID(1)

	product := NewProduct()
	product.SetID(100)

	cartProduct := NewCartProduct(cart, product)

	id := cartProduct.GetID()
	if id == nil {
		t.Fatal("CartProductId should not be nil")
	}

	if got, want := id.GetCartID(), 1; got != want {
		t.Errorf("Cart ID should match: got %d, want %d", got, want)
	}
	if got, want := id.GetProductID(), 100; got != want {
		t.Errorf("Product ID should match: got %d, want %d", got, want)
	}

	if got := cartProduct.GetCart(); got != cart {
		t.Errorf("Cart should match: got %v, want %v", got, cart)
	}
	if got := cartProduct.GetProduct(); got != product {
		t.Errorf("Product should match: got %v, want %v", got, product)
	}
}

// TestCartProductSetters verifies that the CartProduct setters correctly update
// the composite identifier and entity associations. It mirrors the Java
// testSetters.
func TestCartProductSetters(t *testing.T) {
	cartProduct := NewEmptyCartProduct()

	cart := NewCart()
	cart.SetID(2)

	product := NewProduct()
	product.SetID(200)

	cartProduct.SetCart(cart)
	cartProduct.SetProduct(product)
	cartProduct.SetID(NewCartProductId(cart.GetID(), product.GetID()))

	id := cartProduct.GetID()
	if id == nil {
		t.Fatal("CartProductId should not be nil")
	}

	if got, want := id.GetCartID(), 2; got != want {
		t.Errorf("Cart ID should match: got %d, want %d", got, want)
	}
	if got, want := id.GetProductID(), 200; got != want {
		t.Errorf("Product ID should match: got %d, want %d", got, want)
	}
	if got := cartProduct.GetCart(); got != cart {
		t.Errorf("Cart should match: got %v, want %v", got, cart)
	}
	if got := cartProduct.GetProduct(); got != product {
		t.Errorf("Product should match: got %v, want %v", got, product)
	}
}
