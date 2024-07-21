package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	octo "github.com/google/go-github/v63/github"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/noelukwa/git-explorer/internal/events"
	"github.com/noelukwa/git-explorer/internal/pkg/config"
	"github.com/noelukwa/git-explorer/internal/pkg/github"
	"github.com/noelukwa/git-explorer/internal/pkg/messaging"
)

type RepositoryIntent struct {
	Repo        string    `json:"repo"`
	Since       time.Time `json:"since"`
	Until       time.Time `json:"until"`
	LastFetched time.Time `json:"last_fetched"`
}

func main() {
	var cfg config.ExplorerdConfig

	err := envconfig.Process("explorer", &cfg)
	if err != nil {
		log.Fatalln(err)
	}

	nc, err := messaging.NewNATSClient(cfg.MessagingURL)
	if err != nil {
		log.Fatalf("failed to connect to messaging system: %v", err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc.Conn())
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	xctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kv, _ := js.CreateKeyValue(xctx, jetstream.KeyValueConfig{
		Bucket: "intents",
	})

	gc := github.NewClient(cfg.GithubToken)

	if err := nc.Subscribe(events.NEW_INTENT, newIntentHandler(ctx, nc, kv, gc)); err != nil {
		log.Fatalf("failed to subscribe to event: %v", err)
	}

	go startMonitoring(ctx, nc, kv, gc, cfg.MonitoringInterval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(5 * time.Second)
}

func publishCommits(nc *messaging.NATSClient, repo string, since time.Time, commits []events.Commit) error {
	var commitList []events.Commit
	for _, commit := range commits {
		log.Printf("commit: %s\nrepo:%s\n", commit.Hash, repo)
		commitList = append(commitList, events.Commit{
			Hash: commit.Hash,
			Author: events.Author{
				Name:     commit.Author.Name,
				Email:    commit.Author.Email,
				Username: commit.Author.Username,
				ID:       commit.Author.ID,
			},
			Message:   commit.Message,
			URL:       commit.URL,
			CreatedAt: commit.CreatedAt,
		})
	}

	event := events.NewCommitsInfoEvent{
		Repository: repo,
		Since:      since,
		Commits:    commitList,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return nc.Publish(messaging.Event{Subject: events.NEW_COMMITS, Data: data})
}

func convertCommits(githubCommits []*octo.RepositoryCommit) []events.Commit {
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

func newIntentHandler(ctx context.Context, nc *messaging.NATSClient, kv jetstream.KeyValue, gc *github.Client) func(payload []byte) error {
	return func(payload []byte) error {
		var event events.NewIntentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("error unmarshalling payload: %v", err)
			return err
		}

		intentKey := "intent:" + event.Repository
		var intent RepositoryIntent

		existingIntent, err := kv.Get(ctx, intentKey)
		if err == nil {
			if err := json.Unmarshal(existingIntent.Value(), &intent); err != nil {
				log.Printf("error unmarshalling existing intent: %v", err)
				return err
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
			log.Printf("error marshalling intent: %v", err)
			return err
		}
		_, err = kv.Put(ctx, intentKey, intentJSON)
		if err != nil {
			log.Printf("error storing intent: %v", err)
			return err
		}

		err = fetchAndPublishRepoInfo(nc, gc, intent.Repo)
		if err != nil {
			log.Printf("error fetching and publishing repo info: %v", err)
			return err
		}

		// Trigger an immediate fetch
		return fetchAndPublishCommits(ctx, nc, kv, gc, intent)
	}
}

func startMonitoring(ctx context.Context, nc *messaging.NATSClient, kv jetstream.KeyValue, gc *github.Client, interval time.Duration) {
	log.Printf("startMonitoring every: %s", interval)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("checking for repos")
			keys, err := kv.ListKeys(ctx)
			if err != nil {
				log.Printf("error fetching keys: %v", err)
				continue
			}

			for key := range keys.Keys() {
				log.Println(key)
				if strings.HasPrefix(key, "intent:") {

					entry, err := kv.Get(ctx, key)
					if err != nil {
						log.Printf("error fetching intent for key %s: %v", key, err)
						continue
					}

					var intent RepositoryIntent
					if err := json.Unmarshal(entry.Value(), &intent); err != nil {
						log.Printf("error unmarshalling intent for key %s: %v", key, err)
						continue
					}

					if err := fetchAndPublishCommits(ctx, nc, kv, gc, intent); err != nil {
						log.Printf("error fetching and publishing commits for %s: %v", intent.Repo, err)
					}
				}
			}
		}
	}
}

func fetchAndPublishRepoInfo(nc *messaging.NATSClient, gc *github.Client, fullRepo string) error {
	parts := strings.Split(fullRepo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", fullRepo)
	}
	owner, repo := parts[0], parts[1]

	repoInfo, err := gc.FetchRepo(owner, repo)
	if err != nil {
		return fmt.Errorf("error fetching repo info: %w", err)
	}

	event := &events.NewRepoInfoEvent{
		Watchers:   int32(repoInfo.GetWatchersCount()),
		StarGazers: int32(repoInfo.GetStargazersCount()),
		FullName:   repoInfo.GetFullName(),
		ID:         repoInfo.GetID(),
		CreatedAt:  repoInfo.GetCreatedAt().Time,
		UpdatedAt:  repoInfo.GetUpdatedAt().Time,
		Language:   repoInfo.GetLanguage(),
		Forks:      int32(repoInfo.GetForksCount()),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling repo info: %w", err)
	}

	if err := nc.Publish(messaging.Event{Subject: events.NEW_REPO, Data: data}); err != nil {
		return fmt.Errorf("error publishing repo info: %w", err)
	}

	return nil
}

func fetchAndPublishCommits(ctx context.Context, nc *messaging.NATSClient, kv jetstream.KeyValue, gc *github.Client, intent RepositoryIntent) error {
	repo := strings.Split(intent.Repo, "/")

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

	commits, err := gc.FetchCommits(repo[0], repo[1], since, until)
	if err != nil {
		return fmt.Errorf("error fetching commits for %s: %w", intent.Repo, err)
	}

	if len(commits) == 0 {
		return nil
	}

	convertedCommits := convertCommits(commits)

	// Filter commits based on the Until time if it's set
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

	err = publishCommits(nc, intent.Repo, intent.LastFetched, convertedCommits)
	if err != nil {
		return fmt.Errorf("error publishing commits for %s: %w", intent.Repo, err)
	}

	// Update the intent in the KV store
	intentJSON, _ := json.Marshal(intent)
	_, err = kv.Put(ctx, "intent:"+intent.Repo, intentJSON)
	if err != nil {
		return fmt.Errorf("error storing updated intent for %s: %w", intent.Repo, err)
	}

	return nil
}
