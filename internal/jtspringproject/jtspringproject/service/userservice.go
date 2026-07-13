// Package service contains the Go migration of the JtSpringProject service
// layer (the com.jtspringproject.JtSpringProject.services package).
//
// This file migrates userService.java, the service-layer wrapper around the
// userDao that exposes CRUD operations for User entities. Unlike the other
// service classes it carries a small amount of genuine business logic: it
// BCrypt-encodes passwords before persisting them and lazily migrates legacy
// plain-text passwords to BCrypt when a user is loaded by username.
//
// In the Java world it was a Spring @Service with constructor injection:
//
//	@Service
//	public class userService {
//	    private final userDao userDao;
//	    private final PasswordEncoder passwordEncoder;
//	    @Autowired public userService(userDao, PasswordEncoder) { ... }
//	    ...
//	}
//
// Migration decisions:
//   - The DAO dependency is expressed as a UserRepository interface so the
//     service depends on an abstraction (idiomatic Go, mirrors the @Autowired
//     DAO). A concrete implementation lives in the repository package.
//   - Spring's PasswordEncoder strategy is replaced by a small PasswordEncoder
//     interface with a bcrypt-backed default implementation
//     (BCryptPasswordEncoder), preserving the Strategy pattern.
//   - context.Context is threaded through every method since these operations
//     ultimately perform I/O against the database.
//   - Spring's DataIntegrityViolationException -> IllegalStateException
//     translation is preserved by wrapping the repository error in
//     ErrDataIntegrity when the repository reports a data-integrity failure.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/repository"
)

// ErrDataIntegrity is returned by AddUser when the underlying repository
// rejects the user because of a data-integrity constraint (for example a
// duplicate username). It is the Go equivalent of translating Spring's
// DataIntegrityViolationException into an IllegalStateException.
var ErrDataIntegrity = errors.New("unable to create user due to data integrity constraints")

// User models the fields of the domain User the service operates on. It mirrors
// the getters/setters used by userService.java. The concrete repository entity
// type is expected to satisfy this contract through the UserRepository
// interface below.
//
// MIGRATION_NOTE: The Java code manipulated a mutable com.jtspringproject...
// models.User via getters/setters. The Go User type and the UserRepository
// interface below assume a repository-owned User entity. Adjust the field/method
// names to match the actual entity type produced by the repository package
// during integration.
type User interface {
	// GetPassword returns the user's (possibly encoded) password.
	GetPassword() string
	// SetPassword replaces the user's password.
	SetPassword(password string)
	// SetUsername replaces the user's username.
	SetUsername(username string)
	// SetEmail replaces the user's email address.
	SetEmail(email string)
	// SetAddress replaces the user's address.
	SetAddress(address string)
}

// UserRepository abstracts the persistence operations the user service relies
// on. It corresponds to the injected userDao in the Java service. The concrete
// implementation lives in
// internal/jtspringproject/jtspringproject/repository/userdao.go.
type UserRepository interface {
	// GetAllUser returns every persisted user.
	GetAllUser(ctx context.Context) ([]User, error)
	// SaveUser persists (inserts or updates) the given user and returns the
	// stored entity.
	SaveUser(ctx context.Context, user User) (User, error)
	// UserExists reports whether a user with the given username exists.
	UserExists(ctx context.Context, username string) (bool, error)
	// GetUserByUsername looks up a user by username. It returns
	// repository.ErrUserNotFound when no such user exists.
	GetUserByUsername(ctx context.Context, username string) (User, error)
	// GetUserByID looks up a user by id. It returns
	// repository.ErrUserNotFound when no such user exists.
	GetUserByID(ctx context.Context, id int) (User, error)
}

// PasswordEncoder abstracts password hashing so the service is decoupled from a
// concrete algorithm. It is the Go equivalent of Spring Security's
// PasswordEncoder strategy interface.
type PasswordEncoder interface {
	// Encode hashes the given raw password.
	Encode(raw string) (string, error)
}

// BCryptPasswordEncoder is the default PasswordEncoder implementation backed by
// golang.org/x/crypto/bcrypt. It mirrors Spring's BCryptPasswordEncoder.
type BCryptPasswordEncoder struct {
	cost int
}

// NewBCryptPasswordEncoder constructs a BCryptPasswordEncoder. A cost of zero or
// below falls back to bcrypt.DefaultCost.
func NewBCryptPasswordEncoder(cost int) *BCryptPasswordEncoder {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BCryptPasswordEncoder{cost: cost}
}

// Encode hashes raw using bcrypt.
func (e *BCryptPasswordEncoder) Encode(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), e.cost)
	if err != nil {
		return "", fmt.Errorf("encode password: %w", err)
	}
	return string(hash), nil
}

// UserService provides user-management business logic on top of a
// UserRepository, handling BCrypt password encoding and lazy migration of
// legacy plain-text passwords. It is the Go migration of the Spring
// userService @Service bean.
type UserService struct {
	repo    UserRepository
	encoder PasswordEncoder
}

