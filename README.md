```markdown
# E-Commerce Project (Go Migration)

A web-based e-commerce application with admin and user-facing functionality, including product management, categories, shopping cart, and user authentication. Originally built with Java/Spring Boot; this repository is the migrated Go implementation using the standard library and idiomatic Go packages.

> **⚠️ Migration Confidence: 0% — This codebase requires significant manual review before it is production-ready. Do not deploy without completing the steps in the [Manual Review Required](#manual-review-required) section.**

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go (1.21+) |
| HTTP Server | `net/http` (standard library) |
| ORM / DB Access | To be determined — see Migration Notes |
| Database | MySQL (inherited from original) |
| Templating | Go `html/template` (replaces JSP/JSTL) |
| Auth / Sessions | To be determined — replaces Spring Security |
| Schema Migrations | `golang-migrate` or `goose` (replaces `hbm2ddl.auto`) |
| Build | `go build` / `go run` |
| Frontend deps | npm (detected in setup — see note below) |

---

## Prerequisites

- Go 1.21 or later
- MySQL 5.7+ or MariaDB 10.5+
- Node.js / npm (required for any frontend asset pipeline — verify actual usage)
- `golang-migrate` CLI or `goose` (for database migrations)

Install Go: https://go.dev/dl/  
Install Node.js: https://nodejs.org/

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/abdullaharshadd/E-commerce-project-springBoot.git
cd E-commerce-project-springBoot
```

### 2. Install dependencies

```bash
# Go dependencies
go mod download

# Frontend dependencies (if applicable — verify what npm manages here)
npm install
```

### 3. Configure environment

Copy the example environment file and fill in your values:

```bash
cp .env.example .env
```

