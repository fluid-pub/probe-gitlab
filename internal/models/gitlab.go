package models

import "time"

// User represents a GitLab user
type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	State        string     `json:"state"`
	Bio          string     `json:"bio"`
	External     bool       `json:"external"`
	CreatedAt    time.Time  `json:"created_at"`
	LastSignInAt *time.Time `json:"last_sign_in_at"`
	ConfirmedAt  *time.Time `json:"confirmed_at"`
	GroupsNames  []string   `json:"groups_names,omitempty"`
}

// Group represents a GitLab group
type Group struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"created_at"`
	ParentID    *int      `json:"parent_id"`
}

// Namespace represents a GitLab namespace (group or user)
type Namespace struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	FullPath string `json:"full_path"`
	ParentID *int   `json:"parent_id"`
}

// Project represents a GitLab project
type Project struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	Description   string    `json:"description"`
	Visibility    string    `json:"visibility"`
	CreatedAt     time.Time `json:"created_at"`
	Namespace     Namespace `json:"namespace"`
	DefaultBranch string    `json:"default_branch"`
}

// State represents the complete GitLab state
type State struct {
	Probe     string    `yaml:"probe"`
	Timestamp time.Time `yaml:"timestamp"`
	Version   string    `yaml:"version"`
	Data      StateData `yaml:"data"`
}

// StateData contains the state data
type StateData struct {
	Entities StateEntities `yaml:"entities"`
}

// StateEntities contains the state entities
type StateEntities struct {
	Users        []User       `yaml:"users,omitempty"`
	Groups       []Group      `yaml:"groups,omitempty"`
	Projects     []Project    `yaml:"projects,omitempty"`
	Repositories []Repository `yaml:"repositories,omitempty"`
	CodeFiles    []CodeFile   `yaml:"code_files,omitempty"`
}

// Repository represents a cloned repository (entity pushed to control plane)
type Repository struct {
	URL               string         `json:"url"`
	Branch            string         `json:"branch"`
	Path              string         `json:"path"` // local directory name
	Name              string         `json:"name"` // slug (path or derived from URL)
	LastSyncAt        time.Time      `json:"last_sync_at"`
	SourceViewBaseURL string         `json:"source_view_base_url,omitempty"` // resolved blob URL prefix (no trailing slash before file path)
	Rag               *RagFacetRules `json:"rag,omitempty"`
}

// RagFacetRules configures path-based facet tags for RAG (merged with file path on the control plane).
type RagFacetRules struct {
	Rules []RagRule `yaml:"rules" json:"rules"`
}

// RagRule is a single glob match and facet key/values to merge.
type RagRule struct {
	Match  string            `yaml:"match" json:"match"`
	Facets map[string]string `yaml:"facets" json:"facets"`
}
