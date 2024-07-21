package config

import "time"

type ExplorerConfig struct {
	DatabaseURL       string `split_words:"true" default:"postgres://explorer:explorer2025@localhost/explorer?sslmode=disable"`
	Port              int    `required:"true"`
	TestDatabaseURL   string `split_words:"true" required:"true"`
	MessagingProvider string `split_words:"true" default:"nats"`
	MessagingURL      string `split_words:"true" required:"true"`
}

type ExplorerdConfig struct {
	GithubToken        string        `split_words:"true"`
	MessagingProvider  string        `split_words:"true" default:"nats"`
	MessagingURL       string        `split_words:"true" required:"true"`
	BatchSize          int           `split_words:"true" default:"10"`
	MaxRetries         int           `envconfig:"MAX_RETRIES" default:"3"`
	BackoffInitial     time.Duration `split_words:"true" default:"1s"`
	BackoffMax         time.Duration `split_words:"true" default:"1m"`
	MonitoringInterval time.Duration `split_words:"true" default:"1m"`
}
