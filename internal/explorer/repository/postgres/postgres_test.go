package postgres_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
	"github.com/noelukwa/git-explorer/internal/explorer/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDsn = "postgres://explorer:explorer2025@localhost/explorer-test?sslmode=disable"
)
var (
	testDB *pgxpool.Pool
	store  repository.RepositoryFactory
)

func TestMain(m *testing.M) {

	var err error
	testDB, err = pgxpool.New(context.Background(), testDsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer testDB.Close()

	conn, err := testDB.Acquire(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	store, err = postgres.NewStore(conn.Conn())
	if err != nil {
		log.Fatalln(err)
	}
	code := m.Run()
	os.Exit(code)
}

func clearTables(t *testing.T) {
	t.Helper()
	tables := []string{"commits", "repositories", "authors", "intents"}
	_, err := testDB.Exec(context.Background(), "SET CONSTRAINTS ALL DEFERRED;")
	require.NoError(t, err)

	for _, table := range tables {
		_, err := testDB.Exec(context.Background(), "TRUNCATE TABLE "+table+" CASCADE;")
		require.NoError(t, err)
	}

	_, err = testDB.Exec(context.Background(), "SET CONSTRAINTS ALL IMMEDIATE;")
	require.NoError(t, err)
}

func TestSaveRepo(t *testing.T) {

	repo := &models.Repository{
		ID:         int64(1),
		Watchers:   10,
		StarGazers: 20,
		FullName:   "test/repo",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Language:   "Go",
		Forks:      5,
	}

	remoteRepo := store.RemoteRepository()
	err := remoteRepo.SaveRepo(context.Background(), repo)
	assert.NoError(t, err)
	clearTables(t)
}

func TestGetRepo(t *testing.T) {

	repo := &models.Repository{
		ID:         int64(1),
		Watchers:   10,
		StarGazers: 20,
		FullName:   "test/repo",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Language:   "Go",
		Forks:      5,
	}

	remoteRepo := store.RemoteRepository()
	err := remoteRepo.SaveRepo(context.Background(), repo)
	require.NoError(t, err)

	savedRepo, err := remoteRepo.GetRepo(context.Background(), "test/repo")
	assert.NoError(t, err)
	assert.Equal(t, repo.FullName, savedRepo.FullName)
	assert.Equal(t, repo.Language, savedRepo.Language)
	assert.Equal(t, repo.StarGazers, savedRepo.StarGazers)
	clearTables(t)
}

func TestFindCommits(t *testing.T) {
	clearTables(t)

	repo := &models.Repository{
		ID:         int64(1),
		Watchers:   10,
		StarGazers: 20,
		FullName:   "test/repo",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Language:   "Go",
		Forks:      5,
	}

	remoteRepo := store.RemoteRepository()
	err := remoteRepo.SaveRepo(context.Background(), repo)
	require.NoError(t, err)

	author1 := models.Author{ID: 1, Name: "Author1", Email: "author1@example.com", Username: "author1"}
	author2 := models.Author{ID: 2, Name: "Author2", Email: "author2@example.com", Username: "author2"}

	commits := []models.Commit{
		{
			Hash:      "hash4",
			Message:   "message1",
			CreatedAt: time.Now(),
			Author:    author1,
			Url:       nil,
		},
		{
			Hash:      "hash5",
			Message:   "message2",
			CreatedAt: time.Now(),
			Author:    author2,
			Url:       nil,
		},
	}

	err = remoteRepo.SaveManyCommit(context.Background(), repo.ID, commits)
	require.NoError(t, err)

	filter := repository.CommitsFilter{Repository: "test/repo"}
	pagination := repository.Pagination{Page: 1, PerPage: 10}
	response, err := remoteRepo.FindCommits(context.Background(), filter, pagination)
	assert.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, commits[0].Hash, response.Data[0].Hash)
	assert.Equal(t, commits[1].Hash, response.Data[1].Hash)
	assert.Equal(t, commits[0].Author.Username, response.Data[0].Author.Username)
	assert.Equal(t, commits[1].Author.Username, response.Data[1].Author.Username)
	clearTables(t)
}

func TestGetTopCommitters(t *testing.T) {
	clearTables(t)

	repo := &models.Repository{
		ID:         int64(1),
		Watchers:   10,
		StarGazers: 20,
		FullName:   "test/repo",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Language:   "Go",
		Forks:      5,
	}

	remoteRepo := store.RemoteRepository()
	err := remoteRepo.SaveRepo(context.Background(), repo)
	require.NoError(t, err)

	author1 := models.Author{ID: 1, Name: "Author1", Email: "author1@example.com", Username: "author1"}
	author2 := models.Author{ID: 2, Name: "Author2", Email: "author2@example.com", Username: "author2"}

	commits := []models.Commit{
		{
			Hash:      "hash8",
			Message:   "message1",
			CreatedAt: time.Now(),
			Author:    author1,
		},
		{
			Hash:      "hash20",
			Message:   "message2",
			CreatedAt: time.Now(),
			Author:    author2,
		},
		{
			Hash:      "hash30",
			Message:   "message3",
			CreatedAt: time.Now(),
			Author:    author1,
		},
	}

	err = remoteRepo.SaveManyCommit(context.Background(), repo.ID, commits)
	require.NoError(t, err)

	startDate := time.Now().Add(-7 * 24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)
	pagination := repository.Pagination{Page: 1, PerPage: 10}
	stats, err := remoteRepo.GetTopCommitters(context.Background(), "test/repo", &startDate, &endDate, pagination)
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "author1", stats[0].Author.Username)
	assert.Equal(t, int32(2), stats[0].Commits)
	assert.Equal(t, "author2", stats[1].Author.Username)
	assert.Equal(t, int32(1), stats[1].Commits)
	clearTables(t)
}

func TestSaveManyCommit(t *testing.T) {
	clearTables(t)

	repo := &models.Repository{
		ID:         int64(1),
		Watchers:   10,
		StarGazers: 20,
		FullName:   "test/repo",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Language:   "Go",
		Forks:      5,
	}

	remoteRepo := store.RemoteRepository()
	err := remoteRepo.SaveRepo(context.Background(), repo)
	require.NoError(t, err)

	author1 := models.Author{ID: 1, Name: "Author1", Email: "author1@example.com", Username: "author1"}
	author2 := models.Author{ID: 2, Name: "Author2", Email: "author2@example.com", Username: "author2"}

	commits := []models.Commit{
		{
			Hash:      "hash1",
			Message:   "message1",
			CreatedAt: time.Now(),
			Author:    author1,
		},
		{
			Hash:      "hash2",
			Message:   "message2",
			CreatedAt: time.Now(),
			Author:    author2,
		},
	}

	err = remoteRepo.SaveManyCommit(context.Background(), repo.ID, commits)
	assert.NoError(t, err)

	// Verify that commits were saved
	filter := repository.CommitsFilter{Repository: "test/repo"}
	pagination := repository.Pagination{Page: 1, PerPage: 10}
	response, err := remoteRepo.FindCommits(context.Background(), filter, pagination)
	assert.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, commits[0].Hash, response.Data[0].Hash)
	assert.Equal(t, commits[1].Hash, response.Data[1].Hash)

	clearTables(t)
}