// NewUserService constructs a UserService with its required dependencies,
// replacing Spring constructor injection (@Autowired). Both dependencies must
// be non-nil.
func NewUserService(repo UserRepository, encoder PasswordEncoder) *UserService {
	return &UserService{repo: repo, encoder: encoder}
}

// GetUsers returns all users. It corresponds to userService.getUsers().
func (s *UserService) GetUsers(ctx context.Context) ([]User, error) {
	users, err := s.repo.GetAllUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
}

// AddUser encodes the user's password (unless it is already BCrypt-encoded) and
// persists the user. If the repository reports a data-integrity violation the
// error is wrapped with ErrDataIntegrity, mirroring the Java translation of
// DataIntegrityViolationException into IllegalStateException.
func (s *UserService) AddUser(ctx context.Context, user User) (User, error) {
	if user == nil {
		return nil, errors.New("add user: user must not be nil")
	}

	if pwd := user.GetPassword(); pwd != "" && !isPasswordEncoded(pwd) {
		encoded, err := s.encoder.Encode(pwd)
		if err != nil {
			return nil, fmt.Errorf("add user: %w", err)
		}
		user.SetPassword(encoded)
	}

	saved, err := s.repo.SaveUser(ctx, user)
	if err != nil {
		// MIGRATION_NOTE: In Java only DataIntegrityViolationException was
		// translated. Here we detect that condition via the repository's
		// sentinel error. If the repository package exposes a dedicated
		// data-integrity error, match it with errors.Is instead of the
		// generic wrapping below.
		if isDataIntegrityViolation(err) {
			return nil, fmt.Errorf("%w: %v", ErrDataIntegrity, err)
		}
		return nil, fmt.Errorf("add user: %w", err)
	}
	return saved, nil
}

// CheckUserExists reports whether a user with the given username exists. It
// corresponds to userService.checkUserExists(String).
func (s *UserService) CheckUserExists(ctx context.Context, username string) (bool, error) {
	exists, err := s.repo.UserExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists, nil
}

// GetUserByUsername loads a user by username. If the stored password is a legacy
// plain-text value it is BCrypt-encoded and persisted before the user is
// returned, mirroring the lazy migration behaviour of the Java service.
//
// It returns repository.ErrUserNotFound (wrapped) when no such user exists;
// callers can detect this with errors.Is(err, repository.ErrUserNotFound).
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	if pwd := user.GetPassword(); pwd != "" && !isPasswordEncoded(pwd) {
		encoded, err := s.encoder.Encode(pwd)
		if err != nil {
			return nil, fmt.Errorf("get user by username: %w", err)
		}
		user.SetPassword(encoded)
		if _, err := s.repo.SaveUser(ctx, user); err != nil {
			return nil, fmt.Errorf("get user by username: migrate password: %w", err)
		}
	}
	return user, nil
}

// GetUserByID loads a user by id. It corresponds to userService.getUserById(int).
// It returns repository.ErrUserNotFound (wrapped) when no such user exists.
func (s *UserService) GetUserByID(ctx context.Context, id int) (User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// UpdateUserProfile updates the username, email, address and (optionally)
// password of the user identified by userID, then persists the changes.
//
// If no user exists with the given id it returns (nil, nil), mirroring the Java
// method which returned null in that case. A blank password leaves the existing
// password untouched; a non-blank password is stored as-is when already encoded
// and BCrypt-encoded otherwise.
func (s *UserService) UpdateUserProfile(ctx context.Context, userID int, username, email, password, address string) (User, error) {
	existing, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		// Preserve the Java "return null" semantics for a missing user.
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("update user profile: %w", err)
	}
	if existing == nil {
		return nil, nil
	}

	existing.SetUsername(username)
	existing.SetEmail(email)
	existing.SetAddress(address)

	if strings.TrimSpace(password) != "" {
		if isPasswordEncoded(password) {
			existing.SetPassword(password)
		} else {
			encoded, err := s.encoder.Encode(password)
			if err != nil {
				return nil, fmt.Errorf("update user profile: %w", err)
			}
			existing.SetPassword(encoded)
		}
	}

	saved, err := s.repo.SaveUser(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}
	return saved, nil
}

// isPasswordEncoded reports whether password looks like a BCrypt hash. It
// mirrors the private helper of the same name in the Java service.
func isPasswordEncoded(password string) bool {
	return strings.HasPrefix(password, "$2a$") ||
		strings.HasPrefix(password, "$2b$") ||
		strings.HasPrefix(password, "$2y$")
}

// isDataIntegrityViolation reports whether err represents a data-integrity
// constraint violation coming from the repository layer.
//
// MIGRATION_NOTE: The repository package does not (yet) expose a dedicated
// data-integrity sentinel error, so this heuristic inspects the error message.
// Replace this with errors.Is against a proper sentinel (e.g.
// repository.ErrDataIntegrity) once one is available.
func isDataIntegrityViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "data integrity") ||
		strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "constraint")
}
