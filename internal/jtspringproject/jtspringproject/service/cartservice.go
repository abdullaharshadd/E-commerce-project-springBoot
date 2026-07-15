package service

import (
	"context"
	"fmt"

	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/model"
	"github.com/jtspringproject/JtSpringProject/internal/jtspringproject/jtspringproject/repository"
)

// CartRepository is the subset of repository operations required by the
// CartService. Depending on an interface (rather than the concrete
// *repository.CartDAO) keeps the service testable and decoupled from the
// persistence implementation.
//
// MIGRATION_NOTE: The original Spring @Service injected the concrete cartDao
// via @Autowired. In idiomatic Go we accept an interface in the constructor,
// which both documents the required behaviour and allows tests to substitute a
// fake without a DI container.
type CartRepository interface {
	AddCart(ctx context.Context, cart *model.Cart) (*model.Cart, error)
	GetCarts(ctx context.Context) ([]*model.Cart, error)
	UpdateCart(ctx context.Context, cart *model.Cart) error
	DeleteCart(ctx context.Context, cart *model.Cart) error
}

// Ensure the concrete DAO satisfies the interface at compile time.
var _ CartRepository = (*repository.CartDAO)(nil)

// CartService is the service-layer facade for cart operations. It delegates
// persistence to a CartRepository.
//
// MIGRATION_NOTE: The original class contained no business logic beyond
// delegating to the DAO. The Go equivalent is intentionally thin as well; it
// exists so that controllers depend on a service abstraction rather than the
// repository directly, mirroring the original layering.
type CartService struct {
	cartRepo CartRepository
}

// NewCartService constructs a CartService backed by the given CartRepository.
// It returns an error if the dependency is nil to fail fast during wiring.
func NewCartService(cartRepo CartRepository) (*CartService, error) {
	if cartRepo == nil {
		return nil, fmt.Errorf("service: cart repository must not be nil")
	}
	return &CartService{cartRepo: cartRepo}, nil
}

// AddCart persists a new cart and returns the stored entity.
func (s *CartService) AddCart(ctx context.Context, cart *model.Cart) (*model.Cart, error) {
	saved, err := s.cartRepo.AddCart(ctx, cart)
	if err != nil {
		return nil, fmt.Errorf("service: add cart: %w", err)
	}
	return saved, nil
}

// GetCarts returns all carts.
func (s *CartService) GetCarts(ctx context.Context) ([]*model.Cart, error) {
	carts, err := s.cartRepo.GetCarts(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: get carts: %w", err)
	}
	return carts, nil
}

// UpdateCart persists changes to an existing cart.
func (s *CartService) UpdateCart(ctx context.Context, cart *model.Cart) error {
	if err := s.cartRepo.UpdateCart(ctx, cart); err != nil {
		return fmt.Errorf("service: update cart: %w", err)
	}
	return nil
}

// DeleteCart removes an existing cart.
func (s *CartService) DeleteCart(ctx context.Context, cart *model.Cart) error {
	if err := s.cartRepo.DeleteCart(ctx, cart); err != nil {
		return fmt.Errorf("service: delete cart: %w", err)
	}
	return nil
}
