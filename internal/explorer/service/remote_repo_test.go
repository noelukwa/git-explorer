package service_test

import (
	"context"

	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
	"github.com/stretchr/testify/mock"
)

type MockGitRemoteRepository struct {
	mock.Mock
}

func (m *MockGitRemoteRepository) SaveRepo(ctx context.Context, repo *models.Repository) error {
	args := m.Called(ctx, repo)
	return args.Error(0)
}

func (m *MockGitRemoteRepository) GetRepo(ctx context.Context, id string) (*models.Repository, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Repository), args.Error(1)
}

func (m *MockGitRemoteRepository) UpdateRepo(ctx context.Context, repo *models.Repository) error {
	args := m.Called(ctx, repo)
	return args.Error(0)
}

func (m *MockGitRemoteRepository) FindRepos(ctx context.Context, pagination repository.Pagination) (repository.PaginatedResponse[models.Repository], error) {
	args := m.Called(ctx, pagination)
	return args.Get(0).(repository.PaginatedResponse[models.Repository]), args.Error(1)
}

func (m *MockGitRemoteRepository) FindCommits(ctx context.Context, filter repository.CommitsFilter, pagination repository.Pagination) (repository.PaginatedResponse[models.Commit], error) {
	args := m.Called(ctx, filter, pagination)
	return args.Get(0).(repository.PaginatedResponse[models.Commit]), args.Error(1)
}

func (m *MockGitRemoteRepository) SaveManyCommits(ctx context.Context, repoName string, commits []models.Commit) error {
	args := m.Called(ctx, commits)
	return args.Error(0)
}

func (m *MockGitRemoteRepository) GetCommit(ctx context.Context, hash string) (*models.Commit, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*models.Commit), args.Error(1)
}
