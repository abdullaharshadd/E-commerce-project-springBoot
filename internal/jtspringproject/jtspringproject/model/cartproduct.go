// Package model contains the migrated domain entities for the original
// Spring/JPA application.
//
// This file migrates CartProduct.java. The original class was a JPA @Entity
// mapped to the "CART_PRODUCT" join table, modelling the many-to-many
// association between Cart and Product through the association-object pattern.
// It used a composite primary key (CartProductId) declared with @EmbeddedId,
// and two @ManyToOne associations (to Cart and Product) whose foreign-key
// columns ("cart_id", "product_id") were also the components of the composite
// key via @MapsId.
//
// MIGRATION_NOTE: Go has no ORM, @EmbeddedId, @MapsId or @JoinColumn. The
// idiomatic replacement is a plain struct describing the row shape. The
// composite key is expressed as a small value type (CartProductID) that is
// comparable and can be used directly as a Go map key. Persistence concerns
// (table name CART_PRODUCT, columns cart_id / product_id, referential
// integrity) belong in the repository/data-access layer using database/sql,
// not in this struct.
//
// MIGRATION_NOTE: The original entity holds full Cart and Product object
// references (lazy-loaded by Hibernate). This struct instead stores the
// identifiers in the composite key and keeps optional pointers to the loaded
// associations, which the data-access layer may populate. This avoids the
// hidden lazy-loading / N+1 behaviour of the JPA model and makes I/O explicit.
//
// MIGRATION_NOTE: The Product entity (models/Product.java) is referenced here
// but is not part of the provided/already-migrated source set. We reference it
// as *Product from this same package; if it lives in a different package after
// migration, adjust the import and type accordingly (see requires_manual_review).
package model

import "fmt"

// CartProductID is the composite primary key of the CART_PRODUCT join table.
//
// It corresponds to the JPA @EmbeddedId CartProductId, whose components
// (cartId, productId) were derived from the two @ManyToOne associations via
// @MapsId. Because it contains only comparable fields, it can be used directly
// as a Go map key or compared with ==.
type CartProductID struct {
	// CartID is the identifier of the associated Cart (column "cart_id").
	CartID int
	// ProductID is the identifier of the associated Product (column "product_id").
	ProductID int
}

// String returns a human-readable representation of the composite key.
func (id CartProductID) String() string {
	return fmt.Sprintf("CartProductID{CartID:%d, ProductID:%d}", id.CartID, id.ProductID)
}

// CartProduct is the association entity linking a Cart to a Product.
//
// It corresponds to the JPA @Entity mapped to the "CART_PRODUCT" table. The
// composite identifier is held in ID; Cart and Product are optional references
// to the fully-loaded associations, populated by the data-access layer when
// needed.
type CartProduct struct {
	// ID is the composite primary key (cart_id, product_id).
	ID CartProductID

	// Cart is the associated cart. It may be nil when only the identifiers
	// (via ID) have been loaded.
	Cart *Cart

	// Product is the associated product. It may be nil when only the
	// identifiers (via ID) have been loaded.
	//
	// MIGRATION_NOTE: Product is expected to expose an integer identifier via a
	// GetID()/ID accessor. Adjust once the Product model is migrated.
	Product *Product
}

// NewCartProduct constructs a CartProduct linking the given cart and product,
// deriving the composite key from their identifiers.
//
// It mirrors the original two-argument constructor
// CartProduct(Cart, Product), which set the embedded id from cart.getId() and
// product.getId(). It returns an error instead of panicking when either
// association is nil, since a valid composite key cannot be derived otherwise.
func NewCartProduct(cart *Cart, product *Product) (*CartProduct, error) {
	if cart == nil {
		return nil, fmt.Errorf("model: cannot create CartProduct: cart is nil")
	}
	if product == nil {
		return nil, fmt.Errorf("model: cannot create CartProduct: product is nil")
	}

	return &CartProduct{
		ID: CartProductID{
			CartID:    cart.ID,
			ProductID: product.ID,
		},
		Cart:    cart,
		Product: product,
	}, nil
}
