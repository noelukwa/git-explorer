package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/noelukwa/git-explorer/internal/events"
	"github.com/noelukwa/git-explorer/internal/explorer/api"
	"github.com/noelukwa/git-explorer/internal/explorer/repository/postgres"
	"github.com/noelukwa/git-explorer/internal/explorer/service"
	"github.com/noelukwa/git-explorer/internal/pkg/config"
	"github.com/noelukwa/git-explorer/internal/pkg/messaging"
)

func main() {
	var cfg config.ExplorerConfig

	err := envconfig.Process("explorer", &cfg)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mc, err := messaging.NewClient(cfg.MessagingURL)
	if err != nil {
		log.Fatalf("failed to connect to messaging system: %v", err)
	}
	defer mc.Close()

	err = mc.DeclareQueue("gitexpress")
	if err != nil {
		log.Fatalf("unable to declare gitexpress queue: %v\n", err)
	}

	err = mc.DeclareQueue("gitintents")
	if err != nil {
		log.Fatalf("unable to declare gitintents queue: %v\n", err)
	}

	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	pgStore, err := postgres.NewStore(conn)
	if err != nil {
		log.Fatalln(err)
	}

	intentService := service.NewIntentService(
		pgStore.IntentRepository(),
	)

	repoService := service.NewRemoteRepoService(
		pgStore.RemoteRepository(),
	)

	router := api.SetupRoutes(intentService, repoService)

	httpServer := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", cfg.Port),
	}

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("starting HTTP server on %d", cfg.Port)
		serverErrors <- httpServer.ListenAndServe()
	}()

	go func() {
		if err := mc.Subscribe(ctx, "gitexpress", repoService.Process); err != nil {
			serverErrors <- err
		}
	}()

	// Producer: Publish to gitintents
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				message := fmt.Sprintf("Intent message at %s", time.Now().UTC())
				err := mc.Publish(ctx, "gitintents", events.NEW_REPO_INTENT, []byte(message))
				if err != nil {
					log.Printf("Failed to publish to gitintents: %v", err)
				} else {
					log.Printf("Published to gitintents: %s", message)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case sig := <-shutdownSignals:
		log.Printf("received termination signal: %s", sig.String())
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
}
