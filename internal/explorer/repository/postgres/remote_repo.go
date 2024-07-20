package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
	"github.com/noelukwa/git-explorer/internal/explorer/repository/postgres/sqlc"
)

type RemoteRepositoryImpl struct {
	queries *sqlc.Queries
	conn    *pgx.Conn
}

func (r *RemoteRepositoryImpl) SaveManyCommit(ctx context.Context, repoID int64, commits []models.Commit) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	for _, commit := range commits {
		author, err := qtx.GetAuthor(ctx, commit.Author.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		if errors.Is(err, pgx.ErrNoRows) {
			author, err = qtx.SaveAuthor(ctx, sqlc.SaveAuthorParams{
				ID:       commit.Author.ID,
				Name:     commit.Author.Name,
				Email:    commit.Author.Email,
				Username: commit.Author.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to save author %s: %w", commit.Author.Username, err)
			}
		}

		err = qtx.SaveCommit(ctx, sqlc.SaveCommitParams{
			Hash:         commit.Hash,
			AuthorID:     author.ID,
			CreatedAt:    pgtype.Timestamptz{Time: commit.CreatedAt, Valid: true},
			Message:      commit.Message,
			RepositoryID: repoID,
		})
		if err != nil {
			return fmt.Errorf("failed to save commit %s: %w", commit.Hash, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func newRemoteRepository(conn *pgx.Conn) repository.RemoteRepository {
	return &RemoteRepositoryImpl{queries: sqlc.New(conn), conn: conn}
}

func (r *RemoteRepositoryImpl) SaveRepo(ctx context.Context, repo *models.Repository) error {
	var createdAt, updatedAt pgtype.Timestamptz
	createdAt.Time = repo.CreatedAt
	createdAt.Valid = true
	updatedAt.Time = repo.UpdatedAt
	updatedAt.Valid = true

	return r.queries.SaveRepo(ctx, sqlc.SaveRepoParams{
		ID:         repo.ID,
		Watchers:   int32(repo.Watchers),
		Stargazers: int32(repo.StarGazers),
		FullName:   repo.FullName,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Language:   pgtype.Text{String: repo.Language, Valid: true},
		Forks:      int32(repo.Forks),
	})
}

func (r *RemoteRepositoryImpl) GetRepo(ctx context.Context, name string) (*models.Repository, error) {
	repo, err := r.queries.GetRepo(ctx, name)
	if err != nil {
		return nil, err
	}

	return &models.Repository{
		ID:         repo.ID,
		Watchers:   repo.Watchers,
		StarGazers: repo.Stargazers,
		FullName:   repo.FullName,
		CreatedAt:  repo.CreatedAt.Time,
		UpdatedAt:  repo.UpdatedAt.Time,
		Language:   repo.Language.String,
		Forks:      repo.Forks,
	}, nil
}

func (r *RemoteRepositoryImpl) FindCommits(ctx context.Context, filter repository.CommitsFilter, pagination repository.Pagination) (repository.PaginatedResponse[models.Commit], error) {
	var startDate, endDate pgtype.Timestamptz

	// Set startDate if filter.StartDate is provided and not zero
	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		startDate.Time = *filter.StartDate
		startDate.Valid = true
	}

	// Set endDate if filter.EndDate is provided and not zero
	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		endDate.Time = *filter.EndDate
		endDate.Valid = true
	}

	// Execute the FindCommits query
	rows, err := r.queries.FindCommits(ctx, sqlc.FindCommitsParams{
		FullName: filter.RepositoryName,
		Column2:  startDate,
		Column3:  endDate,
		Limit:    int32(pagination.PerPage),
		Offset:   int32((pagination.Page - 1) * pagination.PerPage),
	})
	if err != nil {
		return repository.PaginatedResponse[models.Commit]{}, err
	}

	var commits []models.Commit
	for _, row := range rows {
		commits = append(commits, models.Commit{
			Hash:      row.Hash,
			Message:   row.Message,
			Url:       parseURL(row.Url),
			CreatedAt: row.CreatedAt.Time,
			Repository: models.Repository{
				ID:         row.RepoID,
				Watchers:   row.Watchers,
				StarGazers: row.Stargazers,
				FullName:   row.Repository,
				CreatedAt:  row.RepoCreatedAt.Time,
				UpdatedAt:  row.RepoUpdatedAt.Time,
				Language:   row.Language.String,
				Forks:      row.Forks,
			},
			Author: models.Author{
				ID:       row.AuthorID,
				Name:     row.AuthorName,
				Email:    row.AuthorEmail,
				Username: row.AuthorUsername,
			},
		})
	}

	// Get the total count of commits matching the filter
	totalCount, err := r.queries.CountCommits(ctx, sqlc.CountCommitsParams{
		FullName: filter.RepositoryName,
		Column2:  startDate,
		Column3:  endDate,
	})
	if err != nil {
		return repository.PaginatedResponse[models.Commit]{}, err
	}

	return repository.PaginatedResponse[models.Commit]{
		Data:       commits,
		TotalCount: totalCount,
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
	}, nil
}

func (r *RemoteRepositoryImpl) GetTopCommitters(ctx context.Context, repository string, startDate, endDate *time.Time, pagination repository.Pagination) ([]models.AuthorStats, error) {
	var start, end pgtype.Timestamptz
	if startDate != nil {
		start.Time = *startDate
		start.Valid = true
	}
	if endDate != nil {
		end.Time = *endDate
		end.Valid = true
	}

	rows, err := r.queries.GetTopCommitters(ctx, sqlc.GetTopCommittersParams{
		FullName: repository,
		Column2:  start,
		Column3:  end,
		Limit:    int32(pagination.PerPage),
		Offset:   int32((pagination.Page - 1) * pagination.PerPage),
	})
	if err != nil {
		return nil, err
	}

	var stats []models.AuthorStats
	for _, row := range rows {
		stats = append(stats, models.AuthorStats{
			Author: models.Author{
				ID:       row.ID,
				Name:     row.Name,
				Email:    row.Email,
				Username: row.Username,
			},
			Commits: row.CommitCount,
		})
	}

	return stats, nil
}

func (r *RemoteRepositoryImpl) SaveAuthor(ctx context.Context, author models.Author) error {
	_, err := r.queries.SaveAuthor(ctx, sqlc.SaveAuthorParams{
		ID:       author.ID,
		Name:     author.Name,
		Email:    author.Email,
		Username: author.Username,
	})
	return err
}

func stringOrNull(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func parseURL(rawURL pgtype.Text) *url.URL {

	if rawURL.Valid {
		u, _ := url.Parse(rawURL.String)
		return u
	}
	return nil
}
