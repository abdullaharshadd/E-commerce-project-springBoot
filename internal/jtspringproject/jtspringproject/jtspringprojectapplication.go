package jtspringproject

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const defaultServerAddr = ":8080"
const defaultShutdownTimeout = 15 * time.Second

type Application struct {
	server *http.Server
}

func NewApplication(ctx context.Context, addr string) (*Application, error) {
	if ctx == nil {
		return nil, errors.New("jtspringproject: nil context passed to NewApplication")
	}
	if addr == "" {
		addr = defaultServerAddr
	}

	handler, err := buildHandler(ctx)
	if err != nil {
		return nil, fmt.Errorf("jtspringproject: build handler: %w", err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &Application{server: srv}, nil
}

func stub(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "route: %s\n", name)
	}
}

func buildHandler(_ context.Context) (http.Handler, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/", stub("IndexPage"))
	r.Get("/register", stub("RegisterUser"))
	r.Post("/newuserregister", stub("RegisterNewUser"))
	r.Get("/login", stub("UserLogin"))
	r.Get("/buy", stub("Buy"))
	r.Get("/user/products", stub("GetUserProducts"))
	r.Get("/profileDisplay", stub("ProfileDisplay"))
	r.Post("/updateuser", stub("UpdateUserProfile"))

	r.Route("/admin", func(r chi.Router) {
		r.Get("/", stub("AdminHome"))
		r.Get("/index", stub("AdminIndex"))
		r.Get("/login", stub("AdminLogin"))
		r.Get("/Dashboard", stub("AdminHome"))
		r.Get("/categories", stub("GetCategories"))
		r.Post("/categories", stub("AddCategory"))
		r.Post("/categories/delete", stub("DeleteCategory"))
		r.Post("/categories/update", stub("UpdateCategory"))
		r.Get("/products", stub("GetAdminProducts"))
		r.Get("/products/add", stub("AddProductPage"))
		r.Post("/products/add", stub("AddProduct"))
		r.Get("/products/update/{id}", stub("GetUpdateProductPage"))
		r.Post("/products/update/{id}", stub("UpdateProduct"))
		r.Post("/products/delete", stub("RemoveProduct"))
		r.Post("/products", stub("RedirectProductsPost"))
		r.Get("/customers", stub("GetCustomerDetail"))
		r.Get("/profileDisplay", stub("AdminProfileDisplay"))
		r.Post("/updateuser", stub("AdminUpdateUserProfile"))
	})

	return r, nil
}

func (a *Application) Run(ctx context.Context) error {
	if a == nil || a.server == nil {
		return errors.New("jtspringproject: Run called on uninitialized Application")
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("jtspringproject: server error: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("jtspringproject: graceful shutdown failed: %w", err)
		}
		return nil
	}
}

func Run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := defaultServerAddr
	if len(args) > 0 && args[0] != "" {
		addr = args[0]
	}

	app, err := NewApplication(ctx, addr)
	if err != nil {
		return fmt.Errorf("jtspringproject: initialize application: %w", err)
	}

	log.Printf("jtspringproject: starting server on %s", addr)
	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("jtspringproject: run application: %w", err)
	}
	log.Print("jtspringproject: server stopped cleanly")
	return nil
}
