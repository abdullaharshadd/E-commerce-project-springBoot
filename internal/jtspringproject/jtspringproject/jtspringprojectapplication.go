// Package jtspringproject contains the Go migration of the JtSpringProject
// Spring Boot web application.
//
// This file migrates JtSpringProjectApplication.java, the Spring Boot entry
// point. In the Java world the class was annotated with:
//
//	@SpringBootApplication(exclude = HibernateJpaAutoConfiguration.class)
//
// which does three things at once:
//
//   - @Configuration          — marks the class as a bean definition source
//   - @EnableAutoConfiguration — activates Spring Boot's classpath-driven
//     auto-configuration (and here *excludes* HibernateJpaAutoConfiguration so
//     that Spring does not try to set up JPA/Hibernate automatically; the app
//     wires Hibernate manually in HibernateConfiguration.java)
//   - @ComponentScan           — scans this package (and children) for beans
//
// SpringApplication.run(...) then bootstraps the IoC container, performs
// dependency injection, and starts the embedded servlet container (Tomcat).
//
// MIGRATION_NOTE: Go has no IoC container, no classpath auto-configuration, and
// no component scanning. The idiomatic equivalent of "bootstrap the app and
// start the server" is an explicit main() (or a Run function) that:
//
//  1. loads configuration,
//  2. constructs dependencies explicitly (wiring, not scanning),
//  3. builds the HTTP server / router,
//  4. starts serving and blocks until shutdown,
//  5. shuts down gracefully.
//
// Because the real HTTP handlers, router, and repositories live in other source
// files that are migrated separately, this file provides the bootstrap
// scaffolding (Run + graceful shutdown) and wires in the already-migrated
// Hibernate/data-source layer. The dependency wiring marked below
// REQUIRES MANUAL REVIEW once the controller/service files are migrated.
//
// The HibernateJpaAutoConfiguration *exclusion* has no Go analogue: there is no
// auto-configuration to exclude. Manual data-source setup via NewDataSource /
// LoadHibernateConfig is the migrated behavior.
package jtspringproject

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// defaultServerAddr mirrors Spring Boot's default embedded-server port (8080).
// MIGRATION_NOTE: In Spring this came from application.properties/convention;
// here it is an explicit constant that should be made configurable.
const defaultServerAddr = ":8080"

// defaultShutdownTimeout bounds how long graceful shutdown waits for in-flight
// requests to complete before forcibly closing the server.
const defaultShutdownTimeout = 15 * time.Second

// Application is the top-level composition root for the migrated service. It
// replaces the implicit Spring ApplicationContext: instead of scanning for
// beans, dependencies are constructed explicitly and held here.
//
// MIGRATION_NOTE: The zero value is not usable; construct via NewApplication.
type Application struct {
	server *http.Server
}

// NewApplication constructs the application composition root.
//
// It performs the explicit dependency wiring that Spring's component scanning
// and auto-configuration did implicitly:
//
//   - loads the Hibernate/data-source configuration (previously
//     HibernateConfiguration.java + the excluded auto-configuration),
//   - builds the HTTP handler chain.
//
// The addr parameter is the listen address (e.g. ":8080"); pass "" to use the
// convention default that mirrors Spring Boot.
//
// REQUIRES MANUAL REVIEW: The controllers, services, and repositories from the
// rest of the Spring project must be constructed and mounted onto the returned
// router below. Until those files are migrated, the handler is a minimal
// placeholder. Wire real dependencies (and the migrated *sql.DB from
// NewDataSource) into buildHandler.
func NewApplication(ctx context.Context, addr string) (*Application, error) {
	if ctx == nil {
		return nil, errors.New("jtspringproject: nil context passed to NewApplication")
	}
	if addr == "" {
		addr = defaultServerAddr
	}

	// MIGRATION_NOTE: This mirrors the manual Hibernate wiring that replaced
	// the excluded HibernateJpaAutoConfiguration. The loaded config / data
	// source should be injected into repositories once they are migrated.
	//
	// cfg, err := LoadHibernateConfig()
	// if err != nil {
	// 	return nil, fmt.Errorf("jtspringproject: load hibernate config: %w", err)
	// }
	// db, err := NewDataSource(ctx, cfg)
	// if err != nil {
	// 	return nil, fmt.Errorf("jtspringproject: create data source: %w", err)
	// }

	handler, err := buildHandler(ctx)
	if err != nil {
		return nil, fmt.Errorf("jtspringproject: build handler: %w", err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &Application{server: srv}, nil
}

// buildHandler assembles the HTTP handler chain (routing + middleware).
//
// REQUIRES MANUAL REVIEW: replace this placeholder with the migrated router
// that mounts the project's controllers. It is defined as a separate function
// so that dependency wiring stays explicit and testable.
func buildHandler(_ context.Context) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux, nil
}

// Run starts the embedded HTTP server and blocks until the context is cancelled
// or the server fails. This is the idiomatic Go replacement for
// SpringApplication.run(...): it owns the application lifecycle and performs a
// graceful shutdown.
//
// It returns an error describing why the server stopped; a clean shutdown
// (context cancellation) returns nil.
func (a *Application) Run(ctx context.Context) error {
	if a == nil || a.server == nil {
		return errors.New("jtspringproject: Run called on uninitialized Application")
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("jtspringproject: server error: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("jtspringproject: graceful shutdown failed: %w", err)
		}
		return nil
	}
}

// Run is the package-level convenience entry point equivalent to the Java
// main method body (SpringApplication.run). It wires signal handling for
// SIGINT/SIGTERM so the process shuts down gracefully, mirroring Spring Boot's
// registered shutdown hook.
//
// MIGRATION_NOTE: In Go the actual process entry point must live in package
// main (e.g. cmd/jtspringproject/main.go) and simply call this function. This
// keeps the bootstrap logic importable and testable, unlike Java's static
// main.
func Run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := defaultServerAddr
	// MIGRATION_NOTE: Java received String[] args; here we accept args for
	// parity but only support an optional first positional listen address.
	// Replace with a real flag/config parser (viper) as needed.
	if len(args) > 0 && args[0] != "" {
		addr = args[0]
	}

	app, err := NewApplication(ctx, addr)
	if err != nil {
		return fmt.Errorf("jtspringproject: initialize application: %w", err)
	}

	log.Printf("jtspringproject: starting server on %s", addr)
	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("jtspringproject: run application: %w", err)
	}
	log.Print("jtspringproject: server stopped cleanly")
	return nil
}
