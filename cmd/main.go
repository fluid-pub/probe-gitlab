package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/config"
	"fluid/probes/gitlab/internal/probe"
)

func main() {
	configPath := flag.String("config", "config/probe.yml", "path to probe YAML config")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	var probeConfig state.ConfigProvider = cfg
	if cfg.GetControlplane() != nil {
		merged := core.NewMergedConfigProvider(cfg)
		merged.SetRAGFieldAllowlist(cfg.RAGFieldAllowlist)
		probeConfig = merged
	}

	gitlabProbe, err := probe.NewProbe(probeConfig)
	if err != nil {
		log.Fatalf("init probe: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := gitlabProbe.Start(); err != nil {
		log.Fatalf("start probe: %v", err)
	}

	log.Println("GitLab probe running; Ctrl+C to stop")

	<-sigChan
	gitlabProbe.Stop()
	log.Println("GitLab probe stopped")
}
