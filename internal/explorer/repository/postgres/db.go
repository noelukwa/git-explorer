package postgres

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

type pgStore struct {
	intentsRepo repository.IntentRepository
	remoteRepo  repository.RemoteRepository
}

func (p *pgStore) RemoteRepository() repository.RemoteRepository {
	return p.remoteRepo
}

func (p *pgStore) IntentRepository() repository.IntentRepository {
	return p.intentsRepo
}

func NewStore(conn *pgx.Conn) (repository.RepositoryFactory, error) {
	store := &pgStore{
		intentsRepo: newIntentRepository(conn),
		remoteRepo:  newRemoteRepository(conn),
	}

	log.Println("running database migrations...")
	if err := store.runMigrate(conn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (p *pgStore) runMigrate(conn *pgx.Conn) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Printf("failed to set goose dialect: %v", err)
		return err
	}

	config, err := pgxpool.ParseConfig(conn.Config().ConnString())
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	db := pool.Config().ConnConfig.ConnString()

	dbConn, err := goose.OpenDBWithDriver("pgx", db)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer dbConn.Close()

	if err := goose.Up(dbConn, "migrations"); err != nil {
		log.Printf("failed to run goose migrations: %v", err)
		return err
	}

	return nil
}
