package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/wutscho/registry-ping/internal/checker"
	"github.com/wutscho/registry-ping/internal/config"
	"github.com/wutscho/registry-ping/internal/notify"
	"github.com/wutscho/registry-ping/internal/registry"
	"github.com/wutscho/registry-ping/internal/registry/dockerhub"
	"github.com/wutscho/registry-ping/internal/state"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	scraperRegistry := registry.NewScraperRegistry(
		dockerhub.NewDockerHubScraper(httpClient),
	)
	stateStore := state.NewJSONStateStore(cfg.StateFile)
	notifier := notify.NewStdoutNotifier()
	c := checker.NewChecker(scraperRegistry, stateStore, notifier)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := c.Run(ctx, cfg.Images); err != nil {
		log.Fatalf("checker: %v", err)
	}
}
