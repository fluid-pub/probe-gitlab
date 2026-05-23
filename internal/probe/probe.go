package probe

import (
	"fmt"
	"log"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/config"
	"fluid/probes/gitlab/internal/gitlab"
	"fluid/probes/gitlab/internal/manager"
	"fluid/probes/gitlab/internal/probe/entities"
)

type Probe struct {
	*core.Probe
	config *config.Config
	client *gitlab.Client
}

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

// NewProbe accepts state.ConfigProvider (*config.Config or *core.MergedConfigProvider).
func NewProbe(cfg state.ConfigProvider) (*Probe, error) {
	gitlabCfg := getGitLabConfig(cfg)
	if gitlabCfg == nil {
		return nil, fmt.Errorf("NewProbe requires *config.Config or *core.MergedConfigProvider with *config.Config as local")
	}

	client := gitlab.NewClient(&gitlabCfg.GitLab, gitlabCfg.Data.IncludeSubgroups)
	stateManager, err := manager.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	coreProbe := core.NewProbe(cfg, client, stateManager)
	a := &Probe{
		Probe:  coreProbe,
		config: gitlabCfg,
		client: client,
	}

	registerEntities(coreProbe, cfg)

	if merged, ok := cfg.(*core.MergedConfigProvider); ok && stateManager.GetPushManager() != nil {
		stateManager.SetConfigCallbacks(
			merged.GetConfigVersion,
			func(runtimeJSON []byte, configVersion string) {
				runtime, version, err := core.ParseRuntimeConfig(runtimeJSON)
				if err != nil {
					log.Printf("Parse runtime config on reload: %v", err)
					return
				}
				if version != "" {
					configVersion = version
				}
				if err := merged.SetRemote(runtime, configVersion); err != nil {
					log.Printf("Set remote config on reload: %v", err)
					return
				}
				if err := coreProbe.ReloadConfig(); err != nil {
					log.Printf("ReloadConfig failed: %v", err)
				}
			},
		)
	}

	return a, nil
}

func registerEntities(coreProbe *core.Probe, cfg state.ConfigProvider) {
	factories := map[string]func() core.Entity{
		"users":        func() core.Entity { return entities.NewUsersEntity(cfg) },
		"groups":       func() core.Entity { return entities.NewGroupsEntity(cfg) },
		"projects":     func() core.Entity { return entities.NewProjectsEntity(cfg) },
		"repositories": func() core.Entity { return entities.NewRepositoriesEntity(cfg) },
		"code_files":   func() core.Entity { return entities.NewCodeFilesEntity(cfg) },
	}

	for _, ec := range cfg.GetEntities() {
		factory, ok := factories[ec.Name]
		if !ok {
			log.Printf("Warning: unknown entity %q, skipping registration", ec.Name)
			continue
		}
		coreProbe.RegisterEntity(factory())
	}
}

func (a *Probe) Start() error {
	return a.Probe.Start()
}

func (a *Probe) GetStatus() map[string]interface{} {
	status := a.Probe.GetStatus()
	status["gitlab_url"] = a.config.GitLab.URL
	status["include_subgroups"] = a.config.Data.IncludeSubgroups
	return status
}
