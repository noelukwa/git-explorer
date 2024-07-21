package events

import (
	"time"

	"github.com/noelukwa/git-explorer/internal/explorer/models"
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
	Kind EventKind          `json:"kind"`
	Info *models.Repository `json:"info"`
}

type NewCommitsDataEvent struct {
	Kind       EventKind       `json:"kind"`
	Repository string          `json:"repository"`
	Since      time.Time       `json:"since"`
	Commits    []models.Commit `json:"commits"`
}
