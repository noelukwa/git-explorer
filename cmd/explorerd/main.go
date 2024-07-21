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
	"golang.org/x/sync/errgroup"
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

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		if err := daemon.Start(ctx); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		err := daemon.Subcribe(ctx, consumer, nc, kv)
		if err != nil {
			return err
		}
		return nil
	})

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-shutdownSignals:
			log.Printf("Received signal: %v", sig)
			cancel()
			return nil
		}
	})

	err = g.Wait()
	if err != nil && err != context.Canceled {
		log.Fatalln(err)
	}
	log.Println("gracefully shut down")
}
