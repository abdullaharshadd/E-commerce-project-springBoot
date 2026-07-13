// Package resources contains the Go migration of the JtSpringProject
// externalized configuration that lived in
// src/main/resources/application.properties.
//
// In the Spring Boot world this .properties file was read either by Spring's
// auto-configuration (the spring.* keys) or by a manual @Configuration class
// that pulled the custom-prefixed keys (db.*, hibernate.*, entitymanager.*)
// out of the Environment via @Value. Go has no equivalent auto-wiring, so this
// file provides an explicit, strongly-typed configuration struct together with
// a constructor that mirrors the original property defaults and supports
// environment-variable overrides.
//
// MIGRATION_NOTE: The original file mixed active and commented-out keys. Only
// the ACTIVE keys are preserved as functional defaults below. The commented-out
// alternatives (the spring.datasource.* block and the top database.* block) are
// documented in comments for reference but are intentionally not wired up,
// exactly matching the original runtime behaviour.
//
// MIGRATION_NOTE: Spring-specific concepts do not translate to Go and require
// manual review:
//   - spring.mvc.view.prefix/suffix drove the JSP InternalResourceViewResolver.
//     There is no JSP in Go; whichever Go template/rendering layer replaces the
//     views must be configured to use these prefix/suffix values.
//   - hibernate.* and entitymanager.packagesToScan configured Hibernate/JPA,
//     which has no Go equivalent. If an ORM (e.g. GORM) is adopted, schema
//     auto-migration (hbm2ddl.auto=update) and dialect selection must be set up
//     manually there.
package resources

import (
	"fmt"
	"os"
	"strconv"
)

// Default configuration values transcribed verbatim from the ACTIVE keys in the
// original application.properties file.
const (
	// DefaultServerPort mirrors server.port=8080.
	DefaultServerPort = 8080

	// DefaultViewPrefix mirrors spring.mvc.view.prefix=/views/.
	DefaultViewPrefix = "/views/"
	// DefaultViewSuffix mirrors spring.mvc.view.suffix=.jsp.
	DefaultViewSuffix = ".jsp"

	// DefaultHibernateDialect mirrors hibernate.dialect=org.hibernate.dialect.MySQL5Dialect.
	DefaultHibernateDialect = "org.hibernate.dialect.MySQL5Dialect"
	// DefaultHibernateShowSQL mirrors hibernate.show_sql=true.
	DefaultHibernateShowSQL = true
	// DefaultHibernateDDLAuto mirrors hibernate.hbm2ddl.auto=update.
	DefaultHibernateDDLAuto = "update"
	// DefaultEnableLazyLoadNoTrans mirrors
	// spring.jpa.properties.hibernate.enable_lazy_load_no_trans=true.
	DefaultEnableLazyLoadNoTrans = true

	// DefaultDBDriver mirrors db.driver=com.mysql.cj.jdbc.Driver.
	DefaultDBDriver = "com.mysql.cj.jdbc.Driver"
	// DefaultDBURL mirrors
	// db.url=jdbc:mysql://localhost:3306/ecommjava?createDatabaseIfNotExist=true.
	DefaultDBURL = "jdbc:mysql://localhost:3306/ecommjava?createDatabaseIfNotExist=true"
	// DefaultDBUsername mirrors db.username=root.
	DefaultDBUsername = "root"
	// DefaultDBPassword mirrors db.password= (empty).
	DefaultDBPassword = ""
	// DefaultEntityManagerPackagesToScan mirrors entitymanager.packagesToScan=com.
	DefaultEntityManagerPackagesToScan = "com"
)

// ServerConfig holds HTTP server settings (the server.* keys).
type ServerConfig struct {
	// Port is the TCP port the HTTP server listens on (server.port).
	Port int
}

// ViewConfig holds the server-side view resolver settings (spring.mvc.view.*).
//
// MIGRATION_NOTE: These originally configured the JSP
// InternalResourceViewResolver. Reuse Prefix/Suffix in whatever Go rendering
// layer replaces JSP.
type ViewConfig struct {
	// Prefix is prepended to a logical view name (spring.mvc.view.prefix).
	Prefix string
	// Suffix is appended to a logical view name (spring.mvc.view.suffix).
	Suffix string
}

// DatabaseConfig holds the custom db.* connection properties.
type DatabaseConfig struct {
	// Driver is the JDBC driver class name (db.driver). Preserved for parity;
	// a Go driver import (e.g. github.com/go-sql-driver/mysql) is what actually
	// matters at runtime.
	Driver string
	// URL is the JDBC connection URL (db.url).
	URL string
	// Username is the database user (db.username).
	Username string
	// Password is the database password (db.password).
	Password string
}

