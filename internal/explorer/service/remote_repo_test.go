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

// func TestBatchSave(t *testing.T) {
// 	mockRepo := new(MockGitRemoteRepository)
// 	service := service.NewRemoteRepoService(mockRepo)

// 	ctx := context.Background()
// 	commits := []models.Commit{{Hash: "abc123"}}

// 	mockRepo.On("SaveManyCommits", ctx, commits).Return(nil)

// 	err := service.BatchSaveCommits(ctx, "test/repo", commits)
// 	assert.NoError(t, err)
// 	mockRepo.AssertExpectations(t)
// }

// func TestFindRepository(t *testing.T) {
// 	mockRepo := new(MockGitRemoteRepository)
// 	service := service.NewRemoteRepoService(mockRepo)

// 	ctx := context.Background()
// 	repoID := "test/repo"
// 	expectedRepo := &models.Repository{FullName: repoID}

// 	mockRepo.On("GetRepo", ctx, repoID).Return(expectedRepo, nil)

// 	repo, err := service.FindRepository(ctx, repoID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedRepo, repo)
// 	mockRepo.AssertExpectations(t)
// }

// func TestGetTopCommitters(t *testing.T) {
// 	mockRepo := new(MockGitRemoteRepository)
// 	service := service.NewRemoteRepoService(mockRepo)

// 	ctx := context.Background()
// 	repoID := "test/repo"
// 	limit := 2

// 	mockRepo.On("GetRepo", ctx, repoID).Return(&models.Repository{}, nil)
// 	mockRepo.On("FindCommits", ctx, mock.Anything, mock.Anything).Return(
// 		repository.PaginatedResponse[models.Commit]{
// 			Data: []models.Commit{
// 				{Author: models.Author{ID: 1, Username: "user1"}},
// 				{Author: models.Author{ID: 2, Username: "user2"}},
// 				{Author: models.Author{ID: 1, Username: "user1"}},
// 			},
// 		},
// 		nil,
// 	)

// 	topCommitters, err := service.GetTopCommitters(ctx, repoID, limit)
// 	assert.NoError(t, err)
// 	assert.Len(t, topCommitters, 2)
// 	assert.Equal(t, int64(1), topCommitters[0].Author.ID)
// 	assert.Equal(t, 2, topCommitters[0].Commits)
// 	assert.Equal(t, int64(2), topCommitters[1].Author.ID)
// 	assert.Equal(t, 1, topCommitters[1].Commits)
// 	mockRepo.AssertExpectations(t)
// }

// func TestGetCommits(t *testing.T) {
// 	mockRepo := new(MockGitRemoteRepository)
// 	service := service.NewRemoteRepoService(mockRepo)

// 	ctx := context.Background()
// 	repoID := "test/repo"
// 	startDate := time.Now().Add(-24 * time.Hour)
// 	endDate := time.Now()
// 	page := 1
// 	perPage := 10

// 	mockRepo.On("GetRepo", ctx, repoID).Return(&models.Repository{}, nil)
// 	mockRepo.On("FindCommits", ctx, mock.Anything, mock.Anything).Return(
// 		repository.PaginatedResponse[models.Commit]{
// 			Data:       []models.Commit{{Hash: "abc123"}},
// 			TotalCount: 1,
// 			Page:       page,
// 			PerPage:    perPage,
// 		},
// 		nil,
// 	)

// 	commitPage, err := service.GetCommits(ctx, repoID, startDate, endDate, page, perPage)
// 	assert.NoError(t, err)
// 	assert.Len(t, commitPage.Commits, 1)
// 	assert.Equal(t, int64(1), commitPage.TotalCount)
// 	assert.Equal(t, page, commitPage.Page)
// 	assert.Equal(t, perPage, commitPage.PerPage)
// 	mockRepo.AssertExpectations(t)
// }

// func TestErrorCases(t *testing.T) {
// 	mockRepo := new(MockGitRemoteRepository)
// 	service := service.NewRemoteRepoService(mockRepo)

// 	ctx := context.Background()
// 	repoName := "test/repo"

// 	t.Run("BatchSave error", func(t *testing.T) {
// 		mockRepo.On("SaveRepo", ctx, mock.Anything).Return(errors.New("save error"))
// 		err := service.BatchSaveCommits(ctx, repoName, nil)
// 		assert.Error(t, err)
// 	})

// 	t.Run("FindRepository error", func(t *testing.T) {
// 		mockRepo.On("GetRepo", ctx, repoName).Return((*models.Repository)(nil), errors.New("not found"))
// 		_, err := service.FindRepository(ctx, repoName)
// 		assert.Error(t, err)
// 	})

// 	t.Run("GetTopCommitters repo not found", func(t *testing.T) {
// 		mockRepo.On("GetRepo", ctx, repoName).Return((*models.Repository)(nil), errors.New("not found"))
// 		_, err := service.GetTopCommitters(ctx, repoName, 5)
// 		assert.Error(t, err)
// 	})

// 	t.Run("GetCommits repo not found", func(t *testing.T) {
// 		mockRepo.On("GetRepo", ctx, repoName).Return((*models.Repository)(nil), errors.New("not found"))
// 		_, err := service.GetCommits(ctx, repoName, time.Now(), time.Now(), 1, 10)
// 		assert.Error(t, err)
// 	})
// }
