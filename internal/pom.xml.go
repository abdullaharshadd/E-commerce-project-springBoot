// Package internal documents the migration of the original Maven POM
// (pom.xml) for the JtSpringProject Spring Boot 2.6.4 web application to the
// Go ecosystem.
//
// MIGRATION_NOTE: pom.xml is a *build manifest*, not application code. There is
// no meaningful line-by-line Go translation. Maven's role (dependency
// resolution, version management, packaging) is fulfilled in Go by the module
// system: go.mod / go.sum plus the `go build` / `go test` toolchain. The
// content below is therefore a documented mapping from each Maven artifact to
// its idiomatic Go equivalent, together with a reference go.mod. This file
// compiles (it is valid Go) but contains no runtime logic; it exists so the
// migration is self-describing and reviewable.
//
// REQUIRES MANUAL REVIEW: The actual dependency wiring (web framework, ORM,
// auth middleware, templating) must be implemented in real Go source files.
// This file only records the intended mapping.
package internal

// ProjectMetadata mirrors the identity fields declared in the original POM.
//
// Original Maven coordinates:
//
//	groupId:    com.jtspringproject
//	artifactId: JtSpringProject
//	version:    0.0.1-SNAPSHOT
//
// In Go the equivalent identity is the module path declared in go.mod, e.g.
//
//	module github.com/jtspringproject/jtspringproject
//
// There is no notion of a "SNAPSHOT" version in Go modules; use semantic
// version tags (v0.0.1) or pseudo-versions from commit hashes instead.
type ProjectMetadata struct {
	// Name is the human-readable project name (was <name> in the POM).
	Name string
	// Description is the project description (was <description> in the POM).
	Description string
	// ModulePath is the Go module path replacing groupId/artifactId.
	ModulePath string
	// Version is the semantic version replacing the Maven <version>.
	Version string
}

// NewProjectMetadata returns the metadata describing this project, ported from
// the original pom.xml identity block.
func NewProjectMetadata() ProjectMetadata {
	return ProjectMetadata{
		Name:        "JtSpringProject",
		Description: "Spring project for Java Technology (migrated to Go)",
		ModulePath:  "github.com/jtspringproject/jtspringproject",
		Version:     "v0.0.1",
	}
}

// DependencyMapping documents how a single Maven dependency maps into the Go
// ecosystem. It is metadata only; the real dependency is pinned in go.mod.
type DependencyMapping struct {
	// MavenArtifact is the original groupId:artifactId (:version) coordinate.
	MavenArtifact string
	// GoModule is the recommended replacement Go module import path.
	GoModule string
	// Purpose explains what capability the dependency provides.
	Purpose string
	// DevOnly reports whether this is a development-time-only tool (not a
	// runtime dependency).
	DevOnly bool
	// Note carries a MIGRATION_NOTE-style caveat for reviewers.
	Note string
}

