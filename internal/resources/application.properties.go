// Package resources contains the migrated application configuration for the
// original Spring Boot application.
//
// This file migrates src/main/resources/application.properties. The original
// file was a Spring Boot properties file that Spring's environment abstraction
// loaded automatically at startup, driving auto-configuration of the embedded
// servlet container, the MVC view resolver, the JDBC DataSource and the
// Hibernate/JPA EntityManagerFactory.
//
// MIGRATION_NOTE: Go has no equivalent of Spring's property-driven
// auto-configuration. There is no framework that will read a .properties file
// and wire beans by reflection. The idiomatic replacement is an explicit,
// strongly-typed configuration struct populated from environment variables
// (with sane defaults) and passed to the constructors that build each
// component. This makes the previously implicit magic explicit and testable.
//
// Property mapping:
//
//   - server.port                          -> AppConfig.ServerPort
//   - spring.mvc.view.prefix / .suffix     -> ViewConfig (see MIGRATION_NOTE below)
//   - hibernate.dialect                    -> not applicable in Go (see below)
//   - hibernate.show_sql                   -> AppConfig.ShowSQL (logging toggle)
//   - hibernate.hbm2ddl.auto=update        -> AppConfig.AutoMigrate (see below)
//   - enable_lazy_load_no_trans            -> not applicable (no ORM lazy loading)
//   - db.driver / db.url / db.username /   -> reuse Config in
//     db.password                            hibernateconfiguration.go
//
// The active datasource in the original file used the custom db.* keys (the
// spring.datasource.* and commented database.* blocks were disabled), so only
// the db.* values are migrated here.
package resources

import (
	"fmt"
	"os"
	"strconv"
)

const (
	// DefaultServerPort mirrors the original server.port=8080 setting.
	DefaultServerPort = 8080

	// DefaultViewPrefix mirrors spring.mvc.view.prefix=/views/.
	DefaultViewPrefix = "/views/"

	// DefaultViewSuffix mirrors spring.mvc.view.suffix=.jsp.
	DefaultViewSuffix = ".jsp"

	// DefaultShowSQL mirrors hibernate.show_sql=true. In Go this is a hint to
	// the data-access layer to log the SQL it executes.
	DefaultShowSQL = true

	// DefaultAutoMigrate mirrors hibernate.hbm2ddl.auto=update.
	//
	// MIGRATION_NOTE: Hibernate's hbm2ddl.auto=update automatically created and
	// altered tables to match the entity mappings at startup. Go's database/sql
	// has no schema management. The idiomatic replacement is explicit,
	// version-controlled migrations (e.g. golang-migrate, goose, or hand-written
	// SQL run at startup). This flag only signals intent; a real migration
	// runner must be wired up by the data-access layer. REQUIRES MANUAL REVIEW.
	DefaultAutoMigrate = true
)

// ViewConfig captures the original Spring MVC view resolver settings
// (spring.mvc.view.prefix / .suffix).
//
// MIGRATION_NOTE: The original application used InternalResourceViewResolver to
// render server-side JSP views (e.g. a logical view name "index" resolved to
// "/views/index.jsp"). Go has no JSP engine. The idiomatic replacement is
// html/template with a template directory and file extension. Prefix/Suffix are
// preserved here so handlers can reproduce the same logical-name-to-file
// resolution, but the templates themselves must be ported from JSP by hand.
// REQUIRES MANUAL REVIEW.
type ViewConfig struct {
	// Prefix is the directory prepended to a logical view name.
	Prefix string
	// Suffix is the file extension appended to a logical view name.
	Suffix string
}

// Resolve maps a logical view name to its full template path, reproducing the
// behaviour of Spring's InternalResourceViewResolver.
func (v ViewConfig) Resolve(name string) string {
	return v.Prefix + name + v.Suffix
}

// AppConfig is the strongly-typed replacement for application.properties. It is
// populated from the environment by LoadAppConfig and passed explicitly to the
// components that need it.
type AppConfig struct {
	// ServerPort is the TCP port the HTTP server listens on.
	ServerPort int
	// View holds the server-side view-rendering settings.
	View ViewConfig
	// ShowSQL toggles SQL statement logging in the data-access layer.
	ShowSQL bool
	// AutoMigrate signals that schema migrations should run at startup.
	AutoMigrate bool
	// EntityPackage mirrors entitymanager.packagesToScan. It has no runtime
	// meaning in Go and is retained only for documentation/traceability.
	EntityPackage string
}

// Addr returns the listen address suitable for http.Server, e.g. ":8080".
func (c AppConfig) Addr() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

// LoadAppConfig builds an AppConfig from environment variables, falling back to
// the defaults that were hard-coded in the original application.properties.
//
// It returns an error only when a provided value is malformed (for example a
// non-numeric SERVER_PORT), so that misconfiguration fails fast at startup
// rather than silently using a default.
//
// Database connection settings (db.driver / db.url / db.username / db.password)
// are intentionally not handled here: they are owned by Config /
// LoadConfigFromEnv in
// internal/jtspringproject/jtspringproject/hibernateconfiguration.go, which is
// the single source of truth for the datasource.
func LoadAppConfig() (AppConfig, error) {
	cfg := AppConfig{
		ServerPort: DefaultServerPort,
		View: ViewConfig{
			Prefix: DefaultViewPrefix,
			Suffix: DefaultViewSuffix,
		},
		ShowSQL:       DefaultShowSQL,
		AutoMigrate:   DefaultAutoMigrate,
		EntityPackage: "com",
	}

	if v, ok := os.LookupEnv("SERVER_PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return AppConfig{}, fmt.Errorf("invalid SERVER_PORT %q: %w", v, err)
		}
		if port < 1 || port > 65535 {
			return AppConfig{}, fmt.Errorf("SERVER_PORT %d out of range 1-65535", port)
		}
		cfg.ServerPort = port
	}

	if v, ok := os.LookupEnv("VIEW_PREFIX"); ok {
		cfg.View.Prefix = v
	}
	if v, ok := os.LookupEnv("VIEW_SUFFIX"); ok {
		cfg.View.Suffix = v
	}

	if v, ok := os.LookupEnv("SHOW_SQL"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return AppConfig{}, fmt.Errorf("invalid SHOW_SQL %q: %w", v, err)
		}
		cfg.ShowSQL = b
	}

	if v, ok := os.LookupEnv("AUTO_MIGRATE"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return AppConfig{}, fmt.Errorf("invalid AUTO_MIGRATE %q: %w", v, err)
		}
		cfg.AutoMigrate = b
	}

	if v, ok := os.LookupEnv("ENTITY_PACKAGE"); ok {
		cfg.EntityPackage = v
	}

	return cfg, nil
}
