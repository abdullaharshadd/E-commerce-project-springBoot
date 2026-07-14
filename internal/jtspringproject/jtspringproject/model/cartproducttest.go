package model

// This file migrates CartProductTest.java.
//
// The original was a JUnit 5 unit test for the CartProduct JPA entity. It
// validated the constructor (which derives the composite CartProductId key
// from the supplied Cart and Product), the getters, and the setters.
//
// MIGRATION_NOTE: The migrated CartProduct (see cartproduct.go) is constructed
// via NewCartProduct, which takes a *Cart and a *Product and derives the
// composite CartProductID from their IDs. Fields are exported and accessed
// directly, so the Java getter/setter tests collapse into direct field
// assertions plus a table-driven test over the constructor. Go has no
// zero-arg-constructor + setter idiom, so the Java testSetters case is
// re-expressed by constructing the CartProduct with different inputs and
// verifying the derived key and embedded references.

import (
	"testing"
)

// TestNewCartProductConstructorAndFields verifies that NewCartProduct derives
// the composite CartProductID from the supplied Cart and Product IDs and that
// the embedded Cart and Product references are preserved.
func TestNewCartProductConstructorAndFields(t *testing.T) {
	tests := []struct {
		name          string
		cartID        int
		productID     int
		wantCartID    int
		wantProductID int
	}{
		{
			name:          "constructor derives id from cart and product",
			cartID:        1,
			productID:     100,
			wantCartID:    1,
			wantProductID: 100,
		},
		{
			name:          "setter-equivalent with different ids",
			cartID:        2,
			productID:     200,
			wantCartID:    2,
			wantProductID: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cart := NewCart()
			cart.ID = tt.cartID

			product := NewProduct()
			product.ID = tt.productID

			cartProduct := NewCartProduct(cart, product)

			if cartProduct == nil {
				t.Fatal("NewCartProduct returned nil")
			}

			if got := cartProduct.ID.CartID; got != tt.wantCartID {
				t.Errorf("cartProduct.ID.CartID = %d, want %d", got, tt.wantCartID)
			}
			if got := cartProduct.ID.ProductID; got != tt.wantProductID {
				t.Errorf("cartProduct.ID.ProductID = %d, want %d", got, tt.wantProductID)
			}

			if cartProduct.Cart != cart {
				t.Errorf("cartProduct.Cart = %v, want %v", cartProduct.Cart, cart)
			}
			if cartProduct.Product != product {
				t.Errorf("cartProduct.Product = %v, want %v", cartProduct.Product, product)
			}
		})
	}
}
