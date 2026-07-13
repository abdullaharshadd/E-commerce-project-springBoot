package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"migrated-app/internal/jtspringproject/jtspringproject"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	srv, err := jtspringproject.NewApplication(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("starting server on %s\n", addr)
	if err := srv.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}