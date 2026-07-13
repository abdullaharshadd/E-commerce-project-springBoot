// Package jtspringproject contains the Go migration of the JtSpringProject
// application, including the migrated equivalent of the Spring Boot generated
// test class JtSpringProjectApplicationTests.java.
//
// In the Spring Boot world this test class looked like:
//
//	@SpringBootTest
//	class JtSpringProjectApplicationTests {
//	    @Test
//	    void contextLoads() {
//	    }
//	}
//
// The single @SpringBootTest / @Test method (contextLoads) is a smoke test:
// it bootstraps the entire Spring ApplicationContext and asserts — implicitly,
// by not throwing — that every bean can be constructed and wired together.
//
// MIGRATION_NOTE: Go has no equivalent of the Spring ApplicationContext, so
// there is nothing to "load" in the framework sense. The idiomatic Go analogue
// of this smoke test is to verify that the application's configuration and core
// domain objects can be constructed successfully without error. We therefore
// reproduce the spirit of contextLoads by:
//
//   - Loading the migrated configuration (resources.LoadConfig / DefaultConfig)
//     and validating it, which mirrors Spring reading application.properties
//     and failing fast on invalid values.
//   - Constructing the migrated core domain models (model.NewUser) to ensure
//     the wiring of the domain layer is sound.
//
// This keeps the test meaningful in Go terms while preserving the original
// intent: "the application can start up cleanly."
package jtspringproject

import (
	"testing"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
	"github.com/jtspringproject/JtSpringProject/internal/resources"
)

// TestContextLoads is the Go migration of the Spring Boot contextLoads() smoke
// test. In Spring the test passed simply by the ApplicationContext starting
// without throwing. Here we assert the Go equivalent: the application's
// configuration loads and validates, and the core domain objects can be
// constructed. Any failure is reported via t.Fatalf, mirroring the fail-fast
// behaviour of a broken Spring context.
func TestContextLoads(t *testing.T) {
	// Loading configuration mirrors Spring reading application.properties during
	// context startup. LoadConfig applies defaults and environment overrides.
	cfg, err := resources.LoadConfig()
	if err != nil {
		// MIGRATION_NOTE: If LoadConfig is not available or has a different
		// signature in the migrated resources package, fall back to
		// resources.DefaultConfig() here during manual review.
		t.Fatalf("configuration failed to load (context did not start): %v", err)
	}

	if cfg == nil {
		t.Fatal("expected a non-nil configuration after load")
	}

	// Validate mirrors Spring's bean validation / fail-fast on invalid context.
	if err := cfg.Validate(); err != nil {
		t.Fatalf("configuration is invalid (context would not start): %v", err)
	}

	// Constructing a core domain object verifies the domain layer wires up, in
	// the same spirit as Spring instantiating its beans.
	user := model.NewUser()
	if user == nil {
		t.Fatal("expected NewUser to return a non-nil User")
	}
}
