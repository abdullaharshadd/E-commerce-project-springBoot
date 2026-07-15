package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
	"golang.org/x/crypto/bcrypt"
)

// ErrDataIntegrity indicates that an operation could not be completed because
// it would violate a persistence-layer data-integrity constraint (for example
// a duplicate username).
//
// MIGRATION_NOTE: The original Spring code caught
// DataIntegrityViolationException and re-threw it as an IllegalStateException.
// In Go we surface this as a sentinel error that callers can match with
// errors.Is. The repository layer is expected to wrap its underlying
// constraint-violation errors with ErrDataIntegrity (or callers can detect the
// wrapped error). See notes for manual review.
var ErrDataIntegrity = errors.New("unable to create user due to data integrity constraints")

// UserRepository is the subset of repository operations required by the
// UserService. Depending on an interface (rather than the concrete
// *repository.UserDAO) keeps the service testable and decoupled from the
// persistence implementation.
//
// MIGRATION_NOTE: The original Spring @Service injected the concrete userDao
// via constructor-based @Autowired. In idiomatic Go we accept an interface in
// the constructor, which both documents the required behaviour and allows tests
// to substitute a fake without a DI container.
type UserRepository interface {
	GetAllUser(ctx context.Context) ([]model.User, error)
	SaveUser(ctx context.Context, user model.User) (model.User, error)
	UserExists(ctx context.Context, username string) (bool, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, bool, error)
	GetUserByID(ctx context.Context, id int) (model.User, bool, error)
}

// PasswordEncoder abstracts password hashing and verification so the encoding
// strategy can be swapped out in tests.
//
// MIGRATION_NOTE: This replaces Spring Security's PasswordEncoder abstraction.
// The default implementation (BCryptPasswordEncoder) uses golang.org/x/crypto/bcrypt
// to match the original BCrypt behaviour.
type PasswordEncoder interface {
	// Encode hashes the supplied raw password.
	Encode(raw string) (string, error)
}

// BCryptPasswordEncoder is a PasswordEncoder backed by bcrypt.
type BCryptPasswordEncoder struct {
	cost int
}

// NewBCryptPasswordEncoder returns a BCryptPasswordEncoder. A cost of zero
// falls back to bcrypt.DefaultCost.
func NewBCryptPasswordEncoder(cost int) *BCryptPasswordEncoder {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BCryptPasswordEncoder{cost: cost}
}

// Encode hashes the raw password using bcrypt and returns the resulting hash.
func (e *BCryptPasswordEncoder) Encode(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), e.cost)
	if err != nil {
		return "", fmt.Errorf("encode password: %w", err)
	}
	return string(hash), nil
}

// UserService provides user management operations, layering BCrypt password
// encoding and lazy legacy-password migration on top of a UserRepository.
type UserService struct {
	repo    UserRepository
	encoder PasswordEncoder
}

// NewUserService constructs a UserService with its repository and password
// encoder dependencies.
func NewUserService(repo UserRepository, encoder PasswordEncoder) *UserService {
	return &UserService{repo: repo, encoder: encoder}
}

// GetUsers returns all users.
func (s *UserService) GetUsers(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.GetAllUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
}

// AddUser persists a new user, encoding its password with BCrypt when it is not
// already encoded. Data-integrity violations from the repository are wrapped in
// ErrDataIntegrity.
func (s *UserService) AddUser(ctx context.Context, user model.User) (model.User, error) {
	if user.Password != "" && !isPasswordEncoded(user.Password) {
		hashed, err := s.encoder.Encode(user.Password)
		if err != nil {
			return model.User{}, fmt.Errorf("add user: %w", err)
		}
		user.Password = hashed
	}

	saved, err := s.repo.SaveUser(ctx, user)
	if err != nil {
		// MIGRATION_NOTE: The original translated DataIntegrityViolationException
		// into IllegalStateException. Here we surface ErrDataIntegrity when the
		// repository reports such a constraint violation. The repository is
		// expected to wrap constraint errors so errors.Is(err, ErrDataIntegrity)
		// holds; otherwise the raw error is returned for the caller to inspect.
		if errors.Is(err, ErrDataIntegrity) {
			return model.User{}, ErrDataIntegrity
		}
		return model.User{}, fmt.Errorf("add user: %w", err)
	}
	return saved, nil
}

// CheckUserExists reports whether a user with the given username exists.
func (s *UserService) CheckUserExists(ctx context.Context, username string) (bool, error) {
	exists, err := s.repo.UserExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists, nil
}

// GetUserByUsername looks up a user by username. If the stored password is a
// legacy plain-text value it is lazily migrated to BCrypt and persisted before
// being returned. The boolean result reports whether the user was found.
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (model.User, bool, error) {
	user, found, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return model.User{}, false, fmt.Errorf("get user by username: %w", err)
	}
	if !found {
		return model.User{}, false, nil
	}

	// Migrate legacy plain-text passwords to BCrypt when the user is loaded.
	if user.Password != "" && !isPasswordEncoded(user.Password) {
		hashed, err := s.encoder.Encode(user.Password)
		if err != nil {
			return model.User{}, false, fmt.Errorf("get user by username: migrate password: %w", err)
		}
		user.Password = hashed
		if user, err = s.repo.SaveUser(ctx, user); err != nil {
			return model.User{}, false, fmt.Errorf("get user by username: persist migrated password: %w", err)
		}
	}
	return user, true, nil
}

// GetUserByID looks up a user by ID. The boolean result reports whether the
// user was found.
func (s *UserService) GetUserByID(ctx context.Context, id int) (model.User, bool, error) {
	user, found, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return model.User{}, false, fmt.Errorf("get user by id: %w", err)
	}
	return user, found, nil
}

// UpdateUserProfile updates the username, email, address and (optionally)
// password of an existing user. If the user does not exist it returns
// found=false. A blank or whitespace-only password leaves the existing password
// unchanged; a non-blank password is encoded unless it is already a BCrypt hash.
func (s *UserService) UpdateUserProfile(ctx context.Context, userID int, username, email, password, address string) (model.User, bool, error) {
	existing, found, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return model.User{}, false, fmt.Errorf("update user profile: %w", err)
	}
	if !found {
		return model.User{}, false, nil
	}

	existing.Username = username
	existing.Email = email
	existing.Address = address

	if strings.TrimSpace(password) != "" {
		if isPasswordEncoded(password) {
			existing.Password = password
		} else {
			hashed, err := s.encoder.Encode(password)
			if err != nil {
				return model.User{}, false, fmt.Errorf("update user profile: %w", err)
			}
			existing.Password = hashed
		}
	}

	saved, err := s.repo.SaveUser(ctx, existing)
	if err != nil {
		return model.User{}, false, fmt.Errorf("update user profile: %w", err)
	}
	return saved, true, nil
}

// isPasswordEncoded reports whether the supplied password already appears to be
// a BCrypt hash based on its version prefix.
func isPasswordEncoded(password string) bool {
	return strings.HasPrefix(password, "$2a$") ||
		strings.HasPrefix(password, "$2b$") ||
		strings.HasPrefix(password, "$2y$")
}
