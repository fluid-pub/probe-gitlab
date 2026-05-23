package entities

import (
	"fmt"
	"log"
	"path/filepath"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/config"
	"fluid/probes/gitlab/internal/models"
	"fluid/probes/gitlab/internal/repositories"
)

type CodeFilesEntity struct {
	cfg state.ConfigProvider
}

func NewCodeFilesEntity(cfg state.ConfigProvider) *CodeFilesEntity {
	return &CodeFilesEntity{cfg: cfg}
}

func (e *CodeFilesEntity) Name() string { return "code_files" }

func (e *CodeFilesEntity) Refresh(client core.Client) (interface{}, error) {
	_ = client
	cfg := getGitLabConfig(e.cfg)
	if cfg == nil || cfg.Data.Repositories == nil {
		return []models.CodeFile{}, nil
	}
	reposCfg := cfg.Data.Repositories
	if len(reposCfg.Repos) == 0 {
		return []models.CodeFile{}, nil
	}

	baseDir, err := filepath.Abs(reposCfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("repositories base_dir: %w", err)
	}

	if err := repositories.Sync(baseDir, reposCfg.Repos, cfg.GitLab.URL, cfg.GitLab.Token); err != nil {
		return nil, fmt.Errorf("sync before code_files scan: %w", err)
	}

	all := make([]models.CodeFile, 0)
	ragEnabled := codeFilesRAGEnabled(e.cfg)

	for _, r := range reposCfg.Repos {
		if !r.IndexFiles {
			continue
		}
		blobBase := config.ResolveSourceViewBaseURL(
			reposCfg.SourceViewBaseURL,
			r.SourceViewBaseURL,
			r.URL,
			r.Branch,
		)
		files, err := repositories.ScanCodeFiles(baseDir, r, blobBase, r.MaxFileBytes, r.MaxFilesPerRepo)
		if err != nil {
			log.Printf("code_files: scan %s: %v", r.URL, err)
			continue
		}
		if !ragEnabled {
			for i := range files {
				files[i].Content = ""
			}
		}
		all = append(all, files...)
	}

	log.Printf("Indexed %d code_files", len(all))
	return all, nil
}

func (e *CodeFilesEntity) Save(stateManager core.StateManager, data interface{}) error {
	files, ok := data.([]models.CodeFile)
	if !ok {
		return fmt.Errorf("invalid data type for code_files entity")
	}
	if len(files) == 0 {
		return nil
	}
	if err := stateManager.SaveEntity(e.Name(), files); err != nil {
		return fmt.Errorf("save code_files state: %w", err)
	}
	log.Printf("Code files state saved (%d files)", len(files))
	return nil
}
