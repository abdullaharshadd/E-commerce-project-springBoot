// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates Product.java, a JPA @Entity mapped to the PRODUCT table.
// In the Java world the class carried the following persistence metadata:
//
//   - @Entity(name = "PRODUCT")                       — mapped the class to the
//     PRODUCT table.
//   - @Id + @Column(name = "product_id") + @GeneratedValue(strategy = AUTO) —
//     auto-generated primary key mapped to the product_id column.
//   - name, image, quantity, price, weight, description — plain, unannotated
//     columns. JPA defaults the column name to the field name.
//   - @OneToOne(cascade = ALL) + @JoinColumn("category_id") Category category —
//     one-to-one association to Category with cascading persistence, joined on
//     the category_id foreign key.
//   - @ManyToOne + @JoinColumn("customer_id") User customer — many-to-one
//     association to the customer User, joined on the customer_id foreign key.
//
// MIGRATION_NOTE: Go has no JPA/Hibernate ORM in the standard library, so the
// annotations do not translate directly. The struct below models the plain
// data; persistence (table mapping, key generation, cascade behaviour, and FK
// joins) must be handled explicitly by the chosen data-access layer
// (e.g. sqlx/gorm) in the repository code.
//
// MIGRATION_NOTE: The Java 'User customer' field references a User entity that
// has not yet been migrated. It is modelled here as an *interface{} placeholder
// via the Customer field typed as *User once that entity is available. Until
// then the association is stored as an opaque reference. Human review required
// to wire this to the migrated User type.
package model

// Product is the Go migration of the Product JPA entity. It represents a
// product in the e-commerce domain, previously mapped to the PRODUCT table.
//
// Fields are unexported to preserve the JavaBean encapsulation of the original
// class; use the constructor and accessor methods to interact with a Product.
type Product struct {
	// id is the auto-generated primary key (product_id column).
	id int
	// name is the product name.
	name string
	// image holds the product image reference.
	image string
	// category is the associated Category (OneToOne, category_id FK).
	category *Category
	// quantity is the available stock quantity.
	quantity int
	// price is the product price.
	price int
	// weight is the product weight.
	weight int
	// description is the product description.
	description string
	// customer is the associated customer.
	//
	// MIGRATION_NOTE: In Java this was a @ManyToOne association to the User
	// entity (customer_id FK). The User type has not been migrated yet, so this
	// is left as an interface{} placeholder. Replace with *User (or the
	// migrated customer type) once available.
	customer interface{}
}

// NewProduct constructs an empty Product, mirroring the default JPA entity
// constructor. Fields should be populated via the setter methods.
func NewProduct() *Product {
	return &Product{}
}

// GetID returns the product's primary key.
func (p *Product) GetID() int {
	return p.id
}

// SetID sets the product's primary key.
func (p *Product) SetID(id int) {
	p.id = id
}

// GetName returns the product name.
func (p *Product) GetName() string {
	return p.name
}

// SetName sets the product name.
func (p *Product) SetName(name string) {
	p.name = name
}

// GetImage returns the product image reference.
func (p *Product) GetImage() string {
	return p.image
}

// SetImage sets the product image reference.
func (p *Product) SetImage(image string) {
	p.image = image
}

// GetCategory returns the associated Category, or nil if unset.
func (p *Product) GetCategory() *Category {
	return p.category
}

// SetCategory sets the associated Category.
func (p *Product) SetCategory(category *Category) {
	p.category = category
}

// GetQuantity returns the available stock quantity.
func (p *Product) GetQuantity() int {
	return p.quantity
}

// SetQuantity sets the available stock quantity.
func (p *Product) SetQuantity(quantity int) {
	p.quantity = quantity
}

// GetPrice returns the product price.
func (p *Product) GetPrice() int {
	return p.price
}

// SetPrice sets the product price.
func (p *Product) SetPrice(price int) {
	p.price = price
}

// GetWeight returns the product weight.
func (p *Product) GetWeight() int {
	return p.weight
}

// SetWeight sets the product weight.
func (p *Product) SetWeight(weight int) {
	p.weight = weight
}

// GetDescription returns the product description.
func (p *Product) GetDescription() string {
	return p.description
}

// SetDescription sets the product description.
func (p *Product) SetDescription(description string) {
	p.description = description
}

// GetCustomer returns the associated customer.
//
// MIGRATION_NOTE: Returns interface{} until the User entity is migrated.
// Replace with the concrete *User return type when available.
func (p *Product) GetCustomer() interface{} {
	return p.customer
}

// SetCustomer sets the associated customer.
//
// MIGRATION_NOTE: Accepts interface{} until the User entity is migrated.
// Replace with the concrete *User parameter type when available.
func (p *Product) SetCustomer(customer interface{}) {
	p.customer = customer
}