// HibernateConfig holds the custom hibernate.* / JPA properties.
//
// MIGRATION_NOTE: Go has no Hibernate. These are retained for documentation and
// for a potential ORM adapter; nothing consumes them automatically.
type HibernateConfig struct {
	// Dialect mirrors hibernate.dialect.
	Dialect string
	// ShowSQL mirrors hibernate.show_sql.
	ShowSQL bool
	// DDLAuto mirrors hibernate.hbm2ddl.auto.
	DDLAuto string
	// EnableLazyLoadNoTrans mirrors
	// spring.jpa.properties.hibernate.enable_lazy_load_no_trans.
	EnableLazyLoadNoTrans bool
	// PackagesToScan mirrors entitymanager.packagesToScan.
	PackagesToScan string
}

// Config aggregates every configuration group defined in the original
// application.properties file.
type Config struct {
	// Server holds HTTP server settings.
	Server ServerConfig
	// View holds the view resolver settings.
	View ViewConfig
	// Database holds the db.* connection settings.
	Database DatabaseConfig
	// Hibernate holds the hibernate.*/JPA settings.
	Hibernate HibernateConfig
}

// DefaultConfig returns a Config populated with the exact ACTIVE values from the
// original application.properties file, with no environment overrides applied.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: DefaultServerPort,
		},
		View: ViewConfig{
			Prefix: DefaultViewPrefix,
			Suffix: DefaultViewSuffix,
		},
		Database: DatabaseConfig{
			Driver:   DefaultDBDriver,
			URL:      DefaultDBURL,
			Username: DefaultDBUsername,
			Password: DefaultDBPassword,
		},
		Hibernate: HibernateConfig{
			Dialect:               DefaultHibernateDialect,
			ShowSQL:               DefaultHibernateShowSQL,
			DDLAuto:               DefaultHibernateDDLAuto,
			EnableLazyLoadNoTrans: DefaultEnableLazyLoadNoTrans,
			PackagesToScan:        DefaultEntityManagerPackagesToScan,
		},
	}
}

// LoadConfig builds a Config starting from the defaults and applying any
// environment-variable overrides. It replaces Spring's Environment/@Value
// externalized-configuration mechanism with an explicit, testable loader.
//
// Recognized environment variables (all optional):
//
//	SERVER_PORT              -> Server.Port
//	VIEW_PREFIX             -> View.Prefix
//	VIEW_SUFFIX             -> View.Suffix
//	DB_DRIVER              -> Database.Driver
//	DB_URL                -> Database.URL
//	DB_USERNAME           -> Database.Username
//	DB_PASSWORD           -> Database.Password
//	HIBERNATE_DIALECT     -> Hibernate.Dialect
//	HIBERNATE_SHOW_SQL    -> Hibernate.ShowSQL
//	HIBERNATE_DDL_AUTO    -> Hibernate.DDLAuto
//	HIBERNATE_LAZY_LOAD_NO_TRANS -> Hibernate.EnableLazyLoadNoTrans
//	ENTITYMANAGER_PACKAGES_TO_SCAN -> Hibernate.PackagesToScan
//
// It returns an error when a provided override cannot be parsed (e.g. a
// non-numeric SERVER_PORT), rather than silently falling back to a default.
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	if v, ok := os.LookupEnv("SERVER_PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("parse SERVER_PORT %q: %w", v, err)
		}
		cfg.Server.Port = port
	}

	if v, ok := os.LookupEnv("VIEW_PREFIX"); ok {
		cfg.View.Prefix = v
	}
	if v, ok := os.LookupEnv("VIEW_SUFFIX"); ok {
		cfg.View.Suffix = v
	}

	if v, ok := os.LookupEnv("DB_DRIVER"); ok {
		cfg.Database.Driver = v
	}
	if v, ok := os.LookupEnv("DB_URL"); ok {
		cfg.Database.URL = v
	}
	if v, ok := os.LookupEnv("DB_USERNAME"); ok {
		cfg.Database.Username = v
	}
	if v, ok := os.LookupEnv("DB_PASSWORD"); ok {
		cfg.Database.Password = v
	}

	if v, ok := os.LookupEnv("HIBERNATE_DIALECT"); ok {
		cfg.Hibernate.Dialect = v
	}
	if v, ok := os.LookupEnv("HIBERNATE_SHOW_SQL"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("parse HIBERNATE_SHOW_SQL %q: %w", v, err)
		}
		cfg.Hibernate.ShowSQL = b
	}
	if v, ok := os.LookupEnv("HIBERNATE_DDL_AUTO"); ok {
		cfg.Hibernate.DDLAuto = v
	}
	if v, ok := os.LookupEnv("HIBERNATE_LAZY_LOAD_NO_TRANS"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("parse HIBERNATE_LAZY_LOAD_NO_TRANS %q: %w", v, err)
		}
		cfg.Hibernate.EnableLazyLoadNoTrans = b
	}
	if v, ok := os.LookupEnv("ENTITYMANAGER_PACKAGES_TO_SCAN"); ok {
		cfg.Hibernate.PackagesToScan = v
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate performs basic sanity checks on the resolved configuration.
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port %d: must be between 1 and 65535", c.Server.Port)
	}
	if c.Database.URL == "" {
		return fmt.Errorf("database URL must not be empty")
	}
	return nil
}

// Addr returns the server listen address in ":port" form, suitable for passing
// to http.Server.Addr or net.Listen.
func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}
