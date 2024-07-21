package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v63/github"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/noelukwa/git-explorer/internal/events"
	octo "github.com/noelukwa/git-explorer/internal/pkg/github"
	"github.com/noelukwa/git-explorer/internal/pkg/messaging"
)

type RepositoryIntent struct {
	Repo        string
	Since       time.Time
	Until       time.Time
	LastFetched time.Time
}
type service struct {
	nc       *messaging.NATSClient
	interval time.Duration
	kv       jetstream.KeyValue
	gc       *octo.Client
}

func NewService(nc *messaging.NATSClient, interval time.Duration, kv jetstream.KeyValue, gc *octo.Client) *service {
	return &service{
		nc:       nc,
		interval: interval,
		kv:       kv,
		gc:       gc,
	}
}

func (svc *service) Start(ctx context.Context) error {
	ticker := time.NewTicker(svc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Println("checking for repos")
			keys, err := svc.kv.ListKeys(ctx)
			if err != nil {
				log.Printf("error fetching keys: %v", err)
				continue
			}

			for key := range keys.Keys() {
				if strings.HasPrefix(key, "intent:") {
					entry, err := svc.kv.Get(ctx, key)
					if err != nil {
						log.Printf("Error fetching intent for key %s: %v", key, err)
						continue
					}

					var intent RepositoryIntent
					if err := json.Unmarshal(entry.Value(), &intent); err != nil {
						log.Printf("Error unmarshalling intent for key %s: %v", key, err)
						continue
					}

					if err := svc.fetchAndPublishCommits(ctx, intent); err != nil {
						log.Printf("Error fetching and publishing commits for %s: %v", intent.Repo, err)
					}
				}
			}
		}
	}

}

func (svc *service) Subcribe(ctx context.Context, consumer jetstream.Consumer, nc *messaging.NATSClient, kv jetstream.KeyValue) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msgs, err := consumer.Fetch(1, jetstream.FetchMaxWait(5*time.Second))
			if err != nil {
				continue
			}

			for msg := range msgs.Messages() {
				log.Printf("received intent: %s", msg.Subject())
				if err := svc.handleNewIntent(ctx, msg.Data()); err != nil {
					log.Printf("Error handling intent: %v", err)
				}
				msg.Ack()
			}
		}
	}
}

func (svc *service) handleNewIntent(ctx context.Context, payload []byte) error {
	var event events.NewRepoIntentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("error unmarshalling payload: %w", err)
	}

	intentKey := "intent:" + event.Repository
	var intent RepositoryIntent

	existingIntent, err := svc.kv.Get(ctx, intentKey)
	if err == nil {
		if err := json.Unmarshal(existingIntent.Value(), &intent); err != nil {
			return fmt.Errorf("error unmarshalling existing intent: %w", err)
		}
	} else {
		intent = RepositoryIntent{
			Repo:        event.Repository,
			Since:       event.Since,
			LastFetched: event.Since,
		}
	}

	if !event.Since.IsZero() {
		if event.Since.Before(intent.Since) {
			intent.Since = event.Since
			intent.Until = intent.LastFetched
		} else {
			intent.Since = event.Since
			intent.Until = time.Time{}
		}
		intent.LastFetched = intent.Since
	}

	intentJSON, err := json.Marshal(intent)
	if err != nil {
		return fmt.Errorf("error marshalling intent: %w", err)
	}
	_, err = svc.kv.Put(ctx, intentKey, intentJSON)
	if err != nil {
		return fmt.Errorf("error storing intent: %w", err)
	}

	if err := svc.fetchAndPublishRepoInfo(ctx, intent.Repo); err != nil {
		return fmt.Errorf("error fetching and publishing repo info: %w", err)
	}

	return svc.fetchAndPublishCommits(ctx, intent)
}

