package model

// This file migrates CartProductIdTest.java.
//
// The original was a JUnit 5 unit test for the CartProductId JPA composite-key
// (@Embeddable/@IdClass) identity class. It validated the constructor, getters,
// setters, and the equals/hashCode contract required by JPA for composite keys.
//
// MIGRATION_NOTE: The migrated CartProductID (see cartproductid.go) is a plain
// value struct with exported CartID/ProductID fields. There are no getters or
// setters — fields are accessed directly, so the getter/setter tests collapse
// into direct field assertions. Go's == operator and the provided Equal method
// replace Java's equals/hashCode contract; there is no hashCode equivalent
// because Go maps hash struct keys natively. The "not equal to null / different
// type" test has no Go analogue: CartProductID is a concrete value type, so it
// is impossible to compare it against nil or an unrelated type at compile time.
// That test is therefore intentionally omitted (documented below).

import "testing"

// TestCartProductIDConstructorAndFields verifies that NewCartProductID stores
// the supplied cart and product identifiers on the resulting value.
func TestCartProductIDConstructorAndFields(t *testing.T) {
	tests := []struct {
		name          string
		cartID        int
		productID     int
		wantCartID    int
		wantProductID int
	}{
		{name: "one and two", cartID: 1, productID: 2, wantCartID: 1, wantProductID: 2},
		{name: "five and ten", cartID: 5, productID: 10, wantCartID: 5, wantProductID: 10},
		{name: "zero values", cartID: 0, productID: 0, wantCartID: 0, wantProductID: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewCartProductID(tt.cartID, tt.productID)

			if id.CartID != tt.wantCartID {
				t.Errorf("CartID = %d, want %d", id.CartID, tt.wantCartID)
			}
			if id.ProductID != tt.wantProductID {
				t.Errorf("ProductID = %d, want %d", id.ProductID, tt.wantProductID)
			}
		})
	}
}

// TestCartProductIDEqual verifies the equality semantics of CartProductID,
// replacing the original equals/hashCode contract test.
func TestCartProductIDEqual(t *testing.T) {
	tests := []struct {
		name string
		a    CartProductID
		b    CartProductID
		want bool
	}{
		{
			name: "identical ids are equal",
			a:    NewCartProductID(1, 2),
			b:    NewCartProductID(1, 2),
			want: true,
		},
		{
			name: "different ids are not equal",
			a:    NewCartProductID(1, 2),
			b:    NewCartProductID(2, 3),
			want: false,
		},
		{
			name: "same cart different product",
			a:    NewCartProductID(1, 2),
			b:    NewCartProductID(1, 3),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}

			// MIGRATION_NOTE: Go's == on comparable structs mirrors the
			// equals/hashCode contract. Because CartProductID is used as a
			// map key, == and Equal must agree.
			if got := (tt.a == tt.b); got != tt.want {
				t.Errorf("(a == b) = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCartProductIDAsMapKey confirms that equal CartProductID values collide as
// map keys and unequal values do not. This is the idiomatic Go replacement for
// the Java hashCode() contract assertions, since Go maps hash struct keys
// natively rather than through a user-defined hashCode method.
func TestCartProductIDAsMapKey(t *testing.T) {
	id1 := NewCartProductID(1, 2)
	id2 := NewCartProductID(1, 2)
	id3 := NewCartProductID(2, 3)

	m := map[CartProductID]string{}
	m[id1] = "first"

	if got, ok := m[id2]; !ok || got != "first" {
		t.Errorf("equal key lookup: got %q ok=%v, want \"first\" ok=true", got, ok)
	}
	if _, ok := m[id3]; ok {
		t.Errorf("unequal key %+v should not be present in map", id3)
	}
}
