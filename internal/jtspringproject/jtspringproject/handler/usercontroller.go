// Package handler contains the HTTP handlers for the application's server-
// rendered pages and form-post actions.
//
// MIGRATION_NOTE: The original UserController was a Spring @Controller that
// resolved view names (e.g. "register", "userLogin", "index") to templates and
// used the "redirect:/" convention for the post/redirect/get pattern. In Go we
// render templates explicitly with html/template and issue redirects with
// http.Redirect. Route registration is explicit via RegisterRoutes rather than
// annotations.
//
// MIGRATION_NOTE: Spring Security's SecurityContextHolder is a thread-local
// holding the authenticated principal. Go has no thread-local; the
// authenticated user is carried on the request context via the session
// middleware from the config package. To "refresh" the authenticated principal
// after a profile update, we mutate the session-backed username rather than
// re-creating an Authentication token.
package handler

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/config"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/service"
)

// UserController renders the user-facing web pages (registration, login, index,
// product listing and profile management) and handles their form submissions.
//
// MIGRATION_NOTE: The Spring controller received its dependencies via
// constructor injection (@Autowired). Here we use an explicit constructor
// (NewUserController) and depend on interfaces rather than concrete types where
// practical.
type UserController struct {
	userService    service.UserRepository
	productService service.ProductRepository
	sessions       *config.SessionStore
	templates      *template.Template
}

// NewUserController constructs a UserController with its required
// dependencies. The sessions store is used to read and refresh the currently
// authenticated principal, replacing Spring's SecurityContextHolder. The
// templates are the parsed html/template set used to render the server-side
// views.
func NewUserController(
	userService service.UserRepository,
	productService service.ProductRepository,
	sessions *config.SessionStore,
	templates *template.Template,
) *UserController {
	return &UserController{
		userService:    userService,
		productService: productService,
		sessions:       sessions,
		templates:      templates,
	}
}

// RegisterRoutes registers every route owned by the UserController on the
// provided mux. Routes requiring authentication are wrapped with the supplied
// auth middleware (which corresponds to Spring Security's role-based filter
// chain, e.g. config.RequireRole).
//
// MIGRATION_NOTE: Spring registered these routes declaratively via
// @GetMapping/@PostMapping annotations. Go's net/http mux dispatches on path
// only, so we guard the HTTP method inside each handler and return 405 for
// mismatches.
func (c *UserController) RegisterRoutes(mux *http.ServeMux, auth func(http.Handler) http.Handler) {
	// Public routes.
	mux.HandleFunc("/register", c.RegisterUser)
	mux.HandleFunc("/login", c.UserLogin)
	mux.HandleFunc("/newuserregister", c.RegisterNewUser)

	// Authenticated routes.
	mux.Handle("/buy", auth(http.HandlerFunc(c.Buy)))
	mux.Handle("/", auth(http.HandlerFunc(c.IndexPage)))
	mux.Handle("/user/products", auth(http.HandlerFunc(c.GetProducts)))
	mux.Handle("/profileDisplay", auth(http.HandlerFunc(c.ProfileDisplay)))
	mux.Handle("/updateuser", auth(http.HandlerFunc(c.UpdateUserProfile)))
}

// render executes a named template with the given data and writes it to the
// response. Any template execution error results in a 500.
func (c *UserController) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
	}
}

// authenticatedUsername returns the username of the currently authenticated
// principal, replacing Spring's SecurityContextHolder lookup. It relies on the
// session middleware having placed the session on the request context.
//
// MIGRATION_NOTE: The exact context key and session field must match the
// implementation in config.SecurityConfiguration. Adjust if the session
// middleware stores the username differently.
func (c *UserController) authenticatedUsername(r *http.Request) (string, bool) {
	sess, ok := config.SessionFromContext(r.Context())
	if !ok {
		return "", false
	}
	if sess.Username == "" {
		return "", false
	}
	return sess.Username, true
}

// RegisterUser shows the user registration form.
//
// GET /register
func (c *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c.render(w, "register", nil)
}

// Buy shows the buy page. Authentication is enforced by the middleware in
// RegisterRoutes.
//
// GET /buy
func (c *UserController) Buy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c.render(w, "buy", nil)
}

// UserLogin shows the login form. When invoked with ?error=true it also passes
// an error message to the view.
//
// GET /login
func (c *UserController) UserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data := map[string]any{}
	if r.URL.Query().Get("error") == "true" {
		data["msg"] = "Please enter correct email and password"
	}
	c.render(w, "userLogin", data)
}