// DependencyMap returns the recommended Go equivalents for every dependency
// declared in the original pom.xml.
//
// MIGRATION_NOTE: Spring Boot "starters" are opinionated bundles that pull in
// many transitive dependencies and auto-configure them. Go has no equivalent
// auto-magic; every capability below must be wired explicitly (routing,
// middleware, ORM setup, auth filter chain).
func DependencyMap() []DependencyMapping {
	return []DependencyMapping{
		{
			MavenArtifact: "org.junit.jupiter:junit-jupiter:5.10.0",
			GoModule:      "stdlib testing + github.com/stretchr/testify",
			Purpose:       "Unit/integration testing framework.",
			DevOnly:       true,
			Note:          "Use table-driven tests with the stdlib `testing` package; add testify only for richer assertions/mocks.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-web",
			GoModule:      "github.com/go-chi/chi/v5 (or github.com/gin-gonic/gin) + net/http",
			Purpose:       "Embedded HTTP server + MVC/routing.",
			Note:          "net/http already embeds the server; choose chi or gin for routing/middleware. No embedded Tomcat needed.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-test",
			GoModule:      "stdlib testing + net/http/httptest",
			Purpose:       "Web-layer test utilities.",
			DevOnly:       true,
			Note:          "Use httptest.NewServer / httptest.NewRecorder for handler tests.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-devtools",
			GoModule:      "github.com/air-verse/air (or github.com/cosmtrek/air)",
			Purpose:       "Hot reload during development.",
			DevOnly:       true,
			Note:          "Dev-only watcher; never a runtime/import dependency. Configure via .air.toml, not go.mod.",
		},
		{
			MavenArtifact: "javax.servlet:jstl",
			GoModule:      "html/template (stdlib)",
			Purpose:       "Server-side view templating (was JSP/JSTL).",
			Note:          "Decide: keep server-rendered views via html/template, or switch to an API-only architecture and drop templating entirely.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-data-jpa",
			GoModule:      "gorm.io/gorm (or database/sql + sqlx)",
			Purpose:       "ORM / persistence layer (was Hibernate).",
			Note:          "GORM is the closest ORM analogue; prefer database/sql+sqlx for explicit control. No entity auto-scanning \u2014 register models explicitly.",
		},
		{
			MavenArtifact: "org.apache.tomcat.embed:tomcat-embed-jasper",
			GoModule:      "html/template (stdlib)",
			Purpose:       "JSP (Jasper) rendering engine.",
			Note:          "No JSP in Go. Rendered together with the JSTL mapping above via html/template.",
		},
		{
			MavenArtifact: "mysql:mysql-connector-java:8.0.33",
			GoModule:      "github.com/go-sql-driver/mysql",
			Purpose:       "MySQL database driver.",
			Note:          "DSN replaces application.properties JDBC URL. Set parseTime=true and loc=... for the timezone handling Spring did implicitly; add tls=true/skip-verify for SSL.",
		},
		{
			MavenArtifact: "org.springframework.boot:spring-boot-starter-security",
			GoModule:      "custom middleware + golang.org/x/crypto/bcrypt (+ github.com/golang-jwt/jwt/v5 if using JWT)",
			Purpose:       "Authentication & authorization filter chain.",
			Note:          "CRITICAL: no auto-configured security in Go. Implement authn/authz as explicit middleware; hash passwords with bcrypt; manage sessions or JWTs yourself.",
		},
	}
}

// BuildTooling documents how the Maven build/packaging plugin maps to Go.
//
// MIGRATION_NOTE: <spring-boot-maven-plugin> produced an executable fat JAR.
// The Go equivalent is a statically linked binary produced by `go build`:
//
//	go build -o bin/jtspringproject ./cmd/server
//
// Java language level 11 (<java.version>11</java.version>) maps to the Go
// toolchain version pinned in go.mod (e.g. `go 1.22`). There is no separate
// packaging step \u2014 the binary is self-contained.
func BuildTooling() []string {
	return []string{
		"go build ./...            # compile all packages",
		"go test ./...             # run all tests",
		"go build -o bin/server ./cmd/server  # produce the self-contained binary (replaces fat JAR)",
	}
}

/*
MIGRATION_NOTE: Reference go.mod to be created at the repository root.
This is the true replacement for pom.xml. Pin explicit versions (Maven let the
Spring Boot 2.6.4 BOM resolve the unversioned starters; Go requires explicit
pins in go.mod/go.sum):

    module github.com/jtspringproject/jtspringproject

    go 1.22

    require (
        github.com/go-chi/chi/v5      v5.0.12  // web routing (was spring-boot-starter-web)
        github.com/go-sql-driver/mysql v1.8.1  // MySQL driver (was mysql-connector-java 8.0.33)
        gorm.io/gorm                  v1.25.10 // ORM (was spring-boot-starter-data-jpa / Hibernate)
        gorm.io/driver/mysql          v1.5.6
        golang.org/x/crypto           v0.24.0  // bcrypt for auth (was spring-boot-starter-security)
        github.com/stretchr/testify   v1.9.0   // test assertions (was junit-jupiter) // indirect at runtime
    )

Dev-only tools (air for hot reload, formerly spring-boot-devtools) are installed
via `go install github.com/air-verse/air@latest`, NOT declared as module
dependencies.
*/
