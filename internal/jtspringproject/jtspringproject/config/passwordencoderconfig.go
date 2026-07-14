// Package config contains the migrated security/configuration helpers for the
// original Spring application.
//
// This file migrates PasswordEncoderConfig.java. The original class was a Spring
// @Configuration that exposed a single @Bean of type PasswordEncoder, backed by
// a BCryptPasswordEncoder. The bean was injected wherever password hashing or
// verification was required (typically the authentication / user-registration
// layer).
//
// MIGRATION_NOTE: Go has no IoC container, so there is no "bean" to register.
// The idiomatic replacement is a small, dependency-free type that wraps the
// well-maintained golang.org/x/crypto/bcrypt package. Callers construct it
// explicitly (via NewBCryptPasswordEncoder) and pass it down the call stack,
// exactly as they would have received the injected Spring bean.
//
// MIGRATION_NOTE: The Spring PasswordEncoder interface has two methods:
//
//   String encode(CharSequence rawPassword)
//   boolean matches(CharSequence rawPassword, String encodedPassword)
//
// In idiomatic Go, encode can fail (e.g. password exceeds bcrypt's 72-byte
// limit), so Encode returns (string, error) instead of panicking. matches maps
// cleanly to Matches, which returns a plain bool because a hash mismatch is not
// an error condition — it is an expected, valid outcome.
//
// This file registers NO HTTP routes: the source configuration class defined
// none. Route registration belongs in the migrated controller/handler files.
package config

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordEncoder abstracts password hashing and verification. It is the Go
// analogue of Spring Security's org.springframework.security.crypto.password.PasswordEncoder
// interface and lets callers depend on the behaviour rather than a concrete
// implementation (e.g. for testing with a stub encoder).
type PasswordEncoder interface {
	// Encode hashes the given raw password and returns the encoded form.
	// It returns a non-nil error if the password cannot be hashed (for
	// example, if it exceeds bcrypt's 72-byte maximum length).
	Encode(rawPassword string) (string, error)

	// Matches reports whether the raw password, once hashed, equals the
	// previously encoded password. A mismatch returns false with no error;
	// this is an expected outcome, not a failure.
	Matches(rawPassword, encodedPassword string) bool
}

// BCryptPasswordEncoder is a PasswordEncoder backed by the bcrypt algorithm.
// It is the Go equivalent of Spring Security's BCryptPasswordEncoder.
type BCryptPasswordEncoder struct {
	// cost is the bcrypt work factor. A higher cost is more secure but
	// slower. bcrypt.DefaultCost (10) matches the default used by Spring's
	// BCryptPasswordEncoder, preserving the original behaviour.
	cost int
}

// BCryptOption customises a BCryptPasswordEncoder via the functional-options
// pattern.
type BCryptOption func(*BCryptPasswordEncoder)

// WithCost sets the bcrypt work factor. Values outside the range accepted by
// bcrypt (currently 4-31) are clamped by the underlying library at hashing
// time; callers should pass a sensible value such as bcrypt.DefaultCost.
func WithCost(cost int) BCryptOption {
	return func(e *BCryptPasswordEncoder) {
		e.cost = cost
	}
}

// NewBCryptPasswordEncoder constructs a BCryptPasswordEncoder. With no options
// it uses bcrypt.DefaultCost, matching the default configuration of the
// original Spring @Bean.
//
// This is the idiomatic replacement for the Spring @Bean factory method:
// instead of the container caching a singleton, callers create one instance
// (typically once, in main) and inject it explicitly.
func NewBCryptPasswordEncoder(opts ...BCryptOption) *BCryptPasswordEncoder {
	e := &BCryptPasswordEncoder{cost: bcrypt.DefaultCost}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Encode hashes rawPassword using bcrypt and returns the encoded string.
func (e *BCryptPasswordEncoder) Encode(rawPassword string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), e.cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Matches reports whether rawPassword corresponds to encodedPassword. Any error
// from bcrypt (including a genuine mismatch) results in false, mirroring the
// boolean contract of Spring's PasswordEncoder.matches.
func (e *BCryptPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(rawPassword)) == nil
}

// Compile-time assertion that *BCryptPasswordEncoder satisfies PasswordEncoder.
var _ PasswordEncoder = (*BCryptPasswordEncoder)(nil)
