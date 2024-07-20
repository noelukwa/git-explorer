package inmem

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
)

type RemoteRepository struct {
	repos   map[string]*models.Repository
	commits map[int64][]models.Commit
	mu      sync.RWMutex
}

// SaveAuthor implements repository.RemoteRepository.
func (r *RemoteRepository) SaveAuthor(ctx context.Context, author models.Author) error {
	panic("unimplemented")
}

func (r *RemoteRepository) SaveManyCommit(ctx context.Context, repoID int64, commits []models.Commit) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	found := false
	for _, repo := range r.repos {
		if repo.ID == repoID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("repository with ID %d not found", repoID)
	}

	r.commits[repoID] = append(r.commits[repoID], commits...)

	return nil
}

func (r *RemoteRepository) GetTopCommitters(ctx context.Context, repository string, startDate *time.Time, endDate *time.Time, pagination repository.Pagination) ([]models.AuthorStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repo, exists := r.repos[repository]
	if !exists {
		return nil, nil
	}

	repoCommits, exists := r.commits[repo.ID]
	if !exists {
		return nil, nil
	}

	committerStats := make(map[string]int64)
	for _, commit := range repoCommits {
		if (startDate == nil || !commit.CreatedAt.Before(*startDate)) &&
			(endDate == nil || !commit.CreatedAt.After(*endDate)) {
			committerStats[commit.Author.Username]++
		}
	}

	var authorStats []models.AuthorStats
	for username, count := range committerStats {
		authorStats = append(authorStats, models.AuthorStats{
			Author:  models.Author{Username: username},
			Commits: count,
		})
	}
	sort.Slice(authorStats, func(i, j int) bool {
		return authorStats[i].Commits > authorStats[j].Commits
	})

	start := (pagination.Page - 1) * pagination.PerPage
	end := start + pagination.PerPage
	if end > len(authorStats) {
		end = len(authorStats)
	}

	if start >= len(authorStats) {
		return nil, nil
	}

	return authorStats[start:end], nil
}

func newGitRemoteRepository() repository.RemoteRepository {
	return &RemoteRepository{
		repos:   make(map[string]*models.Repository),
		commits: make(map[int64][]models.Commit),
	}
}

func (r *RemoteRepository) SaveRepo(ctx context.Context, repo *models.Repository) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.repos[repo.FullName] = repo
	return nil
}

func (r *RemoteRepository) GetRepo(ctx context.Context, id string) (*models.Repository, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repo, exists := r.repos[id]
	if !exists {
		return nil, nil
	}

	return repo, nil
}

func (r *RemoteRepository) FindCommits(ctx context.Context, filter repository.CommitsFilter, pagination repository.Pagination) (repository.PaginatedResponse[models.Commit], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repo, exists := r.repos[filter.RepositoryName]
	if !exists {
		return repository.PaginatedResponse[models.Commit]{}, nil
	}
	repoCommits, exists := r.commits[repo.ID]
	if !exists {
		return repository.PaginatedResponse[models.Commit]{}, nil
	}

	var filteredCommits []models.Commit
	for _, commit := range repoCommits {
		if (filter.StartDate == nil || !commit.CreatedAt.Before(*filter.StartDate)) &&
			(filter.EndDate == nil || !commit.CreatedAt.Before(*filter.EndDate)) {
			filteredCommits = append(filteredCommits, commit)
		}
	}

	start := (pagination.Page - 1) * pagination.PerPage
	end := start + pagination.PerPage
	if end > len(filteredCommits) {
		end = len(filteredCommits)
	}

	if start >= len(filteredCommits) {
		return repository.PaginatedResponse[models.Commit]{}, nil
	}

	return repository.PaginatedResponse[models.Commit]{
		Data:       filteredCommits[start:end],
		TotalCount: int64(len(filteredCommits)),
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
	}, nil
}
