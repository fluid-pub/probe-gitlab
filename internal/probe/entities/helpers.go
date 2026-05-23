package entities

import (
	"strings"
	"time"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/config"
	"fluid/probes/gitlab/internal/models"
)

func getGitLabConfig(cfg state.ConfigProvider) *config.Config {
	if c, ok := cfg.(*config.Config); ok {
		return c
	}
	if m, ok := cfg.(*core.MergedConfigProvider); ok {
		if c, ok := m.Local().(*config.Config); ok {
			return c
		}
	}
	return nil
}

func codeFilesRAGEnabled(cfg state.ConfigProvider) bool {
	if cfg == nil {
		return false
	}
	for _, ec := range cfg.GetEntities() {
		if ec.Name == "code_files" && ec.FieldRAG("content") {
			return true
		}
	}
	return false
}

func buildRepositoriesState(reposCfg *config.RepositoriesConfig) []models.Repository {
	if reposCfg == nil {
		return nil
	}
	now := time.Now()
	out := make([]models.Repository, 0, len(reposCfg.Repos))
	for _, r := range reposCfg.Repos {
		name := r.Path
		if name == "" {
			name = repoNameFromURL(r.URL)
		}
		repo := models.Repository{
			URL:        r.URL,
			Branch:     r.Branch,
			Path:       name,
			Name:       name,
			LastSyncAt: now,
		}
		if r.Rag != nil && len(r.Rag.Rules) > 0 {
			repo.Rag = r.Rag
		}
		base := config.ResolveSourceViewBaseURL(
			reposCfg.SourceViewBaseURL,
			r.SourceViewBaseURL,
			r.URL,
			r.Branch,
		)
		if base != "" {
			repo.SourceViewBaseURL = base
		}
		out = append(out, repo)
	}
	return out
}

func repoNameFromURL(repoURL string) string {
	s := strings.TrimSuffix(repoURL, ".git")
	if idx := strings.LastIndex(s, "/"); idx >= 0 {
		return s[idx+1:]
	}
	return s
}
