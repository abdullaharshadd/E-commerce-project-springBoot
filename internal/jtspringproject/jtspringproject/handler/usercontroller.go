// Package handler contains the Go migration of the JtSpringProject web
// layer (the com.jtspringproject.JtSpringProject.controller package).
//
// This file migrates UserController.java, a Spring MVC @Controller that
// handled the user-facing web routes: registration, login page, home/index,
// product listing, profile display, and profile update. In the Java world it
// rendered server-side view templates (returning view-name strings or
// ModelAndView objects) and was backed by two service beans (userService,
// productService) injected via constructor. It reached into Spring Security's
// SecurityContextHolder thread-local to obtain the authenticated principal and
// to re-authenticate the current user after a profile update.
//
// Migration decisions:
//
//   - Spring MVC view rendering (ModelAndView / "redirect:..." view names) has
//     been replaced with the standard library html/template renderer plus
//     net/http helpers. Each handler renders a named template with a data map
//     analogous to the Java Model attributes.
//   - Constructor injection (@Autowired) becomes an explicit NewUserController
//     constructor taking the service dependencies as interfaces.
//   - @GetMapping / @PostMapping become methods registered against an
//     http.ServeMux in RegisterRoutes.
//   - @RequestParam / @ModelAttribute become explicit reads from the parsed
//     request form/query values.
//   - The SecurityContextHolder thread-local is not directly representable in
//     Go. The authenticated principal is read from the request context via the
//     config.UserProvider abstraction. Re-authenticating the principal after a
//     profile update (mutating the SecurityContext) has no clean equivalent in
//     a stateless handler; see the MIGRATION_NOTE in updateUserProfile.
package handler

import (
	"context"
	"html/template"
	"net/http"
	"strconv"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/config"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/models"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/service"
)

// productLister is the subset of the product service used by UserController.
// Depending on an interface (rather than the concrete *service.ProductService)
// keeps the handler testable.
type productLister interface {
	GetProducts(ctx context.Context) ([]models.Product, error)
}

// userManager is the subset of the user service used by UserController.
type userManager interface {
	CheckUserExists(ctx context.Context, username string) (bool, error)
	AddUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID int, username, email, password, address string) (*models.User, error)
}

// principalProvider resolves the currently authenticated principal for a
// request. It stands in for Spring Security's SecurityContextHolder.
type principalProvider interface {
	// CurrentUsername returns the authenticated username for the request and
	// whether an authenticated principal is present.
	CurrentUsername(r *http.Request) (string, bool)
}

// Renderer renders a named server-side template with the supplied data. It
// abstracts the html/template execution so handlers do not depend on a
// concrete template store.
type Renderer interface {
	// Render writes the named template to w using data. It returns an error if
	// the template cannot be found or executed.
	Render(w http.ResponseWriter, name string, data map[string]any) error
}

// UserController handles the user-facing web routes. It is the Go migration of
// the Spring MVC UserController.
type UserController struct {
	userService    userManager
	productService productLister
	principals     principalProvider
	renderer       Renderer
}

// NewUserController constructs a UserController with its dependencies. It is
// the Go equivalent of the @Autowired constructor injection used by the Java
// class.
func NewUserController(
	userService userManager,
	productService productLister,
	principals principalProvider,
	renderer Renderer,
) *UserController {
	return &UserController{
		userService:    userService,
		productService: productService,
		principals:     principals,
		renderer:       renderer,
	}
}

// Compile-time assertion that the concrete services satisfy the handler's
// interfaces. These document the intended production wiring.
var (
	_ userManager    = (*service.UserService)(nil)
	_ productLister  = (*service.ProductService)(nil)
)

// RegisterRoutes wires the controller's handlers onto the supplied mux. It
// mirrors the @GetMapping/@PostMapping annotations on the Java controller.
func (c *UserController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /register", c.RegisterUser)
	mux.HandleFunc("GET /buy", c.Buy)
	mux.HandleFunc("GET /login", c.UserLogin)
	mux.HandleFunc("GET /", c.IndexPage)
	mux.HandleFunc("GET /user/products", c.GetProducts)
	mux.HandleFunc("POST /newuserregister", c.RegisterNewUser)
	mux.HandleFunc("GET /profileDisplay", c.ProfileDisplay)
	mux.HandleFunc("POST /updateuser", c.UpdateUserProfile)
}

// RegisterUser renders the registration page (GET /register).
func (c *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	c.render(w, "register", nil)
}

// Buy renders the buy page (GET /buy).
func (c *UserController) Buy(w http.ResponseWriter, r *http.Request) {
	c.render(w, "buy", nil)
}

// UserLogin renders the login page (GET /login). When the request carries
// error=true it surfaces a validation message, matching the Java behaviour.
func (c *UserController) UserLogin(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	if r.URL.Query().Get("error") == "true" {
		data["msg"] = "Please enter correct email and password"
	}
	c.render(w, "userLogin", data)
}

// IndexPage renders the home page (GET /). It exposes the authenticated
// username and the list of available products, or a message when none exist.
func (c *UserController) IndexPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data := map[string]any{}
	if username, ok := c.principals.CurrentUsername(r); ok {
		data["username"] = username
	}

	products, err := c.productService.GetProducts(ctx)
	if err != nil {
		http.Error(w, "failed to load products", http.StatusInternalServerError)
		return
	}

	if len(products) == 0 {
		data["msg"] = "No products are available"
	} else {
		data["products"] = products
	}

	c.render(w, "index", data)
}

// GetProducts renders the user-facing product listing (GET /user/products).
func (c *UserController) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	products, err := c.productService.GetProducts(ctx)
	if err != nil {
		http.Error(w, "failed to load products", http.StatusInternalServerError)
		return
	}

	data := map[string]any{}
	if len(products) == 0 {
		data["msg"] = "No products are available"
	} else {
		data["products"] = products
	}

	c.render(w, "uproduct", data)
}

// RegisterNewUser handles new-user registration (POST /newuserregister). If the
// chosen username is free the user is created with the ROLE_NORMAL role and the
// login page is rendered; otherwise the registration page is re-rendered with a
// "username taken" message.
func (c *UserController) RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")

	exists, err := c.userService.CheckUserExists(ctx, username)
	if err != nil {
		http.Error(w, "failed to check user", http.StatusInternalServerError)
		return
	}

	if exists {
		c.render(w, "register", map[string]any{
			"msg": username + " is taken. Please choose a different username.",
		})
		return
	}

	user := &models.User{
		Username: username,
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Address:  r.FormValue("address"),
		Role:     "ROLE_NORMAL",
	}

	if _, err := c.userService.AddUser(ctx, user); err != nil {
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	c.render(w, "userLogin", nil)
}

// ProfileDisplay renders the profile update form for the authenticated user
// (GET /profileDisplay). The password field is intentionally rendered blank,
// matching the Java behaviour.
func (c *UserController) ProfileDisplay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username, ok := c.principals.CurrentUsername(r)
	if !ok {
		c.render(w, "updateProfile", map[string]any{"msg": "User not found"})
		return
	}

	user, err := c.userService.GetUserByUsername(ctx, username)
	if err != nil {
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	if user == nil {
		c.render(w, "updateProfile", map[string]any{"msg": "User not found"})
		return
	}

	c.render(w, "updateProfile", map[string]any{
		"userid":   user.ID,
		"username": user.Username,
		"email":    user.Email,
		"password": "",
		"address":  user.Address,
	})
}

// UpdateUserProfile applies a profile update from the submitted form
// (POST /updateuser) and redirects to the home page.
//
// MIGRATION_NOTE: The Java version called refreshAuthenticatedPrincipal after a
// successful update, mutating Spring Security's SecurityContext so the new
// username took effect on the current authentication. Go handlers are
// stateless and there is no thread-local SecurityContext to mutate. The
// principalProvider abstraction below exposes an optional Reauthenticate hook;
// if the wired implementation supports it we invoke it, otherwise the redirect
// proceeds and the session must be refreshed by whatever auth mechanism is in
// use. This requires manual review to match your session/JWT strategy.
func (c *UserController) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("userid"))
	if err != nil {
		http.Error(w, "invalid userid", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	address := r.FormValue("address")

	updatedUser, err := c.userService.UpdateUserProfile(ctx, userID, username, email, password, address)
	if err != nil {
		http.Error(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	if updatedUser != nil {
		if reauth, ok := c.principals.(interface {
			Reauthenticate(w http.ResponseWriter, r *http.Request, username string) error
		}); ok {
			if err := reauth.Reauthenticate(w, r, username); err != nil {
				http.Error(w, "failed to refresh authentication", http.StatusInternalServerError)
				return
			}
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// render executes a named template, translating any failure into a 500
// response. Centralising rendering keeps the handlers free of repetitive error
// handling.
func (c *UserController) render(w http.ResponseWriter, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	if err := c.renderer.Render(w, name, data); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
	}
}

// Ensure html/template is referenced so a template-backed Renderer
// implementation elsewhere in the package compiles against a consistent type.
var _ = template.HTMLEscapeString

// principalFromConfig adapts config.UserProvider-based authentication into the
// principalProvider expected by UserController. It is provided as a starting
// point; the exact mechanism for locating the authenticated username on a
// request (session cookie, JWT claim, etc.) must be supplied during wiring.
//
// MIGRATION_NOTE: config.AuthenticateByUsername / config.UserProvider replace
// the UserDetailsService, but there is no framework-managed request-scoped
// principal in the migrated stack. Provide a concrete implementation that reads
// the authenticated username from your chosen transport.
type principalFromConfig struct {
	provider config.UserProvider
}

var _ config.UserProvider = (config.UserProvider)(nil)
