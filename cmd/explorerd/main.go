package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/noelukwa/git-explorer/internal/explorerd/service"
	"github.com/noelukwa/git-explorer/internal/pkg/config"
	"github.com/noelukwa/git-explorer/internal/pkg/github"
	"github.com/noelukwa/git-explorer/internal/pkg/messaging"
)

type RepositoryIntent struct {
	Repo        string    `json:"repo"`
	Since       time.Time `json:"since"`
	Until       time.Time `json:"until"`
	LastFetched time.Time `json:"last_fetched"`
}

func main() {
	var cfg config.ExplorerdConfig

	err := envconfig.Process("explorerd", &cfg)
	if err != nil {
		log.Fatalf("failed to process config: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := messaging.NewNATSClient(ctx, cfg.MessagingURL)
	if err != nil {
		log.Fatalf("Failed to connect to messaging system: %v", err)
	}
	defer nc.Close()

	kv, err := nc.NeKV(ctx, "intent_events")
	if err != nil {
		log.Fatal(err)
	}

	consumer, err := nc.NewConsumer(ctx, []string{"intent.*"}, "explorerd")
	if err != nil {
		log.Fatal(err)
	}

	gc := github.NewClient(cfg.GithubToken)
	daemon := service.NewService(nc, cfg.MonitoringInterval, kv, gc)

	errChan := make(chan error, 2)

	go func() {
		if err := daemon.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := daemon.Subcribe(ctx, consumer, nc, kv); err != nil {
			errChan <- err
		}
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-shutdownSignals:
			log.Printf("Received signal: %v", sig)
			cancel()
			errChan <- nil
		case <-ctx.Done():
			errChan <- ctx.Err()
		}
	}()

	// Wait for either an error or a shutdown signal
	if err := <-errChan; err != nil && err != context.Canceled {
		log.Fatalf("Error: %v", err)
	} else {
		log.Println("gracefully shut down")
	}
}
