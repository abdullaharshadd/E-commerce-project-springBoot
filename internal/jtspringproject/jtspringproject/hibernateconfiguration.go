// Package jtspringproject contains the Go migration of the JtSpringProject
// Spring Boot web application.
//
// This file migrates HibernateConfiguration.java, a Spring @Configuration class
// that wired up Hibernate ORM infrastructure. In the Spring/Java world this
// class created three IoC-container-managed singletons:
//
//   - DataSource                (JDBC connection factory)
//   - LocalSessionFactoryBean   (Hibernate SessionFactory)
//   - HibernateTransactionManager (declarative transaction manager)
//
// MIGRATION_NOTE: There is no idiomatic 1:1 translation of Hibernate + Spring
// declarative transactions in Go. Go has no ORM auto-configuration, no CGLIB
// bean proxying, and no annotation-driven transaction management
// (@EnableTransactionManagement). Instead we:
//
//   - Represent the externalized "${db.*}" / "${hibernate.*}" properties as a
//     plain Config struct populated from the environment (replacing @Value).
//   - Open a *sql.DB via database/sql, which fills the role of the JDBC
//     DataSource. (SessionFactory / hibernate.hbm2ddl.auto / packagesToScan /
//     dialect have no direct database/sql equivalent — see notes below.)
//   - Rely on explicit sql.Tx transactions at the call site rather than
//     AOP-based declarative transactions.
//
// REQUIRES MANUAL REVIEW: choose a concrete driver (e.g. github.com/go-sql-driver/mysql),
// decide on schema migration tooling (replacing hbm2ddl.auto), and adopt a
// query layer (sqlx / sqlc / gorm) to replace the Hibernate SessionFactory.
package jtspringproject

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

// HibernateConfig holds the externalized database and (former) Hibernate
// settings that were injected via Spring @Value annotations.
//
// The Dialect, ShowSQL, HBM2DDLAuto and PackagesToScan fields are retained for
// documentation and configuration parity with the original application, but
// they have no direct effect in the database/sql world — see the field docs.
type HibernateConfig struct {
	// Driver is the JDBC driver class name in the original config
	// ("${db.driver}"). In Go this maps to a registered database/sql driver
	// name (e.g. "mysql", "postgres").
	Driver string
	// URL is the database connection URL ("${db.url}").
	//
	// MIGRATION_NOTE: a JDBC URL is NOT a valid Go DSN. It must be translated
	// to the driver-specific DSN format for the chosen driver.
	URL string
	// Username is the database user ("${db.username}").
	Username string
	// Password is the database password ("${db.password}").
	Password string

	// Dialect corresponds to "${hibernate.dialect}". There is no equivalent
	// concept in database/sql; it is preserved for reference only.
	Dialect string
	// ShowSQL corresponds to "${hibernate.show_sql}". Query logging in Go is
	// implemented explicitly (e.g. via a wrapping driver or logger).
	ShowSQL string
	// HBM2DDLAuto corresponds to "${hibernate.hbm2ddl.auto}". Automatic schema
	// generation has no equivalent; use an explicit migration tool.
	HBM2DDLAuto string
	// PackagesToScan corresponds to "${entitymanager.packagesToScan}".
	// Entity/package scanning has no equivalent in Go; models are referenced
	// explicitly.
	PackagesToScan string
}

// LoadHibernateConfig builds a HibernateConfig from environment variables,
// replacing Spring's @Value/PropertySource injection.
//
// It returns an error if any of the required database connection properties
// (driver, url, username) are empty.
func LoadHibernateConfig() (*HibernateConfig, error) {
	cfg := &HibernateConfig{
		Driver:         os.Getenv("DB_DRIVER"),
		URL:            os.Getenv("DB_URL"),
		Username:       os.Getenv("DB_USERNAME"),
		Password:       os.Getenv("DB_PASSWORD"),
		Dialect:        os.Getenv("HIBERNATE_DIALECT"),
		ShowSQL:        os.Getenv("HIBERNATE_SHOW_SQL"),
		HBM2DDLAuto:    os.Getenv("HIBERNATE_HBM2DDL_AUTO"),
		PackagesToScan: os.Getenv("ENTITYMANAGER_PACKAGES_TO_SCAN"),
	}

	if cfg.Driver == "" {
		return nil, fmt.Errorf("hibernate config: DB_DRIVER is required")
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("hibernate config: DB_URL is required")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("hibernate config: DB_USERNAME is required")
	}

	return cfg, nil
}

// NewDataSource opens a database connection pool (*sql.DB) using the supplied
// configuration. This is the Go equivalent of the Spring dataSource() @Bean,
// which produced a DriverManagerDataSource.
//
// The provided context bounds the initial connectivity check (Ping). The
// returned *sql.DB is safe for concurrent use and should be closed by the
// caller when the application shuts down.
//
// MIGRATION_NOTE: cfg.URL holds a JDBC URL in the original config. The chosen
// database/sql driver almost certainly requires a different DSN format; the
// translation must be done by the caller or a dedicated helper before this is
// production-ready.
func NewDataSource(ctx context.Context, cfg *HibernateConfig) (*sql.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("new data source: config must not be nil")
	}

	// sql.Open validates arguments but does not establish a connection.
	db, err := sql.Open(cfg.Driver, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("new data source: open %q: %w", cfg.Driver, err)
	}

	// Verify connectivity, bounded by the caller's context.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		// Best-effort cleanup; the connection failed so we cannot use db.
		_ = db.Close()
		return nil, fmt.Errorf("new data source: ping: %w", err)
	}

	return db, nil
}

// BeginTx starts an explicit database transaction on the given data source.
//
// MIGRATION_NOTE: This replaces Spring's HibernateTransactionManager and
// @EnableTransactionManagement. Spring provided declarative, proxy-based
// (AOP) transaction demarcation around annotated methods. Go has no such
// mechanism: transactions must be started, committed and rolled back
// explicitly at the call site. Callers are responsible for calling Commit or
// Rollback on the returned *sql.Tx.
func BeginTx(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	if db == nil {
		return nil, fmt.Errorf("begin tx: data source must not be nil")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	return tx, nil
}
