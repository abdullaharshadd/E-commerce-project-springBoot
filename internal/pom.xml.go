// Package internal documents the build/dependency migration from the original
// Maven POM (pom.xml) to the Go ecosystem.
//
// MIGRATION_NOTE: pom.xml is a *build descriptor*, not runtime Go code. There is
// no idiomatic way to express Maven build configuration as executable Go. The
// correct Go equivalent of a Maven POM is a `go.mod` file (dependency manifest)
// plus optional tooling (Makefile / Taskfile / build scripts). This file exists
// only to record the migration decisions and to map each Maven dependency onto
// its recommended Go counterpart, so a human reviewer can reproduce the build.
//
// This file intentionally contains NO business logic and registers NO HTTP
// routes: the source POM defines none. Route registration belongs in the
// migrated controller/handler files, not in the build manifest.
package internal

// MavenDependency describes a single dependency declared in the original POM,
// together with the recommended Go replacement. It is used purely for
// documentation / migration-tracking purposes.
type MavenDependency struct {
	// GroupID is the Maven groupId (e.g. "org.springframework.boot").
	GroupID string
	// ArtifactID is the Maven artifactId (e.g. "spring-boot-starter-web").
	ArtifactID string
	// Version is the explicit version, or empty when managed by the parent BOM.
	Version string
	// Scope is the Maven scope ("", "test", etc.).
	Scope string
	// GoReplacement is the recommended Go module (or "" if no runtime equivalent).
	GoReplacement string
	// Note captures any manual-review guidance for this dependency.
	Note string
}

// ProjectMetadata mirrors the identifying fields of the original Maven project.
type ProjectMetadata struct {
	// GroupID is the original Maven groupId.
	GroupID string
	// ArtifactID is the original Maven artifactId.
	ArtifactID string
	// Version is the original project version.
	Version string
	// Name is the human-readable project name.
	Name string
	// Description is the project description.
	Description string
	// JavaVersion records the source JDK target (informational only).
	JavaVersion string
	// SpringBootVersion records the parent BOM version (informational only).
	SpringBootVersion string
}

// SourceProject returns the metadata of the original Spring Boot project.
//
// The values are preserved verbatim from pom.xml so that reviewers can confirm
// the migration corresponds to the correct source artifact.
func SourceProject() ProjectMetadata {
	return ProjectMetadata{
		GroupID:           "com.jtspringproject",
		ArtifactID:        "JtSpringProject",
		Version:           "0.0.1-SNAPSHOT",
		Name:              "JtSpringProject",
		Description:       "Spring project for Java Technology",
		JavaVersion:       "11",
		SpringBootVersion: "2.6.4",
	}
}

// DependencyMigrationPlan returns the recommended Go equivalents for every
// dependency declared in the original POM.
//
// MIGRATION_NOTE: Versions left blank in the source POM were pinned implicitly
// by spring-boot-starter-parent:2.6.4. A human must decide equivalent Go module
// versions; the versions below are current, well-maintained defaults.
func DependencyMigrationPlan() []MavenDependency {
	return []MavenDependency{
		{
			GroupID:       "org.junit.jupiter",
			ArtifactID:    "junit-jupiter",
			Version:       "5.10.0",
			Scope:         "test",
			GoReplacement: "github.com/stretchr/testify (assert/require) + stdlib testing",
			Note:          "Use table-driven tests with the stdlib `testing` package.",
		},
		{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-starter-web",
			GoReplacement: "net/http + github.com/go-chi/chi/v5",
			Note:          "Embedded HTTP server via net/http; chi for routing/middleware.",
		},
		{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-starter-test",
			Scope:         "test",
			GoReplacement: "github.com/stretchr/testify + net/http/httptest",
		},
		{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-devtools",
			GoReplacement: "",
			Note:          "Dev-only hot reload. Use `air` (github.com/air-verse/air) or `reflex` during development; no runtime dependency.",
		},
		{
			GroupID:       "javax.servlet",
			ArtifactID:    "jstl",
			GoReplacement: "html/template (stdlib)",
			Note:          "Server-side view rendering: migrate JSP/JSTL views to html/template templates.",
		},
		{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-starter-data-jpa",
			GoReplacement: "gorm.io/gorm (or database/sql + sqlx)",
			Note:          "Replace JPA/Hibernate entities and repositories with GORM models or a repository layer over database/sql.",
		},
		{
			GroupID:       "org.apache.tomcat.embed",
			ArtifactID:    "tomcat-embed-jasper",
			GoReplacement: "html/template (stdlib)",
			Note:          "Jasper is the JSP compiler for embedded Tomcat; superseded by html/template + net/http server.",
		},
		{
			GroupID:       "mysql",
			ArtifactID:    "mysql-connector-java",
			Version:       "8.0.33",
			GoReplacement: "github.com/go-sql-driver/mysql (with gorm.io/driver/mysql if using GORM)",
		},
		{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-starter-security",
			GoReplacement: "custom middleware (net/http) + golang.org/x/crypto/bcrypt",
			Note:          "Spring Security has no drop-in Go equivalent. Reimplement auth explicitly: bcrypt for password hashing, gorilla/sessions or JWT (github.com/golang-jwt/jwt) for auth state, and chi middleware for the filter chain. REQUIRES MANUAL REVIEW.",
		},
	}
}
