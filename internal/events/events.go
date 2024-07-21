package events

import (
	"time"
)

// EventKind represents the type of event.
type EventKind string

const (
	NEW_REPO_INTENT  EventKind = "NEW_INTENT"
	NEW_REPO_DATA    EventKind = "NEW_REPO_DATA"
	NEW_COMMITS_DATA EventKind = "NEW_COMMITS_DATA"
)

type NewRepoIntentEvent struct {
	Repository string    `json:"repository"`
	Since      time.Time `json:"since"`
	Kind       EventKind `json:"kind"`
}

type NewRepoDataEvent struct {
	Kind EventKind `json:"kind"`
	Info Repo      `json:"info"`
}

type Repo struct {
	Watchers   int32     `json:"watchers_count"`
	StarGazers int32     `json:"stargazers_count"`
	FullName   string    `json:"full_name"`
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Language   string    `json:"language"`
	Forks      int32     `json:"forks"`
}

type NewCommitsDataEvent struct {
	Kind       EventKind `json:"kind"`
	Repository string    `json:"repository"`
	Since      time.Time `json:"since"`
	Commits    []Commit  `json:"commits"`
}

type Commit struct {
	Hash      string    `json:"hash"`
	Author    Author    `json:"author"`
	Message   string    `json:"message"`
	URL       *string   `json:"url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	ID       int64  `json:"id"`
}
