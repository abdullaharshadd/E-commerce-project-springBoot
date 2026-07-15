package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"migrated-app/internal/jtspringproject/jtspringproject"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	_ = ctx
	if err := jtspringproject.Run(ctx); err != nil {
		log.Fatal(err)
	}
}