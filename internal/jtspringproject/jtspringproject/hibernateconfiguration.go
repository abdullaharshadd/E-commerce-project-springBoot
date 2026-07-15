// Package jtspringproject contains the migrated configuration and wiring for
// the original Spring application.
//
// This file migrates HibernateConfiguration.java. The original class was a
// Spring @Configuration that declared three beans:
//
//   - DataSource               (JDBC connection pool via DriverManagerDataSource)
//   - LocalSessionFactoryBean  (Hibernate ORM SessionFactory)
//   - HibernateTransactionManager (declarative, AOP-proxied transactions)
//
// MIGRATION_NOTE: Go has no direct equivalent of Hibernate/JPA. The idiomatic
// replacement is the standard library's database/sql package (optionally with a
// thin helper like sqlx or a query builder). Consequently:
//
//   - The Hibernate SessionFactory has NO Go analogue. In Go you open a
//     *sql.DB (which is itself a pooled handle) and pass it to repositories.
//     Entity mapping, dialect selection, show_sql and hbm2ddl.auto (schema
//     auto-generation) are all Hibernate-specific and are dropped. Schema
//     management should be handled explicitly with a migration tool such as
//     golang-migrate or goose. See the field docs on Config below.
//
//   - @EnableTransactionManagement (AOP proxy based declarative transactions)
//     has no equivalent. In Go, transactions are explicit: call db.BeginTx and
//     pass the *sql.Tx down the call stack. This is a deliberate, idiomatic
//     improvement over hidden proxy magic and requires manual review of every
//     original @Transactional method.
//
//   - @Value externalized property injection is replaced by an explicit Config
//     struct populated from environment variables (or a config file). No hidden
//     container wiring.
//
// This file registers NO HTTP routes: the source declared none.
package jtspringproject

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

// Config holds the externalized database settings that were previously injected
// into HibernateConfiguration via Spring @Value annotations.
//
// The Hibernate-specific fields (Dialect, ShowSQL, HBM2DDLAuto, PackagesToScan)
// are retained only for documentation/parity; they have no runtime effect in the
// Go implementation. See the MIGRATION_NOTE on each field.
type Config struct {
	// Driver is the database/sql driver name (e.g. "mysql", "postgres").
	// Source: ${db.driver}. Note: the original held a Java driver *class name*
	// (e.g. com.mysql.cj.jdbc.Driver). It must be mapped to a Go driver name and
	// the corresponding driver package must be imported (blank import) by the
	// composition root.
	Driver string
	// DSN is the data source name / connection string passed to sql.Open.
	// Source: ${db.url} combined with ${db.username} and ${db.password}. JDBC
	// URLs are NOT valid Go DSNs and must be translated for the chosen driver.
	DSN string
	// URL is the raw source ${db.url}, kept for reference/DSN construction.
	URL string
	// Username is the source ${db.username}.
	Username string
	// Password is the source ${db.password}.
	Password string

	// Dialect mirrors ${hibernate.dialect}. MIGRATION_NOTE: no effect in Go;
	// dialect is implied by the chosen driver.
	Dialect string
	// ShowSQL mirrors ${hibernate.show_sql}. MIGRATION_NOTE: no effect in Go;
	// use a logging middleware or driver-level logging if SQL tracing is needed.
	ShowSQL string
	// HBM2DDLAuto mirrors ${hibernate.hbm2ddl.auto}. MIGRATION_NOTE: no effect
	// in Go; schema auto-management is intentionally not replicated. Use an
	// explicit migration tool (golang-migrate, goose).
	HBM2DDLAuto string
	// PackagesToScan mirrors ${entitymanager.packagesToScan}. MIGRATION_NOTE:
	// no effect in Go; there is no automatic entity discovery. Repositories map
	// rows to structs explicitly.
	PackagesToScan string
}

// LoadConfigFromEnv builds a Config from environment variables, replacing the
// Spring externalized @Value injection.
//
// It returns an error if any required setting is missing so that misconfiguration
// fails fast at startup rather than at first query.
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		Driver:         os.Getenv("DB_DRIVER"),
		URL:            os.Getenv("DB_URL"),
		Username:       os.Getenv("DB_USERNAME"),
		Password:       os.Getenv("DB_PASSWORD"),
		DSN:            os.Getenv("DB_DSN"),
		Dialect:        os.Getenv("HIBERNATE_DIALECT"),
		ShowSQL:        os.Getenv("HIBERNATE_SHOW_SQL"),
		HBM2DDLAuto:    os.Getenv("HIBERNATE_HBM2DDL_AUTO"),
		PackagesToScan: os.Getenv("ENTITYMANAGER_PACKAGES_TO_SCAN"),
	}

	if cfg.Driver == "" {
		return Config{}, fmt.Errorf("load config: DB_DRIVER is required")
	}
	if cfg.DSN == "" {
		// MIGRATION_NOTE: the source built the connection from separate url/user/
		// password fields (a JDBC URL). A Go DSN is driver specific and cannot be
		// reliably synthesized here, so we require it explicitly.
		return Config{}, fmt.Errorf("load config: DB_DSN is required (translate the JDBC url/username/password into a driver-specific DSN)")
	}
	return cfg, nil
}

// NewDataSource opens a database handle configured from cfg. It replaces the
// dataSource() bean.
//
// The returned *sql.DB is safe for concurrent use and manages its own connection
// pool, so it is the idiomatic Go equivalent of both the Spring DataSource and
// the Hibernate SessionFactory. The caller owns the handle and must Close it
// during graceful shutdown.
//
// A ctx-bounded PingContext verifies connectivity eagerly, mirroring Spring's
// fail-fast bean initialization.
func NewDataSource(ctx context.Context, cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open datasource (driver=%q): %w", cfg.Driver, err)
	}

	// Reasonable pool defaults; tune per deployment. The source relied on
	// DriverManagerDataSource which pools nothing, so these are an improvement.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping datasource: %w", err)
	}
	return db, nil
}

// Transactor abstracts the ability to begin a database transaction. It is the
// explicit, idiomatic replacement for @EnableTransactionManagement /
// HibernateTransactionManager.
//
// MIGRATION_NOTE: Spring's declarative @Transactional applied transactions via
// AOP proxies invisibly. In Go the boundary is explicit and testable: callers
// use WithinTransaction to run a unit of work atomically.
type Transactor interface {
	// WithinTransaction runs fn inside a transaction, committing on success and
	// rolling back if fn returns an error or panics.
	WithinTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error
}

// SQLTransactor implements Transactor using a *sql.DB. It is the concrete
// replacement for HibernateTransactionManager.
type SQLTransactor struct {
	db *sql.DB
}

// NewSQLTransactor constructs a SQLTransactor backed by db.
func NewSQLTransactor(db *sql.DB) *SQLTransactor {
	return &SQLTransactor{db: db}
}

// WithinTransaction begins a transaction, invokes fn, and commits. If fn returns
// an error the transaction is rolled back and the original error is returned. A
// panic inside fn triggers a rollback before the panic is re-raised.
func (t *SQLTransactor) WithinTransaction(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err = fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback after error %v: %w", err, rbErr)
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
