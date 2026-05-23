package entities

import (
	"fmt"
	"log"
	"path/filepath"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/models"
	"fluid/probes/gitlab/internal/repositories"
)

type RepositoriesEntity struct {
	cfg state.ConfigProvider
}

func NewRepositoriesEntity(cfg state.ConfigProvider) *RepositoriesEntity {
	return &RepositoriesEntity{cfg: cfg}
}

func (e *RepositoriesEntity) Name() string { return "repositories" }

func (e *RepositoriesEntity) Refresh(client core.Client) (interface{}, error) {
	_ = client
	cfg := getGitLabConfig(e.cfg)
	if cfg == nil || cfg.Data.Repositories == nil {
		return []models.Repository{}, nil
	}
	reposCfg := cfg.Data.Repositories
	if len(reposCfg.Repos) == 0 {
		return []models.Repository{}, nil
	}

	baseDir, err := filepath.Abs(reposCfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("repositories base_dir: %w", err)
	}

	log.Printf("Syncing %d Git repositories under %s", len(reposCfg.Repos), baseDir)
	if err := repositories.Sync(baseDir, reposCfg.Repos, cfg.GitLab.URL, cfg.GitLab.Token); err != nil {
		return nil, fmt.Errorf("sync repositories: %w", err)
	}

	repos := buildRepositoriesState(reposCfg)
	log.Printf("Built repositories entity (%d entries)", len(repos))
	return repos, nil
}

func (e *RepositoriesEntity) Save(stateManager core.StateManager, data interface{}) error {
	repos, ok := data.([]models.Repository)
	if !ok {
		return fmt.Errorf("invalid data type for repositories entity")
	}
	if err := stateManager.SaveEntity(e.Name(), repos); err != nil {
		return fmt.Errorf("save repositories state: %w", err)
	}
	log.Printf("Repositories state saved")
	return nil
}
