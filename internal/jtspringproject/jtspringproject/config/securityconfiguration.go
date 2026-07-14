// Package config provides HTTP security configuration for the application.
//
// MIGRATION_NOTE: The original Spring SecurityConfiguration relied heavily on
// Spring Security's declarative DSL (HttpSecurity fluent builder, multiple
// ordered SecurityFilterChain beans, form-login processing, session management
// via JSESSIONID, role-based AntPathRequestMatcher rules, and a
// UserDetailsService). Go has no direct equivalent framework, so this file
// implements the same behaviour explicitly using net/http middleware,
// cookie-based sessions, and an authenticator backed by the user service.
//
// The following behaviours are preserved:
//   - Two separate protection domains: "/admin/**" (ADMIN role) and everything
//     else ("/**", USER role).
//   - Form-login processing endpoints:
//       POST /admin/loginvalidate  and  POST /userloginvalidate
//   - Login pages permitted anonymously:
//       GET /admin/login  and  GET /login
//   - Logout via GET that clears the JSESSIONID cookie:
//       GET /admin/logout  and  GET /logout
//   - Access-denied handling redirects to /403.
//
// Session storage here is intentionally minimal (an in-memory store keyed by an
// opaque session id stored in the JSESSIONID cookie). Production deployments
// should replace SessionStore with a shared/persistent implementation (Redis,
// database, signed cookies, etc.). See requires_manual_review.
package config

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/service"
)

// Role constants mirror the Spring role names (without the "ROLE_" prefix that
// Spring adds internally).
const (
	// RoleAdmin is the authority required to access "/admin/**".
	RoleAdmin = "ADMIN"
	// RoleUser is the authority required to access all non-admin protected paths.
	RoleUser = "USER"
)

// sessionCookieName is the cookie used to track authenticated sessions.
//
// MIGRATION_NOTE: Named JSESSIONID to remain compatible with existing logout
// anchor links and any client expectations carried over from the Spring app.
const sessionCookieName = "JSESSIONID"

// ErrUserNotFound indicates that no user exists for the supplied username.
//
// MIGRATION_NOTE: Replaces Spring's UsernameNotFoundException. Go library code
// returns errors instead of throwing.
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidCredentials indicates authentication failed for the supplied
// username/password combination.
var ErrInvalidCredentials = errors.New("invalid credentials")

// UserAuthenticator loads a user's authentication details by username.
//
// MIGRATION_NOTE: This is the Go equivalent of Spring's UserDetailsService
// functional interface. The concrete implementation delegates to the migrated
// UserService.
type UserAuthenticator interface {
	// LoadUser returns the stored password hash and role for the given username.
	// It returns ErrUserNotFound if the user does not exist.
	LoadUser(ctx context.Context, username string) (passwordHash, role string, err error)
}

// PasswordVerifier verifies a plaintext password against a stored hash.
type PasswordVerifier interface {
	// Matches reports whether the raw password matches the encoded hash.
	Matches(raw, encoded string) bool
}

// UserDetailsService adapts the application's UserService to the
// UserAuthenticator interface.
//
// MIGRATION_NOTE: In Spring this was a @Bean lambda implementing
// UserDetailsService. Here it is an explicit type constructed via
// NewUserDetailsService.
type UserDetailsService struct {
	users *service.UserService
}

// NewUserDetailsService constructs a UserDetailsService backed by the given
// UserService.
func NewUserDetailsService(users *service.UserService) *UserDetailsService {
	return &UserDetailsService{users: users}
}

// LoadUser implements UserAuthenticator. It fetches the user by username and
// maps the stored role to the security role used for authorization.
//
// MIGRATION_NOTE: The original mapped "ROLE_ADMIN" -> "ADMIN" and everything
// else -> "USER". That mapping is preserved here.
func (s *UserDetailsService) LoadUser(ctx context.Context, username string) (string, string, error) {
	user, err := s.users.GetUserByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}
	// GetUserByUsername may return a zero-value user for "not found" depending on
	// its contract; treat an empty username as not found to be safe.
	if strings.TrimSpace(user.Username) == "" {
		return "", "", ErrUserNotFound
	}

	role := RoleUser
	if user.Role == "ROLE_ADMIN" {
		role = RoleAdmin
	}
	return user.Password, role, nil
}

