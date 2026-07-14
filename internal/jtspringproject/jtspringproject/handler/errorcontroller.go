// Package handler contains the migrated HTTP handlers for the original
// Spring MVC controllers.
//
// This file migrates ErrorController.java. The original class was a Spring MVC
// @Controller that mapped GET /403 to a view named "403" (resolved by a
// ViewResolver to a template such as 403.jsp / 403.html). It served as the
// access-denied landing page, typically reached when Spring Security rejected
// an unauthorized request.
//
// MIGRATION_NOTE: Go's standard library has no ViewResolver / JSP concept. The
// idiomatic replacement is an http.Handler that renders an HTML template using
// html/template. Since the original template file (403.jsp/403.html) is not
// part of this Java source, we embed a minimal, self-contained 403 page here.
// If the real template exists, replace accessDeniedTemplate with the migrated
// version (loaded via embed.FS or template.ParseFiles) — see MANUAL REVIEW note.
package handler

import (
	"html/template"
	"net/http"
)

// accessDeniedTemplate is the HTML rendered for the /403 access-denied page.
//
// MIGRATION_NOTE: This is a placeholder standing in for the original Spring
// view named "403". Swap it for the migrated template content if available.
var accessDeniedTemplate = template.Must(template.New("403").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>403 - Access Denied</title>
</head>
<body>
	<h1>403 - Access Denied</h1>
	<p>You do not have permission to access this resource.</p>
</body>
</html>`))

// ErrorController serves error-related pages, migrated from the Spring MVC
// ErrorController. It is constructed via NewErrorController and its routes are
// registered explicitly through RegisterRoutes.
type ErrorController struct {
	tmpl *template.Template
}

// NewErrorController constructs an ErrorController ready to serve its routes.
func NewErrorController() *ErrorController {
	return &ErrorController{tmpl: accessDeniedTemplate}
}

// RegisterRoutes registers all HTTP routes owned by the ErrorController on the
// provided ServeMux. It mirrors the original @GetMapping declarations.
//
// Registered routes:
//
//	GET /403 -> AccessDenied (renders the access-denied view)
func (c *ErrorController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /403", c.AccessDenied)
}

// AccessDenied handles GET /403 and renders the access-denied page.
//
// MIGRATION_NOTE: The original method returned the String "403", which Spring's
// ViewResolver turned into an HTTP 200 response rendering the 403 template. The
// view name did NOT set the HTTP status code, so this handler preserves that
// behaviour by responding with 200 OK and the rendered HTML body. Adjust the
// status code to http.StatusForbidden here if the intended semantics are to
// return an actual 403 status.
func (c *ErrorController) AccessDenied(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.tmpl.Execute(w, nil); err != nil {
		http.Error(w, "failed to render access-denied page", http.StatusInternalServerError)
		return
	}
}
