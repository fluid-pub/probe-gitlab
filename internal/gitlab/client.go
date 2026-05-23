package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fluid/probes/gitlab/internal/config"
	"fluid/probes/gitlab/internal/models"
)

// Client represents a GitLab client
type Client struct {
	baseURL          string
	token            string
	apiVersion       string
	groupID          string
	includeSubgroups bool
	httpClient       *http.Client
}

// NewClient creates a new GitLab client
func NewClient(config *config.GitLabConfig, includeSubgroups bool) *Client {
	return &Client{
		baseURL:          config.URL,
		token:            config.Token,
		apiVersion:       config.APIVersion,
		groupID:          config.GroupID,
		includeSubgroups: includeSubgroups,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetUsers retrieves members of the specified group
func (c *Client) GetUsers() ([]models.User, error) {
	url := fmt.Sprintf("%s/api/%s/groups/%s/members?per_page=100", c.baseURL, c.apiVersion, c.groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("error decoding users response: %w", err)
	}

	// Since we're getting users from a specific group, we can set their group names
	// without making additional API calls
	for i := range users {
		users[i].GroupsNames = []string{c.groupID}
	}

	return users, nil
}

// GetGroups retrieves the specified group
func (c *Client) GetGroups() ([]models.Group, error) {
	url := fmt.Sprintf("%s/api/%s/groups/%s", c.baseURL, c.apiVersion, c.groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Decode as a single group object, not an array
	var group models.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding group response: %w", err)
	}

	return []models.Group{group}, nil
}

// GetProjects retrieves projects from the specified group and optionally from all subgroups
func (c *Client) GetProjects() ([]models.Project, error) {
	var allProjects []models.Project

	// Get projects from the main group
	projects, err := c.getProjectsFromGroup(c.groupID)
	if err != nil {
		return nil, fmt.Errorf("error getting projects from main group: %w", err)
	}
	allProjects = append(allProjects, projects...)

	// If recursive mode is enabled, get projects from all subgroups
	if c.includeSubgroups {
		subgroupProjects, err := c.getProjectsRecursively(c.groupID)
		if err != nil {
			return nil, fmt.Errorf("error getting projects from subgroups: %w", err)
		}
		allProjects = append(allProjects, subgroupProjects...)
	}

	return allProjects, nil
}

// getProjectsFromGroup retrieves projects from a specific group
func (c *Client) getProjectsFromGroup(groupID string) ([]models.Project, error) {
	url := fmt.Sprintf("%s/api/%s/groups/%s/projects?per_page=100", c.baseURL, c.apiVersion, groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var projects []models.Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("error decoding projects response: %w", err)
	}

	return projects, nil
}

// getProjectsRecursively retrieves projects from all subgroups recursively
func (c *Client) getProjectsRecursively(groupID string) ([]models.Project, error) {
	var allProjects []models.Project

	// Get subgroups of the current group
	subgroups, err := c.getSubgroups(groupID)
	if err != nil {
		return nil, fmt.Errorf("error getting subgroups for group %s: %w", groupID, err)
	}

	// For each subgroup, get its projects and recursively get projects from its subgroups
	for _, subgroup := range subgroups {
		// Get projects from this subgroup using its ID
		projects, err := c.getProjectsFromGroup(fmt.Sprintf("%d", subgroup.ID))
		if err != nil {
			return nil, fmt.Errorf("error getting projects from subgroup %s (ID: %d): %w", subgroup.Path, subgroup.ID, err)
		}
		allProjects = append(allProjects, projects...)

		// Recursively get projects from subgroups of this subgroup using its ID
		subProjects, err := c.getProjectsRecursively(fmt.Sprintf("%d", subgroup.ID))
		if err != nil {
			return nil, fmt.Errorf("error getting projects from subgroups of %s (ID: %d): %w", subgroup.Path, subgroup.ID, err)
		}
		allProjects = append(allProjects, subProjects...)
	}

	return allProjects, nil
}

// getSubgroups retrieves subgroups of a specific group
func (c *Client) getSubgroups(groupID string) ([]models.Group, error) {
	url := fmt.Sprintf("%s/api/%s/groups/%s/subgroups?per_page=100", c.baseURL, c.apiVersion, groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var subgroups []models.Group
	if err := json.NewDecoder(resp.Body).Decode(&subgroups); err != nil {
		return nil, fmt.Errorf("error decoding subgroups response: %w", err)
	}

	return subgroups, nil
}