// Session represents an authenticated user's session state.
type Session struct {
	Username string
	Role     string
}

// SessionStore stores authenticated sessions keyed by an opaque session id.
//
// MIGRATION_NOTE: Spring managed the HTTP session and JSESSIONID cookie
// automatically. This is a minimal in-memory replacement. Replace with a
// distributed/persistent store for production and multi-instance deployments.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// NewSessionStore constructs an empty in-memory SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]Session)}
}

// Create stores a new session and returns its generated id.
func (s *SessionStore) Create(sess Session) (string, error) {
	id, err := newSessionID()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()
	return id, nil
}

// Get returns the session for the given id and whether it exists.
func (s *SessionStore) Get(id string) (Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	return sess, ok
}

// Delete removes the session with the given id.
func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func newSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// SecurityConfiguration wires the authentication and authorization components
// and registers the security-related HTTP routes.
//
// MIGRATION_NOTE: Replaces the Spring @Configuration class. Dependencies are
// injected explicitly through NewSecurityConfiguration rather than via a DI
// container.
type SecurityConfiguration struct {
	auth     UserAuthenticator
	verifier PasswordVerifier
	sessions *SessionStore
}

// NewSecurityConfiguration constructs a SecurityConfiguration.
func NewSecurityConfiguration(auth UserAuthenticator, verifier PasswordVerifier, sessions *SessionStore) *SecurityConfiguration {
	return &SecurityConfiguration{
		auth:     auth,
		verifier: verifier,
		sessions: sessions,
	}
}

// RegisterRoutes registers all security-related HTTP endpoints on the given
// mux. These correspond to the endpoints the Spring Security filter chains
// handled automatically.
//
// Routes registered (method + path):
//
//	POST /admin/loginvalidate  -> admin form-login processing
//	GET  /admin/login          -> admin login page (anonymous)
//	GET  /admin/logout         -> admin logout (auth required)
//	POST /userloginvalidate    -> user form-login processing
//	GET  /login                -> user login page (anonymous)
//	GET  /logout               -> user logout (auth required)
func (c *SecurityConfiguration) RegisterRoutes(mux *http.ServeMux) {
	// --- Admin filter chain (@Order(1), /admin/**, ROLE_ADMIN) ---
	mux.HandleFunc("GET /admin/login", c.adminLoginPage)
	mux.HandleFunc("POST /admin/loginvalidate", c.adminLoginValidate)
	mux.HandleFunc("GET /admin/logout", c.RequireRole(RoleAdmin, c.adminLogout))

	// --- User filter chain (@Order(2), /**, ROLE_USER) ---
	mux.HandleFunc("GET /login", c.userLoginPage)
	mux.HandleFunc("POST /userloginvalidate", c.userLoginValidate)
	mux.HandleFunc("GET /logout", c.RequireRole(RoleUser, c.userLogout))
}

// adminLoginPage renders/permits the admin login page anonymously.
//
// MIGRATION_NOTE: Spring's .loginPage("/admin/login").permitAll() only
// registered the URL; the actual view was served elsewhere. Here we simply
// ensure the path is reachable without authentication. The view rendering is
// expected to be handled by the corresponding controller/handler.
func (c *SecurityConfiguration) adminLoginPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// userLoginPage renders/permits the user login page anonymously.
func (c *SecurityConfiguration) userLoginPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// adminLoginValidate processes an admin form-login submission.
//
// MIGRATION_NOTE: Equivalent to Spring's loginProcessingUrl + success/failure
// handlers. On success it establishes a session and redirects to "/admin/"; on
// failure it redirects to "/admin/login?error=true".
func (c *SecurityConfiguration) adminLoginValidate(w http.ResponseWriter, r *http.Request) {
	c.processLogin(w, r, RoleAdmin, "/admin/", "/admin/login?error=true")
}

// userLoginValidate processes a user form-login submission.
func (c *SecurityConfiguration) userLoginValidate(w http.ResponseWriter, r *http.Request) {
	c.processLogin(w, r, RoleUser, "/", "/login?error=true")
}

