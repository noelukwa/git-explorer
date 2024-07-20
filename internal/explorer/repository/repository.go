package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
)

type PaginatedResponse[T any] struct {
	Data       []T
	TotalCount int64
	Page       int
	PerPage    int
}

type Pagination struct {
	Page    int
	PerPage int
}

type IntentFilter struct {
	IsActive bool
}

type IntentRepository interface {
	SaveIntent(ctx context.Context, intent *models.Intent) error
	GetIntentById(ctx context.Context, id uuid.UUID) (*models.Intent, error)
	GetIntentByRepo(ctx context.Context, repo string) (*models.Intent, error)
	UpdateIntent(ctx context.Context, update *models.IntentUpdate) error
	GetIntents(ctx context.Context, filter IntentFilter) ([]*models.Intent, error)
}

type GroupAbleCol string

const (
	GroupByAuthor GroupAbleCol = "author_id"
	CreatedAt     GroupAbleCol = "created_at"
)

type CommitsFilter struct {
	RepositoryName string
	StartDate      *time.Time
	EndDate        *time.Time
}

type RemoteRepository interface {
	SaveRepo(ctx context.Context, repo *models.Repository) error
	GetRepo(ctx context.Context, name string) (*models.Repository, error)
	FindCommits(ctx context.Context, filter CommitsFilter, pagination Pagination) (PaginatedResponse[models.Commit], error)
	GetTopCommitters(ctx context.Context, repository string, startDate, endDate *time.Time, pagination Pagination) ([]models.AuthorStats, error)
	SaveManyCommit(ctx context.Context, repoID int64, commit []models.Commit) error
	SaveAuthor(ctx context.Context, author models.Author) error
}

type RepositoryFactory interface {
	IntentRepository() IntentRepository
	RemoteRepository() RemoteRepository
}
