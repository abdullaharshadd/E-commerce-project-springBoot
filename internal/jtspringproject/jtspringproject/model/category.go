// Package model contains the Go migration of the JtSpringProject JPA entity
// classes (the com.jtspringproject.JtSpringProject.models package).
//
// This file migrates Category.java, a JPA @Entity mapped to the CATEGORY
// table. In the Java world the class carried the following persistence
// metadata:
//
//   - @Entity(name = "CATEGORY")             — mapped the class to the CATEGORY
//     table.
//   - @Id + @Column(name = "category_id") + @GeneratedValue(strategy = AUTO) —
//     auto-generated primary key mapped to the category_id column.
//   - name                                   — a plain, unannotated column.
//     JPA defaults the column name to the field name ("name").
//
// MIGRATION_NOTE: Go has no JPA/Hibernate ORM in the standard library, so the
// annotations do not translate directly. We model Category as a plain struct
// and carry the ORM mapping information as struct tags (using the widely used
// `gorm` and `db` conventions) so a repository layer can wire it to a table.
// The Java field was an `int`; Go models it as an `int` as well. The Java
// JavaBean getters/setters are preserved as idiomatic Go methods so callers
// that depend on accessor semantics keep working, but direct field access is
// perfectly acceptable in Go.
package model

// Category is the Go representation of the JtSpringProject CATEGORY entity.
//
// It models a product category with an auto-generated identifier and a name.
// The struct tags describe the intended relational mapping:
//
//   - ID   -> category_id column, primary key, auto-incremented.
//   - Name -> name column.
type Category struct {
	// ID is the auto-generated primary key (Java: @Id category_id).
	ID int `gorm:"column:category_id;primaryKey;autoIncrement" db:"category_id" json:"id"`

	// Name is the human-readable category name (Java: name).
	Name string `gorm:"column:name" db:"name" json:"name"`
}

// TableName reports the relational table this entity maps to.
//
// MIGRATION_NOTE: this mirrors the JPA @Entity(name = "CATEGORY") mapping and
// follows the gorm.io Tabler convention, so gorm-based repositories pick up the
// correct table name automatically.
func (Category) TableName() string {
	return "CATEGORY"
}

// NewCategory constructs a Category with the given name.
//
// The ID is intentionally left as its zero value: like the Java entity's
// @GeneratedValue(strategy = AUTO) field, the identifier is expected to be
// assigned by the persistence layer when the record is inserted.
func NewCategory(name string) *Category {
	return &Category{Name: name}
}

// GetID returns the category identifier (Java: getId).
func (c *Category) GetID() int {
	return c.ID
}

// SetID sets the category identifier (Java: setId).
func (c *Category) SetID(id int) {
	c.ID = id
}

// GetName returns the category name (Java: getName).
func (c *Category) GetName() string {
	return c.Name
}

// SetName sets the category name (Java: setName).
func (c *Category) SetName(name string) {
	c.Name = name
}
