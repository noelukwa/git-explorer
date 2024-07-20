package github

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v63/github"
)

type Client struct {
	client *github.Client
}

func NewClient(token string) *Client {
	hc := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Client{client: github.NewClient(hc)}
}

func (c *Client) FetchCommits(owner, repo string, since, until time.Time) ([]*github.RepositoryCommit, error) {
	ctx := context.Background()
	opts := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
		Since: since,
		Until: until,
	}
	var allCommits []*github.RepositoryCommit
	for {
		commits, resp, err := c.client.Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allCommits = append(allCommits, commits...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allCommits, nil
}

func (c *Client) FetchRepo(owner, repo string) (*github.Repository, error) {
	ctx := context.Background()
	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	return repository, err
}
