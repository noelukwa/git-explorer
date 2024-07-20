package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/noelukwa/git-explorer/internal/pkg/config"
)

func main() {

	var cfg config.ExplorerConfig

	err := envconfig.Process("git_explorer", &cfg)
	if err != nil {
		log.Fatalln(err)
	}

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}
