// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates CartProduct.java, a JPA @Entity mapped to the
// CART_PRODUCT table. It models the join-table-as-entity between Cart and
// Product (a many-to-many relationship). In the Java world the class carried
// the following persistence metadata:
//
//   - @Entity + @Table(name = "CART_PRODUCT") — mapped the class to the
//     CART_PRODUCT table.
//   - @EmbeddedId CartProductId              — composite primary key made up
//     of (cartId, productId).
//   - @ManyToOne + @MapsId("cartId") + @JoinColumn("cart_id")    — FK to Cart,
//     whose id also feeds the composite key's cartId part.
//   - @ManyToOne + @MapsId("productId") + @JoinColumn("product_id") — FK to
//     Product, whose id also feeds the composite key's productId part.
//
// MIGRATION_NOTE: Go has no JPA/Hibernate ORM in the standard library, so the
// annotations do not translate directly. We model CartProduct as a plain
// struct and carry the ORM mapping information as struct tags (using the widely
// used `gorm` and `db` conventions). The @EmbeddedId composite key is modeled
// as a dedicated CartProductID value type, mirroring the original
// CartProductId embeddable.
//
// MIGRATION_NOTE: The @MapsId behaviour — where Hibernate derives the composite
// key parts from the associated entity IDs — is not automatic in Go. We
// replicate it explicitly in the NewCartProduct constructor, which populates
// the composite ID from the supplied Cart and Product. Any code that mutates
// the associated Cart/Product after construction must keep the composite ID in
// sync (see SetCart/SetProduct, which do this explicitly).
//
// MIGRATION_NOTE: The Product entity had not been migrated at the time this
// file was written. This file references model.Product; ensure product.go
// exists in this package and exposes a GetID() method returning the product's
// primary key. Review once Product.java is migrated.
package model

// CartProductID is the composite primary key of a CartProduct, migrated from
// the JPA @EmbeddedId CartProductId embeddable. It pairs the owning cart's ID
// with the referenced product's ID.
type CartProductID struct {
	// CartID is the primary key of the associated Cart.
	CartID int `gorm:"column:cart_id" db:"cart_id"`
	// ProductID is the primary key of the associated Product.
	ProductID int `gorm:"column:product_id" db:"product_id"`
}

// NewCartProductID returns a CartProductID for the given cart and product IDs.
func NewCartProductID(cartID, productID int) CartProductID {
	return CartProductID{CartID: cartID, ProductID: productID}
}

// CartProduct is the Go representation of the CART_PRODUCT join entity. It
// models a single row linking a Cart to a Product in a many-to-many
// relationship with extra semantics.
//
// The composite primary key (ID) is derived from the associated Cart and
// Product, mirroring the JPA @MapsId behaviour. Use NewCartProduct to
// construct instances so the composite key stays consistent with the
// associations.
type CartProduct struct {
	// ID is the composite primary key (cart_id, product_id).
	ID CartProductID `gorm:"embedded"`

	// Cart is the associated cart. Its ID feeds ID.CartID.
	Cart *Cart `gorm:"foreignKey:CartID;references:ID" db:"-"`

	// Product is the associated product. Its ID feeds ID.ProductID.
	Product *Product `gorm:"foreignKey:ProductID;references:ID" db:"-"`
}

// TableName reports the physical table name backing CartProduct. This mirrors
// the JPA @Table(name = "CART_PRODUCT") mapping and is recognised by GORM.
func (CartProduct) TableName() string {
	return "CART_PRODUCT"
}

// NewCartProduct constructs a CartProduct linking the given cart and product,
// deriving the composite key from their IDs. This replaces the Java
// constructor CartProduct(Cart, Product), which set the associations and built
// a new CartProductId from cart.getId() and product.getId().
//
// MIGRATION_NOTE: The Java constructor dereferenced cart.getId() and
// product.getId() without a nil check, so a nil argument would have thrown a
// NullPointerException. To preserve that logic without panicking (per Go
// conventions), a nil cart or product yields a zero ID part rather than a
// panic; callers should ensure non-nil arguments.
func NewCartProduct(cart *Cart, product *Product) *CartProduct {
	var cartID, productID int
	if cart != nil {
		cartID = cart.GetID()
	}
	if product != nil {
		productID = product.GetID()
	}
	return &CartProduct{
		ID:      NewCartProductID(cartID, productID),
		Cart:    cart,
		Product: product,
	}
}

// GetID returns the composite primary key, migrating Java's getId().
func (cp *CartProduct) GetID() CartProductID {
	return cp.ID
}

// SetID sets the composite primary key, migrating Java's setId(CartProductId).
func (cp *CartProduct) SetID(id CartProductID) {
	cp.ID = id
}

// GetCart returns the associated cart, migrating Java's getCart().
func (cp *CartProduct) GetCart() *Cart {
	return cp.Cart
}

// SetCart sets the associated cart, migrating Java's setCart(Cart).
//
// MIGRATION_NOTE: JPA's @MapsId kept the composite key's cartId part in sync
// with the associated Cart's ID. Java's plain setter did not do this, but to
// preserve the invariant that the key reflects the associations, we update
// ID.CartID here. A nil cart leaves ID.CartID unchanged.
func (cp *CartProduct) SetCart(cart *Cart) {
	cp.Cart = cart
	if cart != nil {
		cp.ID.CartID = cart.GetID()
	}
}

// GetProduct returns the associated product, migrating Java's getProduct().
func (cp *CartProduct) GetProduct() *Product {
	return cp.Product
}

// SetProduct sets the associated product, migrating Java's setProduct(Product).
//
// MIGRATION_NOTE: As with SetCart, we keep the composite key's productId part
// in sync with the associated Product's ID to honour the @MapsId invariant. A
// nil product leaves ID.ProductID unchanged.
func (cp *CartProduct) SetProduct(product *Product) {
	cp.Product = product
	if product != nil {
		cp.ID.ProductID = product.GetID()
	}
}
