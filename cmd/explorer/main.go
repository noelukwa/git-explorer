package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
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

	mc, err := messaging.NewMessagingClient(
		cfg.MessagingProvider,
		cfg.MessagingURL,
	)

	if err != nil {
		log.Fatalf("failed to connect to messaging system: %v", err)
	}
	defer mc.Close()

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
		mc,
	)
	repoService := service.NewRemoteRepoService(
		pgStore.RemoteRepository(),
	)

	router := api.SetupRoutes(intentService, repoService)

	log.Fatal(http.ListenAndServe(":8080", router))
}