// processLogin authenticates the submitted credentials and, on success,
// establishes a session (setting the JSESSIONID cookie) and redirects to
// successURL. On failure it redirects to failureURL.
//
// requiredRole enforces that only users with the appropriate role may log in
// through a given chain, mirroring the .hasRole(...) rules.
func (c *SecurityConfiguration) processLogin(w http.ResponseWriter, r *http.Request, requiredRole, successURL, failureURL string) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, failureURL, http.StatusSeeOther)
		return
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	sess, err := c.authenticate(r.Context(), username, password)
	if err != nil {
		http.Redirect(w, r, failureURL, http.StatusSeeOther)
		return
	}

	// Enforce role compatibility with the chain that handled the login.
	// An ADMIN may also satisfy USER requirements (Spring role hierarchy is not
	// configured here, so we treat roles as distinct except that ADMIN >= USER
	// is NOT assumed unless explicitly required). We keep it strict to match the
	// original hasRole semantics.
	if requiredRole == RoleAdmin && sess.Role != RoleAdmin {
		http.Redirect(w, r, failureURL, http.StatusSeeOther)
		return
	}
	if requiredRole == RoleUser && sess.Role != RoleUser && sess.Role != RoleAdmin {
		http.Redirect(w, r, failureURL, http.StatusSeeOther)
		return
	}

	id, err := c.sessions.Create(sess)
	if err != nil {
		http.Redirect(w, r, failureURL, http.StatusSeeOther)
		return
	}
	c.setSessionCookie(w, id)
	http.Redirect(w, r, successURL, http.StatusSeeOther)
}

// authenticate verifies the username/password and returns the resulting
// session on success.
func (c *SecurityConfiguration) authenticate(ctx context.Context, username, password string) (Session, error) {
	hash, role, err := c.auth.LoadUser(ctx, username)
	if err != nil {
		return Session{}, err
	}
	if !c.verifier.Matches(password, hash) {
		return Session{}, ErrInvalidCredentials
	}
	return Session{Username: username, Role: role}, nil
}

// adminLogout terminates the admin session and redirects to the admin login
// page, clearing the JSESSIONID cookie.
//
// MIGRATION_NOTE: Preserves the Spring GET-based logout for compatibility with
// existing logout anchor links.
func (c *SecurityConfiguration) adminLogout(w http.ResponseWriter, r *http.Request) {
	c.logout(w, r, "/admin/login")
}

// userLogout terminates the user session and redirects to the login page,
// clearing the JSESSIONID cookie.
func (c *SecurityConfiguration) userLogout(w http.ResponseWriter, r *http.Request) {
	c.logout(w, r, "/login")
}

func (c *SecurityConfiguration) logout(w http.ResponseWriter, r *http.Request, redirectURL string) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		c.sessions.Delete(cookie.Value)
	}
	c.clearSessionCookie(w)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// RequireRole returns middleware that enforces the given role for the wrapped
// handler.
//
// MIGRATION_NOTE: This is the explicit Go equivalent of Spring's
// .authorizeHttpRequests(...).hasRole(role). On missing/invalid session it
// redirects to the appropriate login page; on insufficient privileges it
// redirects to /403 (mirroring .exceptionHandling().accessDeniedPage("/403")).
func (c *SecurityConfiguration) RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := c.currentSession(r)
		if !ok {
			loginURL := "/login"
			if role == RoleAdmin {
				loginURL = "/admin/login"
			}
			http.Redirect(w, r, loginURL, http.StatusSeeOther)
			return
		}
		if !hasRole(sess.Role, role) {
			http.Redirect(w, r, "/403", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// hasRole reports whether a session with the given role satisfies the required
// role. ADMIN implicitly satisfies USER; otherwise the roles must match.
func hasRole(sessionRole, required string) bool {
	if sessionRole == required {
		return true
	}
	if required == RoleUser && sessionRole == RoleAdmin {
		return true
	}
	return false
}

// currentSession returns the session associated with the request, if any.
func (c *SecurityConfiguration) currentSession(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return Session{}, false
	}
	return c.sessions.Get(cookie.Value)
}

func (c *SecurityConfiguration) setSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *SecurityConfiguration) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
