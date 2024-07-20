package inmem_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
	inmem "github.com/noelukwa/git-explorer/internal/explorer/repository/in-mem"
	"github.com/stretchr/testify/assert"
)

func TestSaveRepo(t *testing.T) {
	repo := models.Repository{FullName: "test/repo", ID: 1}
	r := inmem.NewRepositoryFactory().RemoteRepository()

	err := r.SaveRepo(context.Background(), &repo)
	assert.NoError(t, err)

	savedRepo, err := r.GetRepo(context.Background(), "test/repo")
	assert.NoError(t, err)
	assert.Equal(t, repo, *savedRepo)
}

func TestSaveManyCommit(t *testing.T) {
	repo := models.Repository{FullName: "test/repo", ID: 1}
	commit := models.Commit{Hash: "123", Author: models.Author{Username: "author1"}, CreatedAt: time.Now()}

	r := inmem.NewRepositoryFactory().RemoteRepository()
	r.SaveRepo(context.Background(), &repo)

	err := r.SaveManyCommit(context.Background(), repo.ID, []models.Commit{commit})
	assert.NoError(t, err)

	commits, err := r.FindCommits(context.Background(), repository.CommitsFilter{RepositoryName: "test/repo"}, repository.Pagination{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(commits.Data))
	assert.Equal(t, commit.Hash, commits.Data[0].Hash)
}

func TestGetTopCommitters(t *testing.T) {
	repo := models.Repository{FullName: "test/repo", ID: 1}
	commit1 := models.Commit{Hash: "123", Author: models.Author{Username: "author1"}, CreatedAt: time.Now()}
	commit2 := models.Commit{Hash: "124", Author: models.Author{Username: "author2"}, CreatedAt: time.Now()}
	commit3 := models.Commit{Hash: "125", Author: models.Author{Username: "author1"}, CreatedAt: time.Now()}

	r := inmem.NewRepositoryFactory().RemoteRepository()
	r.SaveRepo(context.Background(), &repo)
	r.SaveManyCommit(context.Background(), repo.ID, []models.Commit{commit1, commit2, commit3})

	stats, err := r.GetTopCommitters(context.Background(), "test/repo", nil, nil, repository.Pagination{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(stats))
	assert.Equal(t, "author1", stats[0].Author.Username)
	assert.Equal(t, int64(2), stats[0].Commits)
	assert.Equal(t, "author2", stats[1].Author.Username)
	assert.Equal(t, int64(1), stats[1].Commits)
}

func TestFindCommits(t *testing.T) {
	repo := models.Repository{FullName: "test/repo", ID: 1}
	commit1 := models.Commit{Hash: "123", Author: models.Author{Username: "author1"}, CreatedAt: time.Now()}
	commit2 := models.Commit{Hash: "124", Author: models.Author{Username: "author2"}, CreatedAt: time.Now()}

	r := inmem.NewRepositoryFactory().RemoteRepository()
	r.SaveRepo(context.Background(), &repo)
	r.SaveManyCommit(context.Background(), repo.ID, []models.Commit{commit1, commit2})

	filter := repository.CommitsFilter{RepositoryName: "test/repo"}
	pagination := repository.Pagination{Page: 1, PerPage: 10}
	response, err := r.FindCommits(context.Background(), filter, pagination)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Data))
}

func TestSaveIntent(t *testing.T) {
	r := inmem.NewRepositoryFactory().IntentRepository()

	intent := &models.Intent{
		ID:         uuid.New(),
		Repository: "test/repo",
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	err := r.SaveIntent(context.Background(), intent)
	assert.NoError(t, err)

	savedIntent, err := r.GetIntentById(context.Background(), intent.ID)
	assert.NoError(t, err)
	assert.Equal(t, intent, savedIntent)
}

func TestSaveIntent_EmptyID(t *testing.T) {
	r := inmem.NewRepositoryFactory().IntentRepository()

	intent := &models.Intent{
		ID:         uuid.Nil,
		Repository: "test/repo",
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	err := r.SaveIntent(context.Background(), intent)
	assert.Error(t, err)
	assert.Equal(t, "intent ID cannot be empty", err.Error())
}

func TestGetIntentById(t *testing.T) {
	r := inmem.NewRepositoryFactory().IntentRepository()

	intent := &models.Intent{
		ID:         uuid.New(),
		Repository: "test/repo",
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	r.SaveIntent(context.Background(), intent)

	savedIntent, err := r.GetIntentById(context.Background(), intent.ID)
	assert.NoError(t, err)
	assert.Equal(t, intent, savedIntent)

	nonExistentIntent, err := r.GetIntentById(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, nonExistentIntent)
}

func TestGetIntentByRepo(t *testing.T) {
	r := inmem.NewRepositoryFactory().IntentRepository()

	intent := &models.Intent{
		ID:         uuid.New(),
		Repository: "test/repo",
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	r.SaveIntent(context.Background(), intent)

	savedIntent, err := r.GetIntentByRepo(context.Background(), "test/repo")
	assert.NoError(t, err)
	assert.Equal(t, intent, savedIntent)

	nonExistentIntent, err := r.GetIntentByRepo(context.Background(), "non/existent/repo")
	assert.NoError(t, err)
	assert.Nil(t, nonExistentIntent)
}

func TestGetIntents(t *testing.T) {
	r := inmem.NewRepositoryFactory().IntentRepository()

	intent1 := &models.Intent{
		ID:         uuid.New(),
		Repository: "test/repo1",
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		IsActive:   true,
	}
	intent2 := &models.Intent{
		ID:         uuid.New(),
		Repository: "test/repo2",
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		IsActive:   false,
	}

	r.SaveIntent(context.Background(), intent1)
	r.SaveIntent(context.Background(), intent2)

	filter := repository.IntentFilter{IsActive: true}
	activeIntents, err := r.GetIntents(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(activeIntents))
	assert.Equal(t, intent1, activeIntents[0])

	filter = repository.IntentFilter{IsActive: false}
	inactiveIntents, err := r.GetIntents(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(inactiveIntents))
	assert.Equal(t, intent2, inactiveIntents[0])
}

func TestUpdateIntent(t *testing.T) {
	r := inmem.NewRepositoryFactory().IntentRepository()

	intent := &models.Intent{
		ID:         uuid.New(),
		Repository: "test/repo",
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	r.SaveIntent(context.Background(), intent)

	newSince := time.Now().Add(-time.Hour)
	update := &models.IntentUpdate{
		ID:       intent.ID,
		IsActive: false,
		Since:    &newSince,
	}

	err := r.UpdateIntent(context.Background(), update)
	assert.NoError(t, err)

	updatedIntent, err := r.GetIntentById(context.Background(), intent.ID)
	assert.NoError(t, err)
	assert.Equal(t, update.IsActive, updatedIntent.IsActive)
	assert.Equal(t, *update.Since, updatedIntent.Since)
}

func TestUpdateIntent_NonExistent(t *testing.T) {
	r := inmem.NewRepositoryFactory().IntentRepository()

	update := &models.IntentUpdate{
		ID:       uuid.New(),
		IsActive: false,
	}

	err := r.UpdateIntent(context.Background(), update)
	assert.Error(t, err)
	assert.Equal(t, "intent not found", err.Error())
}
