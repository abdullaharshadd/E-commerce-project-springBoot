// Package config contains the Go migration of the JtSpringProject
// configuration layer (the com.jtspringproject.JtSpringProject.configuration
// package).
//
// This file migrates SecurityConfiguration.java, which in the Java world used
// Spring Security to declare two ordered SecurityFilterChain beans and a
// UserDetailsService bean:
//
//   - An admin chain (@Order(1)) scoped to /admin/** with form login at
//     /admin/login, GET logout at /admin/logout, and ROLE_ADMIN required for
//     every /admin/** route except /admin/login.
//   - A user chain (@Order(2)) covering everything else, permitting /login,
//     /register and /newuserregister, and requiring ROLE_USER for /**, with
//     form login at /login and GET logout at /logout.
//   - A UserDetailsService that loads a user by username from the userService,
//     maps the stored role ("ROLE_ADMIN" => ADMIN, otherwise USER) and returns
//     Spring's UserDetails.
//
// Go has no Spring Security. There is no framework-provided filter-chain
// builder, no automatic form-login processing, no session cookie management
// and no @Order bean precedence. Rather than inventing an equivalent DSL, this
// migration keeps the *policy* (which route requires which role, which routes
// are public, the login/logout/access-denied URLs) as plain data and exposes
// an authentication helper that mirrors the Java UserDetailsService lambda.
// Wiring these into an actual net/http middleware chain is the router layer's
// responsibility and must be done during manual review.
//
// MIGRATION_NOTE: The heavy lifting of Spring Security (session management,
// CSRF, JSESSIONID cookies, form-login POST handling, redirect-on-success/
// failure) is NOT reproduced here. See the SecurityPolicy fields and the
// AuthenticateByUsername method for the pieces that carry real business logic;
// everything framework-specific requires a manual middleware implementation.
package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/service"
)

// Role names used by the authorization policy. These correspond to the
// hasRole("ADMIN") / hasRole("USER") checks in the Java filter chains.
const (
	// RoleAdmin is required for the /admin/** routes.
	RoleAdmin = "ADMIN"
	// RoleUser is required for the general /** routes.
	RoleUser = "USER"
)

// storedRoleAdmin is the role string as persisted on the User model. The Java
// code compared getRole() against the literal "ROLE_ADMIN".
const storedRoleAdmin = "ROLE_ADMIN"

// ErrUserNotFound mirrors Spring's UsernameNotFoundException: it is returned
// when no user exists for the supplied username.
var ErrUserNotFound = errors.New("user not found")

// UserProvider is the minimal slice of the user service that the security
// configuration depends on. Defining it as an interface here (rather than
// depending on the concrete *service.UserService) keeps the config decoupled
// and testable, matching Go's "accept interfaces" convention.
//
// MIGRATION_NOTE: The exact return type of GetUserByUsername comes from the
// migrated userService. Adjust the signature to match service.UserService if
// it differs (e.g. if it returns (*model.User, error)).
type UserProvider interface {
	// GetUserByUsername loads a user by username. Implementations should return
	// a nil user (or an error) when no such user exists.
	GetUserByUsername(ctx context.Context, username string) (*service.User, error)
}

// UserDetails is the Go analogue of Spring Security's UserDetails value object
// returned by the UserDetailsService. It carries only what the Java lambda
// populated: username, password hash and a single mapped role.
type UserDetails struct {
	// Username is the authenticated principal's username.
	Username string
	// Password is the (BCrypt) password hash loaded from the store.
	Password string
	// Role is the mapped authority: RoleAdmin or RoleUser.
	Role string
}

// RouteRule describes a single authorization rule: a path pattern (Ant-style,
// as in the Java config) and the role required to access it. A PublicAccess
// rule requires no role.
type RouteRule struct {
	// Pattern is the Ant-style path pattern, e.g. "/admin/**" or "/login".
	Pattern string
	// RequiredRole is the role needed to access the pattern. Empty means the
	// route is public (permitAll in Spring terms).
	RequiredRole string
}

// PublicAccess is a convenience for a RouteRule that permits everyone.
func PublicAccess(pattern string) RouteRule {
	return RouteRule{Pattern: pattern}
}