// IndexPage shows the index page with the authenticated username and the full
// product list. When no products exist it passes an informational message
// instead.
//
// GET /
func (c *UserController) IndexPage(w http.ResponseWriter, r *http.Request) {
	// The mux routes any otherwise-unmatched path to "/"; restrict this handler
	// to the exact root path to preserve the original single-route behaviour.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, _ := c.authenticatedUsername(r)
	data := map[string]any{"username": username}

	c.attachProducts(r.Context(), w, data)
}

// GetProducts shows the product listing page for users.
//
// GET /user/products
func (c *UserController) GetProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data := map[string]any{}
	if !c.loadProducts(r.Context(), w, data) {
		return
	}
	c.render(w, "uproduct", data)
}

// attachProducts loads the products into data and renders the index template.
func (c *UserController) attachProducts(ctx context.Context, w http.ResponseWriter, data map[string]any) {
	if !c.loadProducts(ctx, w, data) {
		return
	}
	c.render(w, "index", data)
}

// loadProducts fetches the product list and populates data with either the
// products or a "no products" message. It returns false (after writing an error
// response) if the fetch fails.
func (c *UserController) loadProducts(ctx context.Context, w http.ResponseWriter, data map[string]any) bool {
	products, err := c.productService.GetProducts(ctx)
	if err != nil {
		http.Error(w, "failed to load products", http.StatusInternalServerError)
		return false
	}
	if len(products) == 0 {
		data["msg"] = "No products are available"
	} else {
		data["products"] = products
	}
	return true
}

// RegisterNewUser registers a new user with the ROLE_NORMAL role if the chosen
// username is available. On success it renders the login view; when the
// username is taken it re-renders the registration form with a message.
//
// POST /newuserregister
func (c *UserController) RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	// MIGRATION_NOTE: @ModelAttribute User bound the whole form onto a User
	// object. We bind the equivalent fields explicitly here.
	user := &model.User{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Email:    r.FormValue("email"),
		Address:  r.FormValue("address"),
	}

	exists, err := c.userService.CheckUserExists(r.Context(), user.Username)
	if err != nil {
		http.Error(w, "failed to check user", http.StatusInternalServerError)
		return
	}

	if exists {
		c.render(w, "register", map[string]any{
			"msg": user.Username + " is taken. Please choose a different username.",
		})
		return
	}

	user.Role = "ROLE_NORMAL"
	if _, err := c.userService.AddUser(r.Context(), user); err != nil {
		if errors.Is(err, service.ErrDataIntegrity) {
			c.render(w, "register", map[string]any{
				"msg": user.Username + " is taken. Please choose a different username.",
			})
			return
		}
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	c.render(w, "userLogin", nil)
}

// ProfileDisplay displays the profile edit form for the currently authenticated
// user.
//
// GET /profileDisplay
func (c *UserController) ProfileDisplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, ok := c.authenticatedUsername(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := c.userService.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, config.ErrUserNotFound) {
			c.render(w, "updateProfile", map[string]any{"msg": "User not found"})
			return
		}
		http.Error(w, "failed to load profile", http.StatusInternalServerError)
		return
	}
	if user == nil {
		c.render(w, "updateProfile", map[string]any{"msg": "User not found"})
		return
	}

	// MIGRATION_NOTE: The password field is intentionally rendered blank, matching
	// the original controller which never echoed the stored password back.
	c.render(w, "updateProfile", map[string]any{
		"userid":   user.ID,
		"username": user.Username,
		"email":    user.Email,
		"password": "",
		"address":  user.Address,
	})
}

// UpdateUserProfile updates the user profile and, on success, refreshes the
// authenticated principal so that subsequent requests reflect the new username.
// It then redirects to the index page (post/redirect/get).
//
// POST /updateuser
func (c *UserController) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

	updatedUser, err := c.userService.UpdateUserProfile(r.Context(), userID, username, email, password, address)
	if err != nil {
		http.Error(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	if updatedUser != nil {
		if err := c.refreshAuthenticatedPrincipal(w, r, username); err != nil {
			http.Error(w, "failed to refresh session", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// refreshAuthenticatedPrincipal updates the authenticated principal's username
// on the current session.
//
// MIGRATION_NOTE: The original code rebuilt a UsernamePasswordAuthenticationToken
// carrying the existing credentials and authorities, and replaced it on the
// SecurityContextHolder. Go's session-based model has no such token; we instead
// update the username stored on the session so the change persists across
// requests. Credentials and authorities are not stored client-side, so nothing
// else needs copying. Verify the SessionStore API matches this usage during
// manual review.
func (c *UserController) refreshAuthenticatedPrincipal(w http.ResponseWriter, r *http.Request, username string) error {
	sess, ok := config.SessionFromContext(r.Context())
	if !ok {
		return errors.New("no active session to refresh")
	}
	sess.Username = username
	return c.sessions.Save(w, r, sess)
}