Edit `.env` with your database credentials and other settings. See the [Environment Variables](#environment-variables) table below.

### 4. Set up the database

Create the database manually:

```sql
CREATE DATABASE ecommerce CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Run schema migrations (replace tool with whichever is adopted):

```bash
# Using golang-migrate
migrate -path ./migrations -database "mysql://USER:PASSWORD@tcp(HOST:PORT)/ecommerce" up

# OR using goose
goose -dir ./migrations mysql "USER:PASSWORD@tcp(HOST:PORT)/ecommerce" up
```

> **Note:** The original project relied on Hibernate `hbm2ddl.auto` for schema generation. This has no Go equivalent. Migration files must be authored manually based on the entity models (`Cart`, `CartProduct`, `Category`, `User`, `Product`). See [Known Limitations](#known-limitations).

### 5. Run the application

```bash
go run ./cmd/server
```

Or build and run the binary:

```bash
go build -o ecommerce ./cmd/server
./ecommerce
```

The server will start on the port defined in your environment configuration (default: `8080`).

---

## Running Tests

```bash
go test ./...
```

To run tests with verbose output:

```bash
go test -v ./...
```

To run tests with coverage:

```bash
go test -cover ./...
```

> **Note:** The original test suite covered `CartProductId`, `CartProduct`, and application context loading. These tests were among the low-confidence migrations and must be manually verified for correctness. See [Manual Review Required](#manual-review-required).

---

## Environment Variables

These variables must be set before running the application. No defaults are safe for production.

| Variable | Description | Example |
|---|---|---|
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_NAME` | Database name | `ecommerce` |
| `DB_USER` | Database username | `root` |
| `DB_PASSWORD` | Database password | `secret` |
| `SERVER_PORT` | HTTP server listen port | `8080` |
| `SESSION_SECRET` | Secret key for session signing | `change-me-32-chars-minimum` |
| `ADMIN_ROLE` | Role identifier for admin users | `ADMIN` |

> The original project used `application.properties` with custom-prefixed keys (`db.*`, `hibernate.*`). These have been replaced with the variables above. Verify that all references in the migrated code match these names exactly.

---

## Architecture Overview

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point; explicit dependency wiring
├── internal/
│   ├── config/              # Environment and database configuration
│   ├── models/              # Structs: User, Cart, CartProduct, CartProductId, Category, Product
│   ├── dao/                 # Data access layer (replaces Spring @Repository / Hibernate DAOs)
│   │   ├── cart.go
│   │   ├── cartproduct.go
│   │   ├── category.go
│   │   ├── product.go
│   │   └── user.go
│   ├── services/            # Business logic (replaces Spring @Service)
│   │   ├── cart.go
│   │   └── user.go
│   ├── handlers/            # HTTP handlers (replaces Spring @Controller)
│   │   ├── admin.go
│   │   ├── user.go
│   │   └── error.go
│   └── middleware/          # Auth, session, and security middleware (replaces Spring Security)
├── migrations/              # SQL migration files (replaces hbm2ddl.auto)
├── templates/               # Go html/template files (replaces JSP/JSTL views)
├── static/                  # Static assets (CSS, JS, images)
├── go.mod
├── go.sum
└── .env.example
```

**Key architectural decisions:**

- Dependencies are wired explicitly in `cmd/server/main.go` — there is no IoC container or component scanning.
- Transactions are managed explicitly at the service layer using `db.Begin()` / `Commit()` / `Rollback()` rather than declarative `@Transactional` annotations.
- HTTP routing is handled via `net/http` `ServeMux` or a thin router. All routes are registered explicitly.
- Password hashing replaces Spring Security's `BCryptPasswordEncoder` — verify the chosen library produces compatible hashes if migrating an existing user database.

---

## Migration Notes

This project was migrated from Java 17 / Spring Boot to Go. The following are the most significant structural changes:

| Original (Spring Boot) | Migrated (Go) |
|---|---|
| `@SpringBootApplication` + auto-configuration | Explicit `main.go` composition root |
| Spring Security (`SecurityConfiguration.java`) | Custom middleware (session + role checks) |
| Hibernate / JPA entity scanning + `hbm2ddl.auto` | Explicit struct registration + SQL migration files |
| `@Transactional` on service methods | Explicit `db.Begin/Commit/Rollback` in service layer |
| `DriverManagerDataSource` / JDBC URL | `database/sql` with Go DSN string |
| Spring `@Controller` / `@RestController` | `net/http` handler functions |
| JSP + JSTL views | `html/template` templates |
| Maven `pom.xml` + `spring-boot-maven-plugin` | `go.mod` + `go build` |
| `BCryptPasswordEncoder` via `PasswordEncoderConfig` | Go bcrypt (`golang.org/x/crypto/bcrypt`) |
| `@Repository` DAOs with `SessionFactory` injection | Plain Go structs with `*sql.DB` |

---

## Known Limitations

The following components could not be mechanically migrated and require manual implementation. The overall migration confidence is **0%** — treat all migrated files as drafts requiring human verification.

### 1. Maven POM → `go.mod` (pom.xml)
Spring Boot starters bundle many transitive dependencies. There is no 1:1 mapping. Each starter's capabilities (security, data, web, etc.) must be satisfied by individually selected Go packages, pinned to explicit versions in `go.mod`.

### 2. Spring Boot Maven Plugin
Fat-JAR packaging and `spring-boot:run` have no equivalent. Use `go build` and `go run` respectively. If containerising, write a multi-stage `Dockerfile` manually.

### 3. JSP / JSTL Views (tomcat-embed-jasper + jstl)
JSP is a Java servlet technology. All views must be rewritten as `html/template` templates or the application must be refactored to a REST API with a separate frontend. No automatic conversion was performed.

### 4. Declarative Transaction Management (`@EnableTransactionManagement` / `@Transactional`)
Go has no AOP proxy equivalent. All service methods that were annotated with `@Transactional` in the original code must be audited. Transaction boundaries must be made explicit in the migrated service layer. **Audit every method in `cartService.java` and `userService.java` before relying on this code.**

### 5. Hibernate Package Scanning + `hbm2ddl.auto`
Automatic schema generation from annotations is not available in Go. SQL migration files must be authored manually from the entity definitions. The `enable_lazy_load_no_trans` property from `application.properties` was intentionally dropped — data access must be rewritten to load all required relations within transaction scope.

### 6. Spring Security
`SecurityConfiguration.java` implemented authentication, authorisation, password encoding, and session management. The Go equivalent is custom middleware that must be implemented and audited thoroughly. CSRF protection, session fixation, and role-based access control must be explicitly re-implemented.

### 7. Custom `application.properties` Keys (`db.*`, `hibernate.*`)
These non-standard property prefixes were bound by hand-written `@Configuration` classes. The mapping to Go environment variables may be incomplete. Verify every config value is correctly read in `internal/config/`.

---

## Manual Review Required

The following files were flagged as low-confidence or unmigrable. A developer must manually verify each one before the application is considered functional.

| File | What to verify |
|---|---|
| `cmd/server/main.go` (from `JtSpringProjectApplication.java`) | All dependencies are wired correctly; server starts and binds routes |
| `internal/config/` (from `HibernateConfiguration.java`) | DB DSN is correctly constructed; connection pool is configured; ORM/driver is registered |
| `internal/middleware/security.go` (from `SecurityConfiguration.java`) | Auth flow, session management, role checks, and password hashing all function correctly |
| `internal/config/password.go` (from `PasswordEncoderConfig.java`) | BCrypt cost factor matches original; existing password hashes from the original DB are compatible |
| `internal/handlers/admin.go` (from `AdminController.java`) | All admin routes are registered; auth middleware is applied; request/response shapes match |
| `internal/handlers/user.go` (from `UserController.java`) | All user routes are registered; session handling is correct |
| `internal/handlers/error.go` (from `ErrorController.java`) | Error responses are handled and formatted correctly |
| `internal/models/cart.go` (from `Cart.java`) | Struct fields, types, and DB column mappings are correct |
| `internal/models/cartproduct.go` (from `CartProduct.java`) | Composite relationship is modelled correctly; no lazy-load assumptions remain |
| `internal/models/cartproductid.go` (from `CartProductId.java`) | Composite key equivalence logic is correct |
| `internal/models/category.go` (from `Category.java`) | Struct fields match DB schema |
| `internal/models/user.go` (from `User.java`) | Password field handling; role field; no sensitive data accidentally serialised |
| `internal/dao/cart.go` (from `cartDao.java`) | All queries are correct SQL; transaction handling is explicit |
| `internal/dao/cartproduct.go` (from `cartProductDao.java`) | Composite key queries work correctly |
| `internal/dao/category.go` (from `categoryDao.java`) | CRUD operations are complete |
| `internal/dao/product.go` (from `productDao.java`) | CRUD operations are complete |
| `internal/dao/user.go` (from `userDao.java`) | Lookup by username/email works; password is never returned in plain text |
| `internal/services/cart.go` (from `cartService.java`) | Transaction boundaries are explicit; all `@Transactional` methods have been audited |
| `internal/services/user.go` (from `userService.java`) | Registration, login, and role assignment logic is correct |
| `migrations/` | Schema accurately reflects all original entity definitions |
| `templates/` | All JSP views have been rewritten as valid Go templates |
| Test files (from `JtSpringProjectApplicationTests.java`, `CartProductIdTest.java`, `CartProductTest.java`) | Tests compile, run, and assert correct behaviour in Go |

---

## Contributing

1. Resolve all items in [Manual Review Required](#manual-review-required) before adding new features.
2. Ensure `go vet ./...` and `go test ./...` pass cleanly.
3. Run `golangci-lint run` before submitting a pull request.

---

## Original Project

The original Spring Boot source is available at:  
https://github.com/abdullaharshadd/E-commerce-project-springBoot
```