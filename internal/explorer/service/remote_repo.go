package service

import (
	"context"
	"sort"
	"time"

	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
)

type RemoteRepoService interface {
	BatchSaveCommits(ctx context.Context, repoName string, commits []models.Commit) error
	FindRepository(ctx context.Context, repoName string) (*models.Repository, error)
	GetTopCommitters(ctx context.Context, repoName string, limit int) ([]models.AuthorStats, error)
	GetCommits(ctx context.Context, repoName string, startDate, endDate time.Time, page, perPage int) (models.CommitPage, error)
}

type remoteRepoService struct {
	repo repository.RemoteRepository
}

func (s *remoteRepoService) BatchSaveCommits(ctx context.Context, repoName string, commits []models.Commit) error {

	repository, err := s.repo.GetRepo(ctx, repoName)
	if err != nil || repository == nil {
		return err
	}
	return s.repo.SaveManyCommit(ctx, repository.ID, commits)
}

func (s *remoteRepoService) FindRepository(ctx context.Context, repoName string) (*models.Repository, error) {
	return s.repo.GetRepo(ctx, repoName)
}

func (s *remoteRepoService) GetTopCommitters(ctx context.Context, repoName string, limit int) ([]models.AuthorStats, error) {

	_, err := s.repo.GetRepo(ctx, repoName)
	if err != nil {
		return nil, err
	}

	filter := repository.CommitsFilter{
		RepositoryName: repoName,
	}

	pagination := repository.Pagination{
		Page:    1,
		PerPage: 1000,
	}

	commitsResp, err := s.repo.FindCommits(ctx, filter, pagination)
	if err != nil {
		return nil, err
	}

	authorCommits := make(map[int64]int)
	authorMap := make(map[int64]models.Author)

	for _, commit := range commitsResp.Data {
		authorCommits[commit.Author.ID]++
		authorMap[commit.Author.ID] = commit.Author
	}

	var topCommitters []models.AuthorStats
	for authorID, count := range authorCommits {
		topCommitters = append(topCommitters, models.AuthorStats{
			Author:  authorMap[authorID],
			Commits: int64(count),
		})
	}

	sort.Slice(topCommitters, func(i, j int) bool {
		return topCommitters[i].Commits > topCommitters[j].Commits
	})

	if limit > len(topCommitters) {
		limit = len(topCommitters)
	}

	return topCommitters[:limit], nil
}

func (s *remoteRepoService) GetCommits(ctx context.Context, repo string, startDate, endDate time.Time, page, perPage int) (models.CommitPage, error) {

	_, err := s.repo.GetRepo(ctx, repo)
	if err != nil {
		return models.CommitPage{}, err
	}

	filter := repository.CommitsFilter{
		RepositoryName: repo,
		StartDate:      &startDate,
		EndDate:        &endDate,
	}
	pagination := repository.Pagination{
		Page:    page,
		PerPage: perPage,
	}

	repoResp, err := s.repo.FindCommits(ctx, filter, pagination)
	if err != nil {
		return models.CommitPage{}, err
	}

	return models.CommitPage{
		Commits:    repoResp.Data,
		TotalCount: repoResp.TotalCount,
		Page:       int32(repoResp.Page),
		PerPage:    int32(repoResp.PerPage),
	}, nil
}

func NewRemoteRepoService(repo repository.RemoteRepository) RemoteRepoService {
	return &remoteRepoService{
		repo: repo,
	}
}
