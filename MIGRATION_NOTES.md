# Migration Notes

**Overall confidence:** 0%  
**Recommendation:** REVIEW RECOMMENDED

---

## What was migrated

- `pom.xml` → `internal/pom.xml.go` (81% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/HibernateConfiguration.java` → `internal/jtspringproject/jtspringproject/hibernateconfiguration.go` (13% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/JtSpringProjectApplication.java` → `internal/jtspringproject/jtspringproject/jtspringprojectapplication.go` (16% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/configuration/PasswordEncoderConfig.java` → `internal/jtspringproject/jtspringproject/config/passwordencoderconfig.go` (22% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/controller/ErrorController.java` → `internal/jtspringproject/jtspringproject/handler/errorcontroller.go` (18% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/models/Cart.java` → `internal/jtspringproject/jtspringproject/model/cart.go` (13% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/models/CartProduct.java` → `internal/jtspringproject/jtspringproject/model/cartproduct.go` (17% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/models/CartProductId.java` → `internal/jtspringproject/jtspringproject/model/cartproductid.go` (18% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/models/Category.java` → `internal/jtspringproject/jtspringproject/model/category.go` (24% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/models/Product.java` → `internal/jtspringproject/jtspringproject/model/product.go` (92% confidence)
- `src/main/java/com/jtspringproject/JtSpringProject/models/User.java` → `internal/jtspringproject/jtspringproject/model/user.go` (24% confidence) ⚠️ needs review
- `src/main/resources/application.properties` → `internal/resources/application.properties.go` (86% confidence)
- `src/test/java/com/jtspringproject/JtSpringProject/JtSpringProjectApplicationTests.java` → `internal/jtspringproject/jtspringproject/jtspringprojectapplicationtests.go` (12% confidence) ⚠️ needs review
- `src/test/java/com/jtspringproject/JtSpringProject/models/CartProductIdTest.java` → `internal/jtspringproject/jtspringproject/model/cartproductidtest.go` (80% confidence) ⚠️ needs review
- `src/test/java/com/jtspringproject/JtSpringProject/models/CartProductTest.java` → `internal/jtspringproject/jtspringproject/model/cartproducttest.go` (13% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/dao/cartDao.java` → `internal/jtspringproject/jtspringproject/repository/cartdao.go` (53% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/dao/categoryDao.java` → `internal/jtspringproject/jtspringproject/repository/categorydao.go` (18% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/dao/cartProductDao.java` → `internal/jtspringproject/jtspringproject/repository/cartproductdao.go` (12% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/dao/productDao.java` → `internal/jtspringproject/jtspringproject/repository/productdao.go` (17% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/dao/userDao.java` → `internal/jtspringproject/jtspringproject/repository/userdao.go` (17% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/services/cartService.java` → `internal/jtspringproject/jtspringproject/service/cartservice.go` (13% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/services/categoryService.java` → `internal/jtspringproject/jtspringproject/service/categoryservice.go` (88% confidence)
- `src/main/java/com/jtspringproject/JtSpringProject/services/productService.java` → `internal/jtspringproject/jtspringproject/service/productservice.go` (88% confidence)
- `src/main/java/com/jtspringproject/JtSpringProject/services/userService.java` → `internal/jtspringproject/jtspringproject/service/userservice.go` (13% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/configuration/SecurityConfiguration.java` → `internal/jtspringproject/jtspringproject/config/securityconfiguration.go` (76% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/controller/AdminController.java` → `internal/jtspringproject/jtspringproject/handler/admincontroller.go` (68% confidence) ⚠️ needs review
- `src/main/java/com/jtspringproject/JtSpringProject/controller/UserController.java` → `internal/jtspringproject/jtspringproject/handler/usercontroller.go` (82% confidence) ⚠️ needs review

## Components that could not be automatically migrated

These components require manual implementation. The migrated code contains
`MIGRATION_NOTE` comments at the relevant locations.

### `spring-boot-starter-parent / entire pom.xml` in `pom.xml`
**Reason:** A Maven POM is Java/JVM-ecosystem-specific build tooling with no 1:1 equivalent in other languages; the starter abstractions bundle many transitive dependencies whose exact composition cannot be mechanically translated.
**Suggestion:** Manually author the target ecosystem's dependency manifest (package.json, pyproject.toml, .csproj, go.mod) and map each starter to the closest framework/library, pinning explicit versions.

### `spring-boot-maven-plugin` in `pom.xml`
**Reason:** Fat-JAR executable packaging and 'spring-boot:run' are JVM/Maven-specific build lifecycle features.
**Suggestion:** Replace with the target ecosystem's build/bundling and run scripts (npm scripts, docker build, dotnet publish, go build).

### `tomcat-embed-jasper + jstl` in `pom.xml`
**Reason:** JSP/JSTL is a Java servlet-container-specific view technology with no direct port to non-JVM runtimes.
**Suggestion:** Manually rewrite JSP views using the target's templating engine, or refactor to a REST API + separate frontend.

### `transactionManager / @EnableTransactionManagement` in `src/main/java/com/jtspringproject/JtSpringProject/HibernateConfiguration.java`
**Reason:** Relies on Spring's AOP proxy-based declarative transaction management, which has no direct equivalent in Go; transaction boundaries are implicit via @Transactional annotations on service methods elsewhere.
**Suggestion:** Rewrite manually: use explicit transaction handling at the service layer (db.Begin/Commit/Rollback or an ORM's Transaction closure). Audit all @Transactional usages in the codebase before migrating.

### `sessionFactory (packagesToScan + hbm2ddl.auto)` in `src/main/java/com/jtspringproject/JtSpringProject/HibernateConfiguration.java`
**Reason:** Annotation-driven entity package scanning and automatic schema generation (hbm2ddl) have no equivalent in Go's explicit ORM ecosystem.
**Suggestion:** Manually register each model struct with the chosen ORM and replace hbm2ddl.auto with explicit migrations (golang-migrate/goose) or a deliberate AutoMigrate call.

### `dataSource (DriverManagerDataSource)` in `src/main/java/com/jtspringproject/JtSpringProject/HibernateConfiguration.java`
**Reason:** Spring's non-pooling DataSource abstraction and JDBC URL format are Java/JDBC-specific.
**Suggestion:** Replace with Go database/sql via a config struct; translate the JDBC URL to a Go DSN and configure connection pool limits explicitly.

### `@SpringBootApplication auto-configuration` in `src/main/java/com/jtspringproject/JtSpringProject/JtSpringProjectApplication.java`
**Reason:** There is no direct equivalent to Spring's classpath-based auto-configuration and component scanning in most target languages; this is framework 'magic' that resolves beans, HTTP server, and configuration at runtime.
**Suggestion:** Manually rewrite as an explicit application bootstrap: instantiate and wire dependencies in a composition root, start the HTTP server explicitly, and register routes/handlers manually or via the target framework's idioms.

### `exclude = HibernateJpaAutoConfiguration.class` in `src/main/java/com/jtspringproject/JtSpringProject/JtSpringProjectApplication.java`
**Reason:** This exclusion only has meaning within Spring's auto-configuration system; in the target it has no direct analog, but signals that ORM setup is manual/custom.
**Suggestion:** Ignore the exclusion annotation itself, but manually trace and migrate the custom persistence configuration this exclusion implies exists elsewhere in the project.

### `spring.jpa.properties.hibernate.enable_lazy_load_no_trans` in `src/main/resources/application.properties`
**Reason:** This is a Hibernate-specific escape hatch that silently opens sessions for lazy loading outside transaction boundaries. It has no clean equivalent in most target ORMs and encodes a design flaw (accessing lazy relations outside a transaction).
**Suggestion:** Do not carry over. Instead redesign data access with explicit eager fetches, fetch joins, entity graphs, or DTO projections in the target so relations are loaded within transaction scope.

### `db.* / hibernate.* / entitymanager.packagesToScan custom-prefixed keys` in `src/main/resources/application.properties`
**Reason:** These non-standard prefixes are only meaningful if a hand-written @Configuration bean binds them (via @Value/Environment). Without locating that class, the mapping to a target config schema is ambiguous and cannot be auto-migrated correctly.
**Suggestion:** Manually locate the corresponding Java DataSource/EntityManagerFactory configuration class, then map these values into the target framework's canonical datasource/ORM config (e.g., connection string, driver, entity package scan) explicitly.

### `spring.mvc.view.prefix/suffix (JSP views)` in `src/main/resources/application.properties`
**Reason:** JSP server-side rendering is tightly coupled to the Servlet/Spring MVC stack and has no direct equivalent in most non-JVM targets.
**Suggestion:** Rewrite the view layer using the target platform's templating engine or a decoupled frontend; the JSP files themselves require manual conversion.

### `@Transactional (all methods)` in `src/main/java/com/jtspringproject/JtSpringProject/dao/cartDao.java`
**Reason:** Spring's declarative transaction management uses AOP proxies and thread-bound sessions that have no automatic equivalent in Go; the transaction boundary is implicit and framework-driven.
**Suggestion:** Manually reimplement transactions in Go using database/sql Tx or the ORM's transaction API (e.g., gorm's db.Transaction(func(tx){...})). Preferably manage transaction boundaries at the service layer and pass the tx handle into DAO methods so multiple operations can be atomic.

### `sessionFactory.getCurrentSession()` in `src/main/java/com/jtspringproject/JtSpringProject/dao/cartDao.java`
**Reason:** Depends on Spring/Hibernate thread-local session context bound by the transaction proxy; Go has no ambient per-thread session mechanism.
**Suggestion:** Replace with an explicitly-passed database handle (*sql.DB, *sql.Tx, or *gorm.DB) provided via struct field or function parameter.

### `@Transactional annotations` in `src/main/java/com/jtspringproject/JtSpringProject/dao/categoryDao.java`
**Reason:** Spring's proxy-based declarative transaction management has no direct equivalent in Go; transaction boundaries are implicit and enforced by AOP proxies at runtime.
**Suggestion:** Manually rewrite as explicit transaction handling. Move transaction control to the service layer (begin/commit/rollback with defer) and pass the transaction/DB handle into DAO functions, or use a helper like a WithTx(func(tx) error) wrapper.

### `sessionFactory.getCurrentSession()` in `src/main/java/com/jtspringproject/JtSpringProject/dao/categoryDao.java`
**Reason:** Depends on Hibernate's thread-bound session context that is populated by the Spring transaction infrastructure; there is no thread-local session concept in Go.
**Suggestion:** Replace with an explicit *sql.DB / *gorm.DB (or *sql.Tx within a transaction) injected into a repository struct or passed per call.

### `updateCategory (dirty-checking)` in `src/main/java/com/jtspringproject/JtSpringProject/dao/categoryDao.java`
**Reason:** Relies on Hibernate automatically detecting modified entity state and flushing at transaction commit; Go ORMs/SQL do not auto-persist mutated structs.
**Suggestion:** Manually issue an explicit UPDATE (e.g. db.Model(&category).Update('name', name) or an UPDATE SQL statement) after checking existence.

### `@Transactional (all methods)` in `src/main/java/com/jtspringproject/JtSpringProject/dao/cartProductDao.java`
**Reason:** Spring's declarative, proxy-based transaction management has no direct equivalent in Go; transaction boundaries are implicit and thread-bound via AOP proxies.
**Suggestion:** Reimplement explicitly: pass a transaction handle (*sql.Tx or gorm.DB tx) into each repository method, or wrap service-level operations in a transaction closure (e.g. db.Transaction(func(tx *gorm.DB) error {...})). Move transaction control to the service layer.

### `sessionFactory.getCurrentSession()` in `src/main/java/com/jtspringproject/JtSpringProject/dao/cartProductDao.java`
**Reason:** Relies on Hibernate/Spring thread-local session binding tied to the transaction context; no thread-local session model exists in idiomatic Go.
**Suggestion:** Replace with an explicitly injected *sql.DB / *gorm.DB (or the transaction handle) held on the repository struct or passed via context.Context.

### `createNativeQuery(...).setParameterList("product_ids", productIds)` in `src/main/java/com/jtspringproject/JtSpringProject/dao/cartProductDao.java`
**Reason:** Hibernate auto-expands a collection parameter into a SQL IN clause; database/sql does not support list parameters directly.
**Suggestion:** Build the IN clause dynamically with generated placeholders and pass a variadic argument slice, or use a driver-specific array type (e.g. pq.Array for PostgreSQL). Alternatively use gorm's Where("id IN ?", ids).

### `@Transactional annotations` in `src/main/java/com/jtspringproject/JtSpringProject/dao/productDao.java`
**Reason:** Spring's declarative transaction management uses AOP proxies and thread-bound session context. There is no direct Go equivalent — the implicit transaction demarcation and current-session binding cannot be auto-translated.
**Suggestion:** Manually reimplement transactions in Go: use *sql.DB.BeginTx / *gorm.DB.Transaction, or a repository that accepts an injected transaction/context. Wrap each DAO method's body in an explicit begin/commit/rollback, or push transaction control up to a service layer via a WithTx helper.

### `SessionFactory.getCurrentSession()` in `src/main/java/com/jtspringproject/JtSpringProject/dao/productDao.java`
**Reason:** Relies on Hibernate's contextual session mechanism tied to the Spring transaction thread-local. There is no Go analog to the Hibernate persistence context / first-level cache / dirty checking.
**Suggestion:** Replace with an explicit DB connection or transaction handle (database/sql *DB/*Tx or *gorm.DB) passed into or held by the repository. Manually manage entity mapping since automatic dirty checking will not exist.

### `HQL query 'from PRODUCT'` in `src/main/java/com/jtspringproject/JtSpringProject/dao/productDao.java`
**Reason:** HQL operates on entity/class names via Hibernate's ORM mapping, not raw SQL; it cannot be executed directly against a Go SQL driver.
**Suggestion:** Rewrite as native SQL ('SELECT ... FROM products') resolving the real table and column names from the Product entity mapping, or use a Go ORM (GORM) query builder.

### `@Transactional / getCurrentSession()` in `src/main/java/com/jtspringproject/JtSpringProject/dao/userDao.java`
**Reason:** Spring's declarative transaction management uses AOP proxies and thread-local session binding, which has no direct equivalent in Go. The transactional boundary is implicit and cannot be auto-translated.
**Suggestion:** Manual rewrite: introduce an explicit transaction manager or pass a *sql.Tx / *gorm.DB into each method. Wrap service-layer operations in a helper (e.g., WithTransaction(func(tx) error)) so callers control commit/rollback explicitly.

### `HQL query strings ('from CUSTOMER ...')` in `src/main/java/com/jtspringproject/JtSpringProject/dao/userDao.java`
**Reason:** HQL is Hibernate-specific and queries entities by mapped name rather than SQL tables/columns; the string cannot be used verbatim by a Go SQL driver or ORM.
**Suggestion:** Manual rewrite each query as parameterized SQL (or GORM/sqlx query) using the real table and column names derived from the User model mapping. Replace named parameters (:username) with driver placeholders ($1 / ?).

### `saveUser (saveOrUpdate upsert)` in `src/main/java/com/jtspringproject/JtSpringProject/dao/userDao.java`
**Reason:** Hibernate's saveOrUpdate decides insert vs update by inspecting the entity's identifier and persistence state, behavior tied to the ORM session's dirty-checking that Go lacks.
**Suggestion:** Manual rewrite: implement explicit upsert logic — check whether id is zero/present, or use INSERT ... ON CONFLICT (Postgres) / GORM's Save/Clauses(clause.OnConflict). Do not assume automatic dirty tracking.

### `addUser (DataIntegrityViolationException handling)` in `src/main/java/com/jtspringproject/JtSpringProject/services/userService.java`
**Reason:** Relies on Spring/JPA's automatic translation of database constraint violations into a typed exception hierarchy, which has no direct Go equivalent.
**Suggestion:** Manually rewrite: inspect the underlying DB driver error (e.g. pq unique_violation code 23505 or MySQL 1062) and return a wrapped domain error such as ErrDataIntegrity.

### `userService (implicit transaction/session context)` in `src/main/java/com/jtspringproject/JtSpringProject/services/userService.java`
**Reason:** The class assumes Spring-managed transactions and Hibernate session lifecycle around DAO calls; the write-on-read in getUserByUsername is not explicitly transactional in source.
**Suggestion:** Make transaction boundaries explicit in Go using *sql.Tx or a repository that accepts a context/transaction; decide explicitly whether the lazy password migration write should be atomic with the read.

### `adminFilterChain / userFilterChain` in `src/main/java/com/jtspringproject/JtSpringProject/configuration/SecurityConfiguration.java`
**Reason:** These are Spring Security-specific DSL fluent builder configurations with no direct 1:1 equivalent in most target frameworks; the ordering, matcher semantics, and integrated form-login/logout/exception handling are Spring magic.
**Suggestion:** Manually rewrite as the target framework's security/middleware configuration (e.g., Express middleware + Passport, ASP.NET Core authorization policies, or FastAPI dependencies). Explicitly implement route-based authorization ordering and the custom redirect handlers.

### `userDetailsService` in `src/main/java/com/jtspringproject/JtSpringProject/configuration/SecurityConfiguration.java`
**Reason:** Relies on Spring's UserDetailsService contract and Spring's User builder; the .roles() prefix behavior and integration with the filter chains are framework-coupled.
**Suggestion:** Reimplement as a user-lookup + credential-verification function in the target's auth system, explicitly adding the ROLE_ prefix and wiring in an explicit password encoder (bcrypt).

### `refreshAuthenticatedPrincipal` in `src/main/java/com/jtspringproject/JtSpringProject/controller/AdminController.java`
**Reason:** Relies on Spring Security's SecurityContextHolder thread-local and UsernamePasswordAuthenticationToken, which have no direct Go equivalent and depend on Spring's per-request security context lifecycle.
**Suggestion:** Manually rewrite according to the target Go auth mechanism: if using JWT, reissue a token with the new username and return it via cookie/header; if using sessions, update the session store. Do not mechanically translate.

### `SecurityContextHolder usage (index, adminHome, profileDisplay, updateUserProfile)` in `src/main/java/com/jtspringproject/JtSpringProject/controller/AdminController.java`
**Reason:** Global thread-local access to the authenticated principal is a Spring Security idiom with no Go equivalent; Go concurrency uses request-scoped context rather than thread-locals.
**Suggestion:** Introduce auth middleware that injects the authenticated username/user into request context (context.Context) and read from there in each handler.

### `refreshAuthenticatedPrincipal / SecurityContextHolder usage` in `src/main/java/com/jtspringproject/JtSpringProject/controller/UserController.java`
**Reason:** Relies on Spring Security's thread-local SecurityContext and UsernamePasswordAuthenticationToken, which have no direct Go equivalent; the concept of mutating an ambient authentication principal mid-request is Spring-specific.
**Suggestion:** Manually rewrite: obtain the authenticated identity from request-scoped context populated by auth middleware, and after a username update re-issue the session token/JWT or update the session store explicitly rather than mutating a global holder.

### `ModelAndView / view-name returns (indexPage, userLogin, getProducts, profileDisplay, etc.)` in `src/main/java/com/jtspringproject/JtSpringProject/controller/UserController.java`
**Reason:** Depends on Spring MVC's ViewResolver auto-mapping string names to JSP/Thymeleaf templates; this implicit resolution has no automatic Go equivalent.
**Suggestion:** Manual rewrite: set up an explicit html/template registry and render the corresponding template file with the model data map, or convert endpoints to return JSON if migrating to an API + SPA architecture.

## Observer agent findings

The Observer agent monitored the migration and identified these patterns:

- **After 3 modules:** Could not parse observer output
- **After 6 modules:** Could not parse observer output
- **After 9 modules:** Could not parse observer output
- **After 12 modules:** Could not parse observer output
- **After 15 modules:** Could not parse observer output
- **After 18 modules:** Could not parse observer output
- **After 21 modules:** Could not parse observer output
- **After 24 modules:** Could not parse observer output
- **After 27 modules:** Could not parse observer output

## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `pom.xml`
Confidence: 81%

### `src/main/java/com/jtspringproject/JtSpringProject/HibernateConfiguration.java`
Confidence: 13%

### `src/main/java/com/jtspringproject/JtSpringProject/JtSpringProjectApplication.java`
Confidence: 16%

### `src/main/java/com/jtspringproject/JtSpringProject/configuration/PasswordEncoderConfig.java`
Confidence: 22%

### `src/main/java/com/jtspringproject/JtSpringProject/controller/ErrorController.java`
Confidence: 18%

### `src/main/java/com/jtspringproject/JtSpringProject/models/Cart.java`
Confidence: 13%

### `src/main/java/com/jtspringproject/JtSpringProject/models/CartProduct.java`
Confidence: 17%

### `src/main/java/com/jtspringproject/JtSpringProject/models/CartProductId.java`
Confidence: 18%

### `src/main/java/com/jtspringproject/JtSpringProject/models/Category.java`
Confidence: 24%

### `src/main/java/com/jtspringproject/JtSpringProject/models/User.java`
Confidence: 24%

### `src/test/java/com/jtspringproject/JtSpringProject/JtSpringProjectApplicationTests.java`
Confidence: 12%

### `src/test/java/com/jtspringproject/JtSpringProject/models/CartProductIdTest.java`
Confidence: 80%

### `src/test/java/com/jtspringproject/JtSpringProject/models/CartProductTest.java`
Confidence: 13%

### `src/main/java/com/jtspringproject/JtSpringProject/dao/cartDao.java`
Confidence: 53%
Issues:
  - [warning] The INSERT only persists 'DEFAULT VALUES' and does not write any of the Cart entity's actual fields to the database. The original Hibernate save() persists all mapped fields. This is a known placeholder that would produce incorrect persisted rows once the Cart model has real fields.
  - [warning] The UPDATE statement 'UPDATE cart SET id = id WHERE id = ?' is a no-op — it updates no actual fields, so no changes are persisted. The original Hibernate update() writes all mapped field changes. It also does not signal an error if the record does not exist (Hibernate's update on a non-existent entity would fail).
  - [warning] Only the 'id' column is selected and scanned; other Cart fields are not loaded. The original HQL 'from CART' returns fully-hydrated entities. This is a placeholder pending the Cart model.

### `src/main/java/com/jtspringproject/JtSpringProject/dao/categoryDao.java`
Confidence: 18%

### `src/main/java/com/jtspringproject/JtSpringProject/dao/cartProductDao.java`
Confidence: 12%

### `src/main/java/com/jtspringproject/JtSpringProject/dao/productDao.java`
Confidence: 17%

### `src/main/java/com/jtspringproject/JtSpringProject/dao/userDao.java`
Confidence: 17%

### `src/main/java/com/jtspringproject/JtSpringProject/services/cartService.java`
Confidence: 13%

### `src/main/java/com/jtspringproject/JtSpringProject/services/userService.java`
Confidence: 13%

### `src/main/java/com/jtspringproject/JtSpringProject/configuration/SecurityConfiguration.java`
Confidence: 76%

### `src/main/java/com/jtspringproject/JtSpringProject/controller/AdminController.java`
Confidence: 68%
Issues:
  - [info] The buildProduct helper is truncated at the end of the provided file (`func (c *AdminController)` with no body). Cannot fully verify the product field mapping (name, category resolution, price, weight, quantity, description, image) is preserved. If actually complete in the real file, this is a non-issue.

### `src/main/java/com/jtspringproject/JtSpringProject/controller/UserController.java`
Confidence: 82%
Issues:
  - [info] The Java version relies on SecurityContext having an authenticated principal (throws NPE when absent). The Go version treats missing authentication as 'User not found' and renders the updateProfile view. This is a minor behavioral difference on the unauthenticated edge case, but arguably more graceful and acceptable in a migration.
  - [info] The Java refreshAuthenticatedPrincipal preserves original credentials and authorities while updating the username. The Go migration replaces this with an optional Reauthenticate hook that is only invoked if the principalProvider happens to implement the interface; the default wiring does not refresh authentication. This is an expected/documented gap for the Spring Security thread-local behavior with no clean Go equivalent.
