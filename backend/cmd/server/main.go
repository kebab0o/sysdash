package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kebab0o/sysdash/backend/internal/collect"
	api "github.com/kebab0o/sysdash/backend/internal/http"
	"github.com/kebab0o/sysdash/backend/internal/store"
)

func main() {
	mem := store.NewMemory()
	app := &api.App{Store: mem}
	srv := api.NewServer(app.Routes())

	// start collectors
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	go collect.Start(ctx, mem, 30*time.Second)

	// daily prune for retention
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				mem.PruneForRetention()
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("listening on %s", addr)
	server := &http.Server{Addr: addr, Handler: srv.Router}

	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shCtx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
