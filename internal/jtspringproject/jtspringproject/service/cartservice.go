// Package service contains the Go migration of the JtSpringProject service
// layer (the com.jtspringproject.JtSpringProject.services package).
//
// This file migrates cartService.java, a thin service-layer wrapper around the
// cartDao that exposes CRUD operations for shopping Cart entities. In the Java
// world it was a Spring @Service that delegated every call directly to the DAO
// with no additional business logic:
//
//	@Service
//	public class cartService {
//	    private final cartDao cartDao;
//	    @Autowired public cartService(cartDao cartDao) { ... }
//	    public Cart addCart(Cart cart) { return cartDao.addCart(cart); }
//	    public List<Cart> getCarts() { return cartDao.getCarts(); }
//	    public void updateCart(Cart cart) { cartDao.updateCart(cart); }
//	    public void deleteCart(Cart cart) { cartDao.deleteCart(cart); }
//	}
//
// MIGRATION_NOTE: The corresponding cartDao.java has not yet been migrated to
// the repository package (there is no CartRepository listed among the already
// migrated symbols). To keep this service decoupled from any concrete DAO
// implementation and to follow idiomatic Go ("prefer interfaces over concrete
// types in function signatures"), this file defines a CartRepository interface
// describing the operations the service depends on. Once cartDao.java is
// migrated, its repository type should satisfy this interface. The Cart model
// (models/Cart.java) is likewise not yet migrated, so a placeholder Cart type
// is declared here and MUST be replaced with the real model type / import once
// available.
package service

import "context"

// Cart is a placeholder for the shopping cart entity.
//
// MIGRATION_NOTE: The Java Cart model (com.jtspringproject.JtSpringProject
// .models.Cart) has not been migrated yet. This minimal struct exists only so
// that this service compiles. Replace it with the real migrated model type
// (e.g. repository.Cart or model.Cart) once it is available, updating the
// CartRepository interface and CartService methods accordingly.
type Cart struct {
	ID int
}

// CartRepository describes the persistence operations the CartService depends
// on. It mirrors the Java cartDao contract.
//
// MIGRATION_NOTE: cartDao.java has not been migrated yet. When it is, the
// concrete repository type should implement this interface. The Java DAO
// methods (addCart, getCarts, updateCart, deleteCart) did not accept a context
// or return errors; idiomatic Go requires both for I/O operations, so the
// signatures here add context.Context and error returns. Adjust to match the
// migrated repository's actual signatures.
type CartRepository interface {
	// AddCart persists the given cart and returns the stored entity.
	AddCart(ctx context.Context, cart Cart) (Cart, error)
	// GetCarts returns all carts.
	GetCarts(ctx context.Context) ([]Cart, error)
	// UpdateCart updates the given cart.
	UpdateCart(ctx context.Context, cart Cart) error
	// DeleteCart removes the given cart.
	DeleteCart(ctx context.Context, cart Cart) error
}

// CartService provides CRUD operations for shopping carts. It is a thin
// pass-through over a CartRepository, mirroring the Java cartService.
type CartService struct {
	repo CartRepository
}

// NewCartService constructs a CartService backed by the given CartRepository.
// It replaces the Spring @Autowired constructor injection with explicit
// dependency injection.
func NewCartService(repo CartRepository) *CartService {
	return &CartService{repo: repo}
}

// AddCart persists the given cart and returns the stored entity.
func (s *CartService) AddCart(ctx context.Context, cart Cart) (Cart, error) {
	return s.repo.AddCart(ctx, cart)
}

// GetCarts returns all carts.
func (s *CartService) GetCarts(ctx context.Context) ([]Cart, error) {
	return s.repo.GetCarts(ctx)
}

// UpdateCart updates the given cart.
func (s *CartService) UpdateCart(ctx context.Context, cart Cart) error {
	return s.repo.UpdateCart(ctx, cart)
}

// DeleteCart removes the given cart.
func (s *CartService) DeleteCart(ctx context.Context, cart Cart) error {
	return s.repo.DeleteCart(ctx, cart)
}
