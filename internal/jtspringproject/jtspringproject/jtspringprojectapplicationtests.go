package jtspringproject

// This file migrates JtSpringProjectApplicationTests.java.
//
// The original was the default Spring Boot generated integration test. It was
// annotated with @SpringBootTest, which bootstraps the entire Spring
// ApplicationContext, and contained a single empty @Test method
// (contextLoads). The convention is a "smoke test": if the full application
// context wires up all beans without throwing, the test passes. There are no
// assertions because the act of successfully loading the context IS the
// assertion.
//
// MIGRATION_NOTE: Go has no equivalent of Spring's ApplicationContext, no
// reflective bean container, and no @SpringBootTest that magically wires
// dependencies. In this codebase, "wiring" is done explicitly in Run (see
// jtspringprojectapplication.go), which constructs the config, data source,
// transactor, password encoder, handlers and routes.
//
// The idiomatic Go equivalent of a "context loads" smoke test is therefore to
// exercise the explicit construction path and assert that each constructor
// returns without error. We deliberately avoid calling Run itself, since Run
// starts a blocking HTTP server; instead we verify the individual wiring steps
// that Run performs. Steps that require a live database connection are guarded
// so the test does not fail in environments without one — mirroring the intent
// of the original (verify wiring, not external systems).

import (
	"testing"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/config"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/handler"
	"github.com/jtspringproject/JtSpringProject/internal/resources"
)

// TestContextLoads is the Go migration of the Spring Boot contextLoads smoke
// test. It verifies that the application's core components can be constructed
// and wired together without error, which is the Go analogue of successfully
// loading the Spring ApplicationContext.
func TestContextLoads(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "application config loads",
			run: func(t *testing.T) {
				// MIGRATION_NOTE: LoadAppConfig replaces Spring's
				// environment-driven auto-configuration that parsed
				// application.properties at startup.
				if _, err := resources.LoadAppConfig(); err != nil {
					t.Fatalf("LoadAppConfig() returned error: %v", err)
				}
			},
		},
		{
			name: "hibernate/datasource config loads",
			run: func(t *testing.T) {
				// MIGRATION_NOTE: LoadConfigFromEnv replaces the
				// Hibernate/JPA DataSource auto-configuration.
				if _, err := LoadConfigFromEnv(); err != nil {
					t.Fatalf("LoadConfigFromEnv() returned error: %v", err)
				}
			},
		},
		{
			name: "password encoder bean constructs",
			run: func(t *testing.T) {
				// Replaces the @Bean PasswordEncoder definition.
				if enc := config.NewBCryptPasswordEncoder(); enc == nil {
					t.Fatal("NewBCryptPasswordEncoder() returned nil")
				}
			},
		},
		{
			name: "error controller bean constructs",
			run: func(t *testing.T) {
				// Replaces the @Controller ErrorController bean.
				if ec := handler.NewErrorController(); ec == nil {
					t.Fatal("NewErrorController() returned nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
