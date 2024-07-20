package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
	"github.com/noelukwa/git-explorer/internal/explorer/repository/postgres/sqlc"
)

type IntentRepositoryImpl struct {
	queries *sqlc.Queries
}

// GetIntentById implements repository.IntentRepository.
func (r *IntentRepositoryImpl) GetIntentById(ctx context.Context, id uuid.UUID) (*models.Intent, error) {
	intent, err := r.queries.GetIntentById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var since, createdAt time.Time
	if intent.Since.Valid {
		since = intent.Since.Time
	}
	if intent.CreatedAt.Valid {
		createdAt = intent.CreatedAt.Time
	}

	return &models.Intent{
		ID:         intent.ID,
		Repository: intent.Repository,
		Since:      since,
		CreatedAt:  createdAt,
		IsActive:   intent.IsActive,
	}, nil
}

// GetIntentByRepo implements repository.IntentRepository.
func (r *IntentRepositoryImpl) GetIntentByRepo(ctx context.Context, repo string) (*models.Intent, error) {
	intent, err := r.queries.GetIntentByRepoName(ctx, repo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var since, createdAt time.Time
	if intent.Since.Valid {
		since = intent.Since.Time
	}
	if intent.CreatedAt.Valid {
		createdAt = intent.CreatedAt.Time
	}

	return &models.Intent{
		ID:         intent.ID,
		Repository: intent.Repository,
		Since:      since,
		CreatedAt:  createdAt,
		IsActive:   intent.IsActive,
	}, nil
}

func (r *IntentRepositoryImpl) GetIntents(ctx context.Context, filter repository.IntentFilter) ([]*models.Intent, error) {
	intents, err := r.queries.GetIntents(ctx, filter.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var result []*models.Intent
	for _, intent := range intents {
		result = append(result, &models.Intent{
			ID:         intent.ID,
			Repository: intent.Repository,
			Since:      intent.Since.Time,
			CreatedAt:  intent.CreatedAt.Time,
			IsActive:   intent.IsActive,
		})
	}

	return result, nil
}

func newIntentRepository(conn *pgx.Conn) repository.IntentRepository {
	return &IntentRepositoryImpl{queries: sqlc.New(conn)}
}
func (r *IntentRepositoryImpl) SaveIntent(ctx context.Context, intent *models.Intent) error {
	var since, createdAt pgtype.Timestamptz
	if !intent.Since.IsZero() {
		since.Time = intent.Since
		since.Valid = true
	}
	createdAt.Time = intent.CreatedAt
	createdAt.Valid = true

	err := r.queries.SaveIntent(ctx, sqlc.SaveIntentParams{
		ID:         intent.ID,
		Repository: intent.Repository,
		Since:      since,
		CreatedAt:  createdAt,
		IsActive:   intent.IsActive,
	})
	return err
}

func (r *IntentRepositoryImpl) UpdateIntent(ctx context.Context, update *models.IntentUpdate) error {
	var since pgtype.Timestamptz
	if !update.Since.IsZero() {
		since.Time = *update.Since
		since.Valid = true
	}

	err := r.queries.UpdateIntent(ctx, sqlc.UpdateIntentParams{
		ID:       update.ID,
		IsActive: update.IsActive,
		Since:    since,
	})
	return err
}
