package config

type ExplorerConfig struct {
	DatabaseURL     string `split_words:"true" default:"postgres://explorer:explorer2025@localhost/explorer?sslmode=disable"`
	Port            int    `default:"9800"`
	TestDatabaseURL string `split_words:"true" required:"true"`
}