func (svc *service) fetchAndPublishRepoInfo(ctx context.Context, fullRepo string) error {
	parts := strings.Split(fullRepo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", fullRepo)
	}
	owner, repo := parts[0], parts[1]

	gc := octo.NewClient("")

	repoInfo, err := gc.FetchRepo(owner, repo)
	if err != nil {
		return fmt.Errorf("error fetching repo info: %w", err)
	}

	event := &events.NewRepoDataEvent{
		Kind: events.NEW_REPO_DATA,
		Info: events.Repo{
			Watchers:   int32(repoInfo.GetWatchersCount()),
			StarGazers: int32(repoInfo.GetStargazersCount()),
			FullName:   repoInfo.GetFullName(),
			ID:         repoInfo.GetID(),
			CreatedAt:  repoInfo.GetCreatedAt().Time,
			UpdatedAt:  repoInfo.GetUpdatedAt().Time,
			Language:   repoInfo.GetLanguage(),
			Forks:      int32(repoInfo.GetForksCount()),
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling repo info: %w", err)
	}

	if err := svc.nc.Publish(ctx, messaging.Event{Subject: events.NEW_REPO_DATA, Data: data}); err != nil {
		return fmt.Errorf("error publishing repo info: %w", err)
	}

	return nil
}

func (svc *service) fetchAndPublishCommits(ctx context.Context, intent RepositoryIntent) error {
	repo := strings.Split(intent.Repo, "/")
	if len(repo) != 2 {
		return fmt.Errorf("invalid repository format: %s", intent.Repo)
	}

	minTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	maxTime := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

	since := intent.LastFetched
	if since.Before(minTime) {
		since = minTime
	}

	var until time.Time
	if !intent.Until.IsZero() {
		until = intent.Until
		if until.After(maxTime) {
			until = maxTime
		}
	}
	gc := octo.NewClient("")

	commits, err := gc.FetchCommits(repo[0], repo[1], since, until)
	if err != nil {
		return fmt.Errorf("error fetching commits for %s: %w", intent.Repo, err)
	}

	if len(commits) == 0 {
		return nil
	}

	convertedCommits := convertCommits(commits)

	if !intent.Until.IsZero() {
		var filteredCommits []events.Commit
		for _, commit := range convertedCommits {
			if commit.CreatedAt.Before(intent.Until) || commit.CreatedAt.Equal(intent.Until) {
				filteredCommits = append(filteredCommits, commit)
			} else {
				break
			}
		}
		convertedCommits = filteredCommits
	}

	if len(convertedCommits) > 0 {
		intent.LastFetched = convertedCommits[0].CreatedAt
	}

	event := events.NewCommitsDataEvent{
		Repository: intent.Repo,
		Since:      intent.LastFetched,
		Commits:    convertedCommits,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling commits: %w", err)
	}

	if err := svc.nc.Publish(ctx, messaging.Event{Subject: events.NEW_COMMITS_DATA, Data: data}); err != nil {
		return fmt.Errorf("error publishing commits: %w", err)
	}

	intentJSON, _ := json.Marshal(intent)
	_, err = svc.kv.Put(ctx, "intent:"+intent.Repo, intentJSON)
	if err != nil {
		return fmt.Errorf("error storing updated intent for %s: %w", intent.Repo, err)
	}

	return nil
}

func convertCommits(githubCommits []*github.RepositoryCommit) []events.Commit {
	var commits []events.Commit
	for _, commit := range githubCommits {
		if commit.Commit != nil && commit.Commit.Author != nil && commit.Commit.Committer != nil {
			commits = append(commits, events.Commit{
				Hash: commit.GetSHA(),
				Author: events.Author{
					Name:     commit.Commit.Author.GetName(),
					Email:    commit.Commit.Author.GetEmail(),
					Username: commit.Author.GetLogin(),
					ID:       commit.Author.GetID(),
				},
				Message:   commit.Commit.GetMessage(),
				URL:       commit.URL,
				CreatedAt: commit.Commit.Committer.GetDate().Time,
			})
		}
	}
	return commits
}
