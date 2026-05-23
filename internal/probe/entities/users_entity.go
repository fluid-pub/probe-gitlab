package entities

import (
	"fmt"
	"log"

	"fluid/probes/core"
	"fluid/probes/core/state"
	"fluid/probes/gitlab/internal/gitlab"
	"fluid/probes/gitlab/internal/models"
)

type UsersEntity struct {
	cfg state.ConfigProvider
}

func NewUsersEntity(cfg state.ConfigProvider) *UsersEntity {
	return &UsersEntity{cfg: cfg}
}

func (e *UsersEntity) Name() string { return "users" }

func (e *UsersEntity) Refresh(client core.Client) (interface{}, error) {
	gitlabClient, ok := client.(*gitlab.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type for users entity, expected *gitlab.Client")
	}
	log.Printf("Fetching GitLab users...")
	users, err := gitlabClient.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("fetch users: %w", err)
	}
	log.Printf("Fetched %d users", len(users))
	return users, nil
}

func (e *UsersEntity) Save(stateManager core.StateManager, data interface{}) error {
	users, ok := data.([]models.User)
	if !ok {
		return fmt.Errorf("invalid data type for users entity")
	}
	if err := stateManager.SaveEntity(e.Name(), users); err != nil {
		return fmt.Errorf("save users state: %w", err)
	}
	log.Printf("Users state saved")
	return nil
}
