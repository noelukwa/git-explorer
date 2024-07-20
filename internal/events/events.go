package events

import (
	"time"
)

// EventKind represents the type of event.
type EventKind string

const (
	NEW_INTENT  EventKind = "NEW_INTENT"
	NEW_REPO    EventKind = "NEW_REPO"
	NEW_COMMITS EventKind = "NEW_COMMITS"
)

type NewIntentEvent struct {
	Repository string    `json:"repository"`
	Since      time.Time `json:"since"`
}

type NewRepoInfoEvent struct {
	Watchers   int32     `json:"watchers_count"`
	StarGazers int32     `json:"stargazers_count"`
	FullName   string    `json:"full_name"`
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Language   string    `json:"language"`
	Forks      int32     `json:"forks"`
}

type NewCommitsInfoEvent struct {
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
