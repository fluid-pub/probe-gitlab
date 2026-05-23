package entities

import (
	"fmt"
	"log"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/gitlab"
	"fluid/probes/gitlab/internal/models"
)

type ProjectsEntity struct {
	cfg state.ConfigProvider
}

func NewProjectsEntity(cfg state.ConfigProvider) *ProjectsEntity {
	return &ProjectsEntity{cfg: cfg}
}

func (e *ProjectsEntity) Name() string { return "projects" }

func (e *ProjectsEntity) Refresh(client core.Client) (interface{}, error) {
	gitlabClient, ok := client.(*gitlab.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type for projects entity, expected *gitlab.Client")
	}
	log.Printf("Fetching GitLab projects...")
	projects, err := gitlabClient.GetProjects()
	if err != nil {
		return nil, fmt.Errorf("fetch projects: %w", err)
	}
	log.Printf("Fetched %d projects", len(projects))
	return projects, nil
}

func (e *ProjectsEntity) Save(stateManager core.StateManager, data interface{}) error {
	projects, ok := data.([]models.Project)
	if !ok {
		return fmt.Errorf("invalid data type for projects entity")
	}
	if err := stateManager.SaveEntity(e.Name(), projects); err != nil {
		return fmt.Errorf("save projects state: %w", err)
	}
	log.Printf("Projects state saved")
	return nil
}