// FilterChainPolicy captures the intent of one Spring SecurityFilterChain
// bean as plain configuration data: the routes it protects, its form-login
// endpoints, its logout behaviour and its access-denied page.
//
// Rules are evaluated in order, matching Spring's authorizeHttpRequests
// semantics where the first matching rule wins.
type FilterChainPolicy struct {
	// Name identifies the chain ("admin" or "user") for logging/debugging.
	Name string
	// BasePathMatcher, when non-empty, scopes the chain to a path prefix. It
	// mirrors Spring's http.antMatcher("/admin/**"). Empty means the chain
	// applies to all remaining requests.
	BasePathMatcher string
	// Rules are the ordered authorization rules for this chain.
	Rules []RouteRule
	// LoginPage is the URL of the login form (Spring loginPage).
	LoginPage string
	// LoginProcessingURL is the URL that processes login POSTs (Spring
	// loginProcessingUrl).
	LoginProcessingURL string
	// LoginSuccessRedirect is where to send the user after a successful login.
	LoginSuccessRedirect string
	// LoginFailureRedirect is where to send the user after a failed login.
	LoginFailureRedirect string
	// LogoutURL is the (GET) logout URL, kept as GET to remain compatible with
	// existing anchor links, exactly as noted in the Java source.
	LogoutURL string
	// LogoutMethod is the HTTP method accepted for logout ("GET" here).
	LogoutMethod string
	// LogoutSuccessURL is where to redirect after logout.
	LogoutSuccessURL string
	// DeleteCookies lists cookies to clear on logout (JSESSIONID in Java).
	DeleteCookies []string
	// AccessDeniedPage is the page shown on a 403 (Spring accessDeniedPage).
	AccessDeniedPage string
}

// SecurityConfiguration is the Go replacement for the Spring @Configuration
// class. It holds the dependency (the user provider) and produces the
// declarative filter-chain policies plus the authentication helper.
type SecurityConfiguration struct {
	users UserProvider
}

// NewSecurityConfiguration constructs a SecurityConfiguration. It replaces the
// constructor-based dependency injection performed by Spring's @Autowired.
func NewSecurityConfiguration(users UserProvider) (*SecurityConfiguration, error) {
	if users == nil {
		return nil, errors.New("config: users provider must not be nil")
	}
	return &SecurityConfiguration{users: users}, nil
}

// AdminFilterChain returns the policy for the admin routes. It corresponds to
// the @Order(1) AdminConfigurationAdapter filter chain: /admin/login is public
// while every other /admin/** route requires ROLE_ADMIN.
func (c *SecurityConfiguration) AdminFilterChain() FilterChainPolicy {
	return FilterChainPolicy{
		Name:            "admin",
		BasePathMatcher: "/admin/**",
		Rules: []RouteRule{
			PublicAccess("/admin/login"),
			{Pattern: "/admin/**", RequiredRole: RoleAdmin},
		},
		LoginPage:            "/admin/login",
		LoginProcessingURL:   "/admin/loginvalidate",
		LoginSuccessRedirect: "/admin/",
		LoginFailureRedirect: "/admin/login?error=true",
		LogoutURL:            "/admin/logout",
		LogoutMethod:         "GET",
		LogoutSuccessURL:     "/admin/login",
		DeleteCookies:        []string{"JSESSIONID"},
		AccessDeniedPage:     "/403",
	}
}

// UserFilterChain returns the policy for the general user routes. It
// corresponds to the @Order(2) UserConfigurationAdapter filter chain: /login,
// /register and /newuserregister are public while everything else requires
// ROLE_USER.
func (c *SecurityConfiguration) UserFilterChain() FilterChainPolicy {
	return FilterChainPolicy{
		Name: "user",
		Rules: []RouteRule{
			PublicAccess("/login"),
			PublicAccess("/register"),
			PublicAccess("/newuserregister"),
			{Pattern: "/**", RequiredRole: RoleUser},
		},
		LoginPage:            "/login",
		LoginProcessingURL:   "/userloginvalidate",
		LoginSuccessRedirect: "/",
		LoginFailureRedirect: "/login?error=true",
		LogoutURL:            "/logout",
		LogoutMethod:         "GET",
		LogoutSuccessURL:     "/login",
		DeleteCookies:        []string{"JSESSIONID"},
		AccessDeniedPage:     "/403",
	}
}

// FilterChains returns the policies in precedence order, mirroring Spring's
// @Order(1) (admin) before @Order(2) (user).
func (c *SecurityConfiguration) FilterChains() []FilterChainPolicy {
	return []FilterChainPolicy{
		c.AdminFilterChain(),
		c.UserFilterChain(),
	}
}

// AuthenticateByUsername is the Go equivalent of the Spring
// UserDetailsService bean. It loads a user by username, maps the stored role
// to the authorization role and returns the resulting UserDetails.
//
// It returns ErrUserNotFound (wrapped) when no user exists, replacing Spring's
// UsernameNotFoundException.
func (c *SecurityConfiguration) AuthenticateByUsername(ctx context.Context, username string) (UserDetails, error) {
	user, err := c.users.GetUserByUsername(ctx, username)
	if err != nil {
		return UserDetails{}, fmt.Errorf("config: loading user %q: %w", username, err)
	}
	if user == nil {
		return UserDetails{}, fmt.Errorf("config: user with username %q: %w", username, ErrUserNotFound)
	}

	return UserDetails{
		Username: username,
		Password: user.Password,
		Role:     mapRole(user.Role),
	}, nil
}

// mapRole reproduces the Java expression:
//
//	"ROLE_ADMIN".equals(user.getRole()) ? "ADMIN" : "USER"
func mapRole(storedRole string) string {
	if storedRole == storedRoleAdmin {
		return RoleAdmin
	}
	return RoleUser
}