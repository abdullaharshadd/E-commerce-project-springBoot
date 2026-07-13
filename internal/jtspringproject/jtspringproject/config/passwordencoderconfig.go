// Package config contains the Go migration of the JtSpringProject Spring
// configuration classes.
//
// This file migrates PasswordEncoderConfig.java, a Spring @Configuration class
// whose sole responsibility was to expose a BCryptPasswordEncoder as a
// container-managed PasswordEncoder bean. Spring Security consumed this bean
// for hashing new passwords and verifying login credentials.
//
// MIGRATION_NOTE: Go has no IoC container, so the Spring @Configuration/@Bean
// factory pattern does not translate directly. Instead of registering a
// singleton in a container, we expose:
//
//   - a PasswordEncoder interface (the abstraction Spring Security relied on),
//   - a bcrypt-backed implementation, and
//   - a NewBCryptPasswordEncoder constructor.
//
// Callers perform explicit dependency injection by constructing the encoder
// once (e.g. in main / application wiring) and passing it down the call stack,
// which is the idiomatic Go replacement for a singleton bean.
//
// MIGRATION_NOTE: Spring's BCryptPasswordEncoder uses a default strength
// (log rounds) of 10. golang.org/x/crypto/bcrypt exposes bcrypt.DefaultCost,
// which is also 10, so the default behaviour is preserved exactly.
package config

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordEncoder abstracts password hashing and verification. It is the Go
// equivalent of Spring Security's org.springframework.security.crypto.password
// .PasswordEncoder interface, exposing only the two operations the application
// actually used.
type PasswordEncoder interface {
	// Encode hashes the given raw password and returns the encoded form.
	Encode(rawPassword string) (string, error)

	// Matches reports whether rawPassword, once encoded, matches the
	// previously encoded password. It never returns an error for a simple
	// mismatch; a non-nil error indicates the encodedPassword was malformed
	// or otherwise unusable.
	Matches(rawPassword, encodedPassword string) (bool, error)
}

// bcryptPasswordEncoder is the bcrypt-backed implementation of PasswordEncoder.
// It mirrors Spring Security's BCryptPasswordEncoder.
type bcryptPasswordEncoder struct {
	// cost is the bcrypt cost factor (log rounds). Spring's default strength
	// is 10, matching bcrypt.DefaultCost.
	cost int
}

// NewBCryptPasswordEncoder returns a PasswordEncoder backed by bcrypt using the
// default cost (bcrypt.DefaultCost == 10), matching Spring Security's
// BCryptPasswordEncoder default strength.
//
// This constructor replaces the Spring @Bean factory method
// PasswordEncoderConfig#passwordEncoder(). Wire the returned value into the
// components that need it instead of relying on container injection.
func NewBCryptPasswordEncoder() PasswordEncoder {
	return &bcryptPasswordEncoder{cost: bcrypt.DefaultCost}
}

// NewBCryptPasswordEncoderWithCost returns a PasswordEncoder backed by bcrypt
// using the supplied cost factor. It returns an error if cost is outside the
// range supported by bcrypt (bcrypt.MinCost..bcrypt.MaxCost).
//
// This has no direct Spring equivalent in the source file but is provided for
// callers that need to tune the work factor explicitly.
func NewBCryptPasswordEncoderWithCost(cost int) (PasswordEncoder, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("config: bcrypt cost %d out of range [%d, %d]", cost, bcrypt.MinCost, bcrypt.MaxCost)
	}
	return &bcryptPasswordEncoder{cost: cost}, nil
}

// Encode hashes rawPassword with bcrypt and returns the encoded hash string.
func (e *bcryptPasswordEncoder) Encode(rawPassword string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), e.cost)
	if err != nil {
		return "", fmt.Errorf("config: encoding password: %w", err)
	}
	return string(hashed), nil
}

// Matches reports whether rawPassword matches encodedPassword.
//
// A false result with a nil error means the passwords simply do not match. A
// non-nil error indicates encodedPassword was not a valid bcrypt hash (or some
// other unexpected failure), which callers should treat as a verification
// failure and log for investigation.
func (e *bcryptPasswordEncoder) Matches(rawPassword, encodedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(rawPassword))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, fmt.Errorf("config: verifying password: %w", err)
}
