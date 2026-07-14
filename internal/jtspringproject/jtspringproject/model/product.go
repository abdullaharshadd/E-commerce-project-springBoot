// Package model contains the migrated domain entities for the original
// Spring/JPA application.
//
// This file migrates Product.java. The original class was a JPA @Entity mapped
// to the "PRODUCT" table with an auto-generated integer primary key
// ("product_id" column), scalar attributes (name, image, quantity, price,
// weight, description), a @OneToOne association to Category (foreign key
// "category_id", cascade ALL) and a @ManyToOne association to a User customer
// (foreign key "customer_id").
//
// MIGRATION_NOTE: Go has no ORM, JPA annotations, cascade semantics or lazy
// loading. The idiomatic replacement is a plain struct describing the row
// shape. Persistence concerns (table/column names, ID generation, cascades and
// association loading) are handled explicitly by the repository/data-access
// layer using database/sql rather than by reflection-driven annotations.
//
//   - JPA @Id + @GeneratedValue(AUTO) -> ID is populated by the database on
//     insert; the repository reads it back (e.g. via LastInsertId or RETURNING).
//   - @OneToOne(cascade = ALL) Category -> modelled as an embedded *Category
//     pointer. Cascading persistence (saving/deleting the Category alongside
//     the Product) must be done explicitly in the repository within a
//     transaction (see WithinTransaction). The raw foreign-key value is kept in
//     CategoryID for cases where the full association is not loaded.
//   - @ManyToOne User customer -> modelled as a *User pointer plus a raw
//     CustomerID foreign key. The User type is not yet migrated, so the
//     association is represented by its foreign key only for now.
//   - JavaBean getters/setters are dropped in favour of exported struct fields,
//     which is idiomatic Go.
package model

// Product is the migrated form of the original JPA Product entity. It describes
// a single row of the "PRODUCT" table in the e-commerce domain.
//
// Fields map to the original columns as follows:
//
//   - ID          -> "product_id" (auto-generated primary key)
//   - Name        -> "name"
//   - Image       -> "image"
//   - Category    -> @OneToOne association ("category_id")
//   - Quantity    -> "quantity"
//   - Price       -> "price"
//   - Weight      -> "weight"
//   - Description -> "description"
//   - Customer    -> @ManyToOne association ("customer_id")
type Product struct {
	// ID is the auto-generated primary key ("product_id"). A zero value means
	// the product has not yet been persisted.
	ID int

	// Name is the free-text product name.
	Name string

	// Image holds the product image reference (path/URL).
	Image string

	// Category is the associated Category (original @OneToOne, cascade ALL).
	// It may be nil when the association has not been loaded.
	Category *Category

	// CategoryID is the raw "category_id" foreign key. It is kept so callers
	// can reference the association without loading the full Category.
	CategoryID int

	// Quantity is the available stock count.
	Quantity int

	// Price is the product price. MIGRATION_NOTE: the source used a plain int
	// for currency; this is preserved exactly but should be reviewed, as
	// monetary values are usually better represented as integer minor units
	// (cents) or a dedicated decimal type.
	Price int

	// Weight is the product weight.
	Weight int

	// Description is the free-text product description.
	Description string

	// CustomerID is the raw "customer_id" foreign key for the original
	// @ManyToOne User association.
	//
	// MIGRATION_NOTE: the original entity referenced a User type via a
	// *User pointer. The User model has not been migrated yet, so only the
	// foreign-key value is represented here. Once User is migrated, add a
	// *User field mirroring the Category association and populate it in the
	// repository layer.
	CustomerID int
}

// NewProduct constructs a Product with the given attributes. The ID is left as
// its zero value and is expected to be populated by the persistence layer on
// insert. The category association may be nil; when non-nil its ID is used to
// populate the CategoryID foreign key for convenience.
func NewProduct(name, image string, category *Category, quantity, price, weight int, description string, customerID int) *Product {
	p := &Product{
		Name:        name,
		Image:       image,
		Category:    category,
		Quantity:    quantity,
		Price:       price,
		Weight:      weight,
		Description: description,
		CustomerID:  customerID,
	}
	if category != nil {
		p.CategoryID = category.ID
	}
	return p
}
