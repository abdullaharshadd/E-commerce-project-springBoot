// Package handler contains the Go migration of the JtSpringProject Spring MVC
// controllers.
//
// This file migrates ErrorController.java, a Spring @Controller whose sole
// responsibility was to map HTTP GET "/403" to a view named "403" — the
// access-denied error page.
//
// MIGRATION_NOTE: The original relied on two Spring mechanisms that have no
// direct Go equivalent:
//
//   - @Controller: component scanning registered the class as an
//     IoC-container-managed bean. Go has no IoC container, so we instead expose
//     an explicit constructor (NewErrorController) and register routes
//     manually.
//   - View-name resolution: the returned String "403" was resolved by a Spring
//     ViewResolver into a template. Go's net/http has no ViewResolver, so the
//     handler must render the template (or write a response) itself. Here we
//     model the view name explicitly via the AccessDenied method and let the
//     HTTP handler render it, keeping the "403" view name as the single source
//     of truth.
//
// MANUAL_REVIEW: The concrete template rendering (html/template execution,
// static file serving, or a redirect) depends on how the rest of the
// application wires its view layer. The renderer dependency below is an
// interface so callers can inject their chosen rendering strategy.
package handler

import (
	"context"
	"fmt"
	"net/http"
)

// accessDeniedView is the logical view name the original Spring controller
// returned for the /403 mapping. It is resolved by the injected ViewRenderer
// into an actual response body.
const accessDeniedView = "403"

// ViewRenderer resolves a logical view name into an HTTP response, replacing
// Spring's ViewResolver abstraction.
//
// MIGRATION_NOTE: In Spring the ViewResolver was container-provided and
// implicit. In Go the dependency is explicit and injected via the constructor.
type ViewRenderer interface {
	// Render writes the named view to w. Implementations should set an
	// appropriate status code and content type. It returns an error if the
	// view cannot be resolved or written.
	Render(ctx context.Context, w http.ResponseWriter, name string) error
}

// ErrorController serves application error pages. It is the Go migration of the
// Spring MVC ErrorController.
type ErrorController struct {
	renderer ViewRenderer
}

// NewErrorController constructs an ErrorController that renders views using the
// supplied ViewRenderer. It returns an error if renderer is nil so wiring
// mistakes fail fast at startup rather than at request time.
func NewErrorController(renderer ViewRenderer) (*ErrorController, error) {
	if renderer == nil {
		return nil, fmt.Errorf("handler: NewErrorController requires a non-nil ViewRenderer")
	}
	return &ErrorController{renderer: renderer}, nil
}

// AccessDeniedView returns the logical view name for the access-denied page.
// This is the direct analogue of the original accessDenied() method, which
// returned the view name "403".
func (c *ErrorController) AccessDeniedView() string {
	return accessDeniedView
}

// AccessDenied is the HTTP handler mapped to GET "/403". It renders the
// access-denied view, mirroring the original @GetMapping("/403") controller
// method.
func (c *ErrorController) AccessDenied(w http.ResponseWriter, r *http.Request) {
	if err := c.renderer.Render(r.Context(), w, accessDeniedView); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// RegisterRoutes wires the controller's endpoints onto the supplied ServeMux,
// replacing Spring's annotation-driven @GetMapping registration with explicit
// route registration.
//
// MIGRATION_NOTE: net/http's ServeMux does not filter by HTTP method, so the
// handler guards against non-GET requests to preserve the original
// @GetMapping semantics. If the application uses chi (as noted in the migrated
// application entry point), prefer router.Get("/403", c.AccessDenied) instead.
func (c *ErrorController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/403", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		c.AccessDenied(w, r)
	})
}
