package service

import (
	"time"

	octo "github.com/noelukwa/git-explorer/internal/pkg/github"
)

type RepositoryIntent struct {
	Repo        string
	Since       time.Time
	Until       time.Time
	LastFetched time.Time
}
type service struct {
	interval time.Duration
	gc       *octo.Client
}

func NewService(interval time.Duration, gc *octo.Client) *service {
	return &service{

		interval: interval,
		gc:       gc,
	}
}

// func (svc *service) Start(ctx context.Context) error {
// 	ticker := time.NewTicker(svc.interval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case <-ticker.C:
// 			log.Println("checking for repos")

// 		}
// 	}

// }

// func (svc *service) Subcribe(ctx context.Context,  mc *messaging.Client) error {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		default:

// 		}
// 	}
// }

// func (svc *service) handleNewIntent(ctx context.Context, payload []byte) error {
// 	var event events.NewRepoIntentEvent
// 	if err := json.Unmarshal(payload, &event); err != nil {
// 		return fmt.Errorf("error unmarshalling payload: %w", err)
// 	}

// 	intentKey := strings.ReplaceAll(event.Repository, "/", "_")
// 	var intent RepositoryIntent

// 	existingIntent, err := svc.kv.Get(ctx, intentKey)
// 	if err == nil {
// 		if err := json.Unmarshal(existingIntent.Value(), &intent); err != nil {
// 			return fmt.Errorf("error unmarshalling existing intent: %w", err)
// 		}
// 	} else {
// 		intent = RepositoryIntent{
// 			Repo:        event.Repository,
// 			Since:       event.Since,
// 			LastFetched: event.Since,
// 		}
// 	}

// 	if !event.Since.IsZero() {
// 		if event.Since.Before(intent.Since) {
// 			intent.Since = event.Since
// 			intent.Until = intent.LastFetched
// 		} else {
// 			intent.Since = event.Since
// 			intent.Until = time.Time{}
// 		}
// 		intent.LastFetched = intent.Since
// 	}

// 	intentJSON, err := json.Marshal(intent)
// 	if err != nil {
// 		return fmt.Errorf("error marshalling intent: %w", err)
// 	}
// 	_, err = svc.kv.Put(ctx, intentKey, intentJSON)
// 	if err != nil {
// 		return fmt.Errorf("error storing intent: %w", err)
// 	}

// 	if err := svc.fetchAndPublishRepoInfo(ctx, intent.Repo); err != nil {
// 		return fmt.Errorf("error fetching and publishing repo info: %w", err)
// 	}

// 	return svc.fetchAndPublishCommits(ctx, intent)
// }

// func (svc *service) fetchAndPublishRepoInfo(ctx context.Context, fullRepo string) error {
// 	parts := strings.Split(fullRepo, "/")
// 	if len(parts) != 2 {
// 		return fmt.Errorf("invalid repository format: %s", fullRepo)
// 	}
// 	owner, repo := parts[0], parts[1]

// 	gc := octo.NewClient("")

// 	repoInfo, err := gc.FetchRepo(owner, repo)
// 	if err != nil {
// 		return fmt.Errorf("error fetching repo info: %w", err)
// 	}

// 	event := &events.NewRepoDataEvent{

// 		Info: &models.Repository{
// 			Watchers:   int32(repoInfo.GetWatchersCount()),
// 			StarGazers: int32(repoInfo.GetStargazersCount()),
// 			FullName:   repoInfo.GetFullName(),
// 			ID:         repoInfo.GetID(),
// 			CreatedAt:  repoInfo.GetCreatedAt().Time,
// 			UpdatedAt:  repoInfo.GetUpdatedAt().Time,
// 			Language:   repoInfo.GetLanguage(),
// 			Forks:      int32(repoInfo.GetForksCount()),
// 		},
// 	}

// 	data, err := json.Marshal(event)
// 	if err != nil {
// 		return fmt.Errorf("error marshalling repo info: %w", err)
// 	}

// 	if err := svc.mc.Publish(ctx); err != nil {
// 		return fmt.Errorf("error publishing repo info: %w", err)
// 	}

// 	return nil
// }

// func (svc *service) fetchAndPublishCommits(ctx context.Context, intent RepositoryIntent) error {
// 	repo := strings.Split(intent.Repo, "/")
// 	if len(repo) != 2 {
// 		return fmt.Errorf("invalid repository format: %s", intent.Repo)
// 	}

// 	minTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
// 	maxTime := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

// 	since := intent.LastFetched
// 	if since.Before(minTime) {
// 		since = minTime
// 	}

// 	var until time.Time
// 	if !intent.Until.IsZero() {
// 		until = intent.Until
// 		if until.After(maxTime) {
// 			until = maxTime
// 		}
// 	}
// 	gc := octo.NewClient("")

// 	commits, err := gc.FetchCommits(repo[0], repo[1], since, until)
// 	if err != nil {
// 		return fmt.Errorf("error fetching commits for %s: %w", intent.Repo, err)
// 	}

// 	if len(commits) == 0 {
// 		return nil
// 	}

// 	convertedCommits := convertCommits(commits)

// 	if !intent.Until.IsZero() {
// 		var filteredCommits []models.Commit
// 		for _, commit := range convertedCommits {
// 			if commit.CreatedAt.Before(intent.Until) || commit.CreatedAt.Equal(intent.Until) {
// 				filteredCommits = append(filteredCommits, commit)
// 			} else {
// 				break
// 			}
// 		}
// 		convertedCommits = filteredCommits
// 	}

// 	if len(convertedCommits) > 0 {
// 		intent.LastFetched = convertedCommits[0].CreatedAt
// 	}

// 	event := events.NewCommitsDataEvent{
// 		Repository: intent.Repo,
// 		Since:      intent.LastFetched,
// 		Commits:    convertedCommits,
// 	}

// 	data, err := json.Marshal(event)
// 	if err != nil {
// 		return fmt.Errorf("error marshalling commits: %w", err)
// 	}

// 	if err := svc.mc.Publish(ctx); err != nil {
// 		return fmt.Errorf("error publishing commits: %w", err)
// 	}

// 	intentJSON, _ := json.Marshal(intent)
// 	_, err = svc.kv.Put(ctx, "intent:"+intent.Repo, intentJSON)
// 	if err != nil {
// 		return fmt.Errorf("error storing updated intent for %s: %w", intent.Repo, err)
// 	}

// 	return nil
// }

// func convertCommits(githubCommits []*github.RepositoryCommit) []models.Commit {
// 	var commits []models.Commit
// 	for _, commit := range githubCommits {
// 		if commit.Commit != nil && commit.Commit.Author != nil && commit.Commit.Committer != nil {
// 			commits = append(commits, models.Commit{
// 				Hash: commit.GetSHA(),
// 				Author: models.Author{
// 					Name:     commit.Commit.Author.GetName(),
// 					Email:    commit.Commit.Author.GetEmail(),
// 					Username: commit.Author.GetLogin(),
// 					ID:       commit.Author.GetID(),
// 				},
// 				Message:   commit.Commit.GetMessage(),
// 				CreatedAt: commit.Commit.Committer.GetDate().Time,
// 			})
// 		}
// 	}
// 	return commits
// }
