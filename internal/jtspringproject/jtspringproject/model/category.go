// Package model contains the migrated domain entities for the original
// Spring/JPA application.
//
// This file migrates Category.java. The original class was a JPA @Entity mapped
// to the "CATEGORY" table with an auto-generated integer primary key
// ("category_id" column) and a free-text name.
//
// MIGRATION_NOTE: Go has no ORM annotations or JPA. The idiomatic replacement is
// a plain struct describing the row shape. Persistence concerns (table name,
// column names, ID generation) are handled explicitly by the
// repository/data-access layer using database/sql rather than by
// reflection-driven annotations.
//
//   - JPA @Id + @GeneratedValue(AUTO) -> ID is populated by the database on
//     insert; the repository reads it back (e.g. via LastInsertId or RETURNING).
//   - JavaBean getters/setters are dropped: exported struct fields are the
//     idiomatic Go accessor mechanism.
package model

// Category represents a product category, migrated from the JPA "CATEGORY"
// entity. The zero value is a valid, empty Category.
type Category struct {
	// ID is the auto-generated primary key, mapped to the "category_id" column.
	// It is zero for a Category that has not yet been persisted.
	ID int

	// Name is the human-readable category name.
	Name string
}

// NewCategory constructs a Category with the given name. The ID is left as the
// zero value and is expected to be assigned by the persistence layer on insert.
func NewCategory(name string) *Category {
	return &Category{Name: name}
}
