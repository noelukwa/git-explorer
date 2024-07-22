package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/noelukwa/git-explorer/internal/events"
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

	mc, err := messaging.NewClient(cfg.MessagingURL)
	if err != nil {
		log.Fatalf("Failed to connect to messaging system: %v", err)
	}
	defer mc.Close()

	err = mc.DeclareQueue("gitexpress")
	if err != nil {
		log.Fatalf("Failed to connect to messaging system: %v", err)
	}

	gc := github.NewClient(cfg.GithubToken)
	_ = service.NewService(cfg.MonitoringInterval, gc)

	errChan := make(chan error, 2)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	id := 1
	go func() {
		for range ticker.C {
			message := "hello " + " " + time.Now().UTC().String() + " id " + strconv.Itoa(id)
			err = mc.Publish(ctx, "gitexpress", events.NEW_REPO_INTENT, message)
			if err != nil {
				log.Printf("Failed to publish message: %s", err)
			}
			log.Printf("Published message: %s", message)
			id++
		}
	}()

	// go func() {
	// 	if err := mc.Subscribe(ctx, "gitexpress", func(ctx context.Context, ek events.EventKind, b []byte) {

	// 	}); err != nil {
	// 		errChan <- err
	// 	}
	// }()

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

	if err := <-errChan; err != nil && err != context.Canceled {
		log.Fatalf("Error: %v", err)
	} else {
		log.Println("gracefully shut down")
	}
}
