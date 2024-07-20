package models

import (
	"net/url"
	"time"
)

// Repository represent the remote git repository
type Repository struct {
	Watchers   int32     `json:"watchers_count"`
	StarGazers int32     `json:"stargazers_count"`
	FullName   string    `json:"full_name"`
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Language   string    `json:"language"`
	Forks      int32     `json:"forks"`
}

// Commit represent the individual commits in the remote repository
type Commit struct {
	Hash      string    `json:"hash"`
	Author    Author    `json:"author"`
	Message   string    `json:"message"`
	Url       *url.URL  `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Repository
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	ID       int64  `json:"id"`
}

type AuthorStats struct {
	Author  Author
	Commits int64
}

type CommitPage struct {
	Commits    []Commit
	TotalCount int64
	Page       int32
	PerPage    int32
}
