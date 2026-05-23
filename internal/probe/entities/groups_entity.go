package entities

import (
	"fmt"
	"log"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/gitlab"
	"fluid/probes/gitlab/internal/models"
)

type GroupsEntity struct {
	cfg state.ConfigProvider
}

func NewGroupsEntity(cfg state.ConfigProvider) *GroupsEntity {
	return &GroupsEntity{cfg: cfg}
}

func (e *GroupsEntity) Name() string { return "groups" }

func (e *GroupsEntity) Refresh(client core.Client) (interface{}, error) {
	gitlabClient, ok := client.(*gitlab.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type for groups entity, expected *gitlab.Client")
	}
	log.Printf("Fetching GitLab groups...")
	groups, err := gitlabClient.GetGroups()
	if err != nil {
		return nil, fmt.Errorf("fetch groups: %w", err)
	}
	log.Printf("Fetched %d groups", len(groups))
	return groups, nil
}

func (e *GroupsEntity) Save(stateManager core.StateManager, data interface{}) error {
	groups, ok := data.([]models.Group)
	if !ok {
		return fmt.Errorf("invalid data type for groups entity")
	}
	if err := stateManager.SaveEntity(e.Name(), groups); err != nil {
		return fmt.Errorf("save groups state: %w", err)
	}
	log.Printf("Groups state saved")
	return nil
}
