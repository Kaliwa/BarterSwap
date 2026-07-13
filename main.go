package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run wires the layers together and serves until interrupted.
func run() error {
	dsn := env("DATABASE_URL", "postgres://barterswap:barterswap@localhost:5432/barterswap?sslmode=disable")
	port := env("PORT", "8080")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := NewStore(ctx, dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	server := NewServer(NewApp(store))

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("arrêt en cours...")
		shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	log.Printf("BarterSwap API à l'écoute sur :%s", port)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// env returns the environment variable or a fallback when unset.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
