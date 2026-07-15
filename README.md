```markdown
# E-Commerce Project (Go Migration)

A web-based e-commerce application with admin and user-facing functionality, including product catalog management, category management, shopping cart, and user authentication. This repository is a migration of [abdullaharshadd/E-commerce-project-springBoot](https://github.com/abdullaharshadd/E-commerce-project-springBoot) from Java/Spring Boot to Go using the standard library and idiomatic Go tooling.

> **⚠️ Migration Confidence: 0% — This codebase requires significant manual review and completion before it is production-ready. Do not deploy without addressing the items listed in [Known Limitations](#known-limitations) and [Manual Review Required](#manual-review-required).**

---

## Tech Stack

| Concern | Technology |
|---|---|
| Language | Go (1.21+) |
| HTTP server | `net/http` (standard library) |
| ORM / DB access | [GORM](https://gorm.io) |
| Database | MySQL (migrated from the original MySQL config) |
| Authentication | Manual session/cookie handling or [gorilla/sessions](https://github.com/gorilla/sessions) |
| Templating | `html/template` (standard library) |
| Dependency management | Go modules (`go.mod` / `go.sum`) |
| Schema migrations | [golang-migrate](https://github.com/golang-migrate/migrate) or [goose](https://github.com/pressly/goose) |
| Testing | `testing` (standard library) |

---

## Prerequisites

- Go 1.21 or later — [install](https://go.dev/dl/)
- MySQL 8.0 or later running and accessible
- Node.js / npm (only if the project includes front-end asset build steps — `npm install` was detected in the setup plan; verify whether this applies to static assets)
- `golang-migrate` CLI or `goose` CLI if you are running schema migrations manually

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/abdullaharshadd/E-commerce-project-springBoot.git
cd E-commerce-project-springBoot
```

### 2. Install Go dependencies

```bash
go mod tidy
```

### 3. Install front-end dependencies (if applicable)

A `package.json` / `npm install` step was detected during migration analysis. Verify whether static assets require a build step:

```bash
npm install
```

If no `package.json` exists in the migrated output, skip this step.

### 4. Configure environment variables

Copy the example environment file and fill in values for your environment:

```bash
cp .env.example .env
```

Edit `.env` with your database credentials and any other required values. See the [Environment Variables](#environment-variables) table below.

### 5. Set up the database

Create the database in MySQL:

```sql
CREATE DATABASE ecommerce CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Run schema migrations. The original project used `hibernate.hbm2ddl.auto=update` (auto-schema generation), which has **not** been ported. You must apply migrations explicitly:

```bash
# Using golang-migrate
migrate -path ./migrations -database "mysql://USER:PASSWORD@tcp(HOST:PORT)/ecommerce" up

# OR using goose
goose -dir ./migrations mysql "USER:PASSWORD@tcp(HOST:PORT)/ecommerce" up
```

> **Note:** Migration SQL files must be authored manually from the original entity definitions. See [Migration Notes](#migration-notes) for details.

### 6. Run the application

```bash
go run ./cmd/server
```

Or build and run the binary:

```bash
go build -o ecommerce ./cmd/server
./ecommerce
```

The server will start on the port defined by the `APP_PORT` environment variable (default: `8080`).

---

## Running Tests

```bash
go test ./...
```

To run tests with verbose output:

```bash
go test -v ./...
```

> **Note:** Test coverage is likely incomplete as a result of the migration. The original Spring Boot test infrastructure (JUnit, MockMvc, Spring test context) has no direct equivalent and must be rewritten using Go's `testing` package and `net/http/httptest`.

---

## Environment Variables

These variables must be set before running the application. None were automatically detected from the original `application.properties` due to the use of non-standard Spring property keys; the table below reflects the values that must be manually mapped.

| Variable | Description | Example |
|---|---|---|
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_NAME` | Database name | `ecommerce` |
| `DB_USER` | Database username | `root` |
| `DB_PASSWORD` | Database password | `secret` |
| `APP_PORT` | Port the HTTP server listens on | `8080` |
| `SESSION_SECRET` | Secret key for signing session cookies | `change-me-in-production` |

Create a `.env.example` file in the repository root with these keys (values blank) for developer onboarding.

---

## Architecture Overview

The migrated code follows a layered structure that mirrors the original Spring MVC package layout, translated to Go conventions:

```
.
├── cmd/
│   └── server/         # Application entry point (main.go)
├── internal/
│   ├── config/         # Database connection setup, environment loading
│   ├── models/         # GORM struct definitions (User, Product, Category, Cart, CartProduct)
│   ├── dao/            # Data access layer (replaces Spring @Repository / Hibernate DAOs)
│   ├── handler/        # HTTP handlers (replaces Spring @Controller classes)
│   │   ├── admin/      # Admin controller logic
│   │   └── user/       # User-facing controller logic
│   ├── middleware/      # Authentication and error-handling middleware
│   └── router/         # Route registration (replaces Spring RequestMapping)
├── migrations/         # SQL schema migration files (must be authored manually)
├── static/             # Static assets (CSS, JS, images)
├── templates/          # HTML templates (html/template format)
├── go.mod
├── go.sum
└── .env.example
```

### Request flow

```
HTTP Request
  → router (net/http ServeMux or chi/gorilla/mux)
    → middleware (auth check, session validation)
      → handler (business logic)
        → dao (GORM database calls)
          → MySQL
```

Transaction management is **explicit**: each DAO function that requires a transaction accepts a `*gorm.DB` handle. Callers in the handler or a service layer are responsible for calling `db.Begin()`, `tx.Commit()`, and `tx.Rollback()`. There is no `@Transactional` annotation equivalent.

---

## Migration Notes

The following summarizes the significant structural changes made during migration from Java/Spring Boot to Go.

### Build system
- **Before:** Maven `pom.xml` with Spring Boot starter parent, Hibernate, Tomcat Jasper, Spring Security.
- **After:** Go modules (`go.mod`). Each Spring Boot starter has been replaced with the closest Go library. The Maven dependency tree does not translate automatically; dependencies were selected manually per capability.

### Dependency injection
- **Before:** Spring IoC container with `@Autowired`, `@Component`, `@Service`, `@Repository`.
- **After:** No DI framework. Dependencies are passed explicitly via constructors or function parameters.

### ORM and database access
- **Before:** Hibernate `SessionFactory`, `@Entity`, `@Transactional`, `hbm2ddl.auto=update` for schema management.
- **After:** GORM with explicit struct registration. Schema management is handled by `golang-migrate` or `goose` with versioned SQL files. `hbm2ddl.auto=update` behavior has been deliberately removed.

### Transaction management
- **Before:** Spring AOP `@Transactional` on DAO and service methods.
- **After:** Explicit `db.Transaction(func(tx *gorm.DB) error { ... })` blocks or manual `tx.Begin()` / `tx.Commit()` / `tx.Rollback()` in handler/service code.

### Security / Authentication
- **Before:** Spring Security `SecurityConfiguration` with filter chains, role-based access control, password encoding, and CSRF protection.
- **After:** Manual middleware. Session cookies are managed with `gorilla/sessions` or equivalent. Role checks are inline in handler middleware. CSRF protection must be implemented explicitly if required.

### Templating
- **Before:** JSP templates rendered server-side by Tomcat Jasper.
- **After:** `html/template` Go templates. JSP syntax (JSTL, EL expressions, taglibs) has been rewritten in Go template syntax. Template file locations and names may have changed.

### Error handling
- **Before:** Spring `@ControllerAdvice` / `ErrorController`.
- **After:** Custom middleware or handler wrapper functions that write appropriate HTTP status codes and error pages.

### Lazy loading
- **Before:** `hibernate.enable_lazy_load_no_trans=true` allowed lazy-loaded associations to be resolved outside a transaction session (an anti-pattern).
- **After:** All associations are loaded eagerly via GORM `Preload()` calls or explicit joins at the point of query. No out-of-transaction lazy loading exists.

---

## Known Limitations

The following components could not be automatically migrated and require manual implementation. The migration tool reported **0% overall confidence**, meaning the entire codebase should be treated as a scaffold requiring developer completion rather than a finished port.

| File / Component | Reason | Recommended Action |
|---|---|---|
| `pom.xml` — entire file | Maven/JVM-specific; no 1:1 Go equivalent | Author `go.mod` manually, mapping each Spring starter to the closest Go library |
| `HibernateConfiguration.java` — `HibernateTransactionManager` / `@EnableTransactionManagement` | Spring AOP declarative transactions have no Go equivalent | Rewrite all transaction boundaries explicitly using `db.Transaction()` or `sql.Tx`; audit every former `@Transactional` method |
| `HibernateConfiguration.java` — `LocalSessionFactoryBean` with `packagesToScan` | Hibernate entity scanning and dialect concepts don't exist in Go | Register GORM models explicitly; use `AutoMigrate` for development or migration files for production |
| `HibernateConfiguration.java` — `DriverManagerDataSource` | JDBC URL format and no-pool behavior are Java-specific | Use GORM with `database/sql` and configure a connection pool (`SetMaxOpenConns`, `SetMaxIdleConns`) |
| `application.properties` — custom `db.*` keys | Non-standard keys require a companion `@Configuration` class to bind; that wiring logic is what actually matters | Rewrite DataSource/connection setup in `internal/config/` using environment variables |
| `application.properties` — `hibernate.enable_lazy_load_no_trans` | Anti-pattern with no Go equivalent | Fix all call sites to load required associations eagerly within the same query/transaction |
| `application.properties` — `hbm2ddl.auto=update` | Auto-schema mutation is Hibernate runtime magic | Write explicit migration SQL files and manage them with `golang-migrate` or `goose` |
| `SecurityConfiguration.java` | Spring Security filter chain, BCrypt, CSRF, role-based URL rules | Manually implement authentication middleware, password hashing (`golang.org/x/crypto/bcrypt`), and route-level role checks |
| All DAO files (`cartDao`, `categoryDao`, `cartProductDao`, `productDao`, `userDao`) | `@Transactional` and `SessionFactory.getCurrentSession()` thread-bound session have no Go equivalent | Rewrite each DAO to accept `*gorm.DB` explicitly; move transaction boundaries to the handler or service layer |

---

## Manual Review Required

The following files were flagged as low-confidence during migration and **must be manually verified** by a developer before the application is considered functional. Treat these files as first drafts or placeholders.

- `pom.xml`
- `src/main/resources/application.properties` → `internal/config/config.go`
- `src/main/java/com/jtspringproject/JtSpringProject/configuration/SecurityConfiguration.java` → authentication middleware
- `src/main/java/com/jtspringproject/JtSpringProject/controller/AdminController.java` → `internal/handler/admin/`
- `src/main/java/com/jtspringproject/JtSpringProject/controller/UserController.java` → `internal/handler/user/`
- `src/main/java/com/jtspringproject/JtSpringProject/controller/ErrorController.java` → error middleware
- `src/main/java/com/jtspringproject/JtSpringProject/models/Cart.java` → `internal/models/cart.go`
- `src/main/java/com/jtspringproject/JtSpringProject/models/CartProduct.java` → `internal/models/cart_product.go`
- `src/main/java/com/jtspringproject/JtSpringProject/models/Category.java` → `internal/models/category.go`
- `src/main/java/com/jtspringproject/JtSpringProject/dao/cartDao.java` → `internal/dao/cart_dao.go`
- `src/main/java/com/jtspringproject/JtSpringProject/dao/categoryDao.java` → `internal/dao/category_dao.go`
- `src/main/java/com/jtspringproject/JtSpringProject/dao/cartProductDao.java` → `internal/dao/cart_product_dao.go`
- `src/main/java/com/jtspringproject/JtSpringProject/dao/productDao.java` → `internal/dao/product_dao.go`
- `src/main/java/com/jtspringproject/JtSpringProject/dao/userDao.java` → `internal/dao/user_dao.go`

### Suggested review checklist for each file

- [ ] GORM struct tags match the original database schema column names and types
- [ ] All former `@Transactional` boundaries are now explicit `db.Transaction()` calls
- [ ] No remaining references to Spring, Hibernate, or Java-specific imports
- [ ] HTTP route paths and HTTP methods match the original `@RequestMapping` / `@GetMapping` / `@PostMapping` annotations
- [ ] Form field names in templates match the struct field names used in handlers
- [ ] Password hashing uses `bcrypt` and is not stored in plaintext
- [ ] Admin-only routes are protected by role-checking middleware
- [ ] Schema migration SQL files are present for every GORM model

---

## Original Project Reference

- Original repository: [abdullaharshadd/E-commerce-project-springBoot](https://github.com/abdullaharshadd/E-commerce-project-springBoot)
- Original stack: Java 8+, Spring Boot, Spring Security, Hibernate, MySQL, JSP/JSTL, Maven
- Modules migrated: 27 / 27 (all files have a Go counterpart, but confidence is 0% — counterparts are scaffolds, not verified implementations)
```