// Package jtspringproject contains the migrated application entry point for
// the original Spring Boot application (JtSpringProjectApplication.java).
//
// MIGRATION_NOTE: The original class was annotated with
// @SpringBootApplication(exclude = HibernateJpaAutoConfiguration.class) and its
// sole job was to bootstrap the Spring application context and start the
// embedded servlet container via SpringApplication.run(...).
//
// In Go there is no auto-configuration / component scanning. The idiomatic
// equivalent is an explicit main() that:
//
//   1. Loads configuration (see LoadConfigFromEnv in hibernateconfiguration.go).
//   2. Opens the database connection (NewDataSource) and wires the transactor.
//   3. Builds the HTTP handler with all routes (buildHandler).
//   4. Starts an *http.Server and blocks, with graceful shutdown on SIGINT/SIGTERM.
//
// The exclusion of HibernateJpaAutoConfiguration is a no-op here: Go uses the
// standard database/sql package directly (see NewDataSource), so there is no
// ORM auto-configuration to exclude.
package jtspringproject

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// defaultAddr is the address the HTTP server listens on when none is provided
// via configuration. Spring Boot defaults to :8080, so we preserve that.
const defaultAddr = ":8080"

// buildHandler constructs the fully-wired HTTP handler for the application.
//
// It mirrors the route surface that Spring's component scanning would have
// discovered from the migrated controllers. Each route is registered as a
// functional stub returning HTTP 200 with its logical route name; wire the
// real handler logic in as the corresponding controller migrations land.
func buildHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Liveness/readiness probe.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Public / user-facing routes.
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: Index")
	})
	r.Get("/login", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: UserLogin")
	})
	r.Post("/login", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: UserLoginSubmit")
	})
	r.Get("/register", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: UserRegister")
	})
	r.Post("/newuserregister", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: NewUserRegister")
	})
	r.Get("/logout", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: UserLogout")
	})

	// Product browsing (customer-facing).
	r.Get("/index", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: CustomerIndex")
	})
	r.Get("/getcategory", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: GetCategory")
	})
	r.Get("/buy", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "route: Buy")
	})

	// Admin routes.
	r.Route("/admin", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminHome")
		})
		r.Get("/login", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminLogin")
		})
		r.Post("/login", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminLoginSubmit")
		})
		r.Get("/dashboard", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminDashboard")
		})
		r.Get("/logout", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminLogout")
		})

		// Category management.
		r.Get("/categories", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminCategories")
		})
		r.Post("/addcategory", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminAddCategory")
		})
		r.Get("/deletecategory", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminDeleteCategory")
		})
		r.Post("/updatecategory", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminUpdateCategory")
		})

		// Product management.
		r.Get("/products", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminProducts")
		})
		r.Get("/addproduct", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminAddProductForm")
		})
		r.Post("/addproduct", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminAddProduct")
		})
		r.Get("/deleteproduct", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminDeleteProduct")
		})
		r.Get("/updateproduct", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminUpdateProductForm")
		})
		r.Post("/updateproduct", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminUpdateProduct")
		})

		// User management.
		r.Get("/customers", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "route: AdminCustomers")
		})
	})

	return r
}

// Run bootstraps and starts the application. It is the idiomatic Go equivalent
// of SpringApplication.run(...): it wires dependencies, starts the HTTP server,
// and blocks until the process receives an interrupt/terminate signal, at which
// point it shuts the server down gracefully.
//
// Run returns an error if the server fails to start or does not shut down
// cleanly; callers (typically main) should log and exit non-zero on error.
func Run(ctx context.Context) error {
	// Load configuration. MIGRATION_NOTE: In Spring this was handled by
	// auto-configuration + application.properties; here it is explicit.
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Open the database connection pool. The returned *sql.DB replaces the
	// Spring DataSource bean. We keep it wired here even though the current
	// route stubs do not use it yet, so that controller migrations can depend
	// on it. It is deferred-closed on shutdown.
	db, err := NewDataSource(cfg)
	if err != nil {
		return fmt.Errorf("open datasource: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// The transactor replaces Spring's HibernateTransactionManager. It is
	// constructed here and will be injected into repositories/services as the
	// controller migrations land.
	_ = NewSQLTransactor(db)

	addr := cfg.ServerAddr
	if addr == "" {
		addr = defaultAddr
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           buildHandler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start the server in a goroutine so we can wait for either a fatal serve
	// error or a shutdown signal.
	serveErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	// Wait for a termination signal or a serve error.
	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	case <-sigCtx.Done():
		// Graceful shutdown with a bounded timeout.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	}
}

// main is the process entry point, equivalent to Spring's
// JtSpringProjectApplication.main(String[]). It delegates to Run and maps any
// error onto a non-zero exit code.
//
// MIGRATION_NOTE: The original main lived in the same class as the application
// bootstrap. In a larger Go project this main would typically live under
// cmd/<app>/main.go; it is kept here to mirror the source file's structure.
func main() {
	if err := Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
