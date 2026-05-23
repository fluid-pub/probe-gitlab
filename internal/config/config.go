package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fluid/probes/core"
	"fluid/probes/gitlab/internal/models"

	"gopkg.in/yaml.v3"
)

// Config is the GitLab probe configuration.
type Config struct {
	Probe             core.ProbeConfig         `yaml:"probe"`
	GitLab            GitLabConfig             `yaml:"gitlab"`
	Data              DataConfig               `yaml:"data"`
	State             core.StateConfig         `yaml:"state"`
	Controlplane      *core.ControlplaneConfig `yaml:"controlplane,omitempty"`
	RAGFieldAllowlist core.RAGFieldSet         `yaml:"-"`
}

// GitLabConfig contains GitLab API settings.
type GitLabConfig struct {
	URL              string `yaml:"url"`
	Token            string `yaml:"token"`
	APIVersion       string `yaml:"api_version"`
	InstallationType string `yaml:"installation_type,omitempty"`
	GroupID          string `yaml:"group_id,omitempty"`
}

// DataConfig extends the core entity list with GitLab-specific collection options.
type DataConfig struct {
	Entities         []core.EntityConfig `yaml:"entities"`
	IncludeSubgroups bool                `yaml:"include_subgroups"`
	Repositories     *RepositoriesConfig `yaml:"repositories,omitempty"`
}

// RepositoriesConfig configures git clone/sync for configured repos.
type RepositoriesConfig struct {
	BaseDir           string       `yaml:"base_dir"`
	SourceViewBaseURL string       `yaml:"source_view_base_url"`
	Repos             []RepoConfig `yaml:"repos"`
}

// RepoConfig describes a repository to clone or update.
type RepoConfig struct {
	URL               string                `yaml:"url"`
	Branch            string                `yaml:"branch"`
	Path              string                `yaml:"path,omitempty"`
	IndexFiles        bool                  `yaml:"index_files"`
	MaxFileBytes      int                   `yaml:"max_file_bytes"`
	MaxFilesPerRepo   int                   `yaml:"max_files_per_repo"`
	SourceViewBaseURL string                `yaml:"source_view_base_url"`
	Rag               *models.RagFacetRules `yaml:"rag,omitempty"`
}

// LoadConfig loads and validates configuration from a YAML file.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read configuration file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse configuration: %w", err)
	}

	if err := config.resolveEnvironmentVariables(); err != nil {
		return nil, fmt.Errorf("resolve environment variables: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	schemaPath := filepath.Join(filepath.Dir(configPath), "schema.yml")
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("schema.yml required (same directory as config) to validate fields.*.rag: %w", err)
	}
	allowed, err := core.ParseRAGFieldSetFromSchemaYAML(schemaData)
	if err != nil {
		return nil, fmt.Errorf("schema.yml (RAG): %w", err)
	}
	if err := core.ValidateRAGEntityFields(config.Data.Entities, allowed); err != nil {
		return nil, err
	}
	config.RAGFieldAllowlist = allowed

	return &config, nil
}

// Validate checks required fields and defaults.
func (c *Config) Validate() error {
	if c.Probe.Name == "" {
		return fmt.Errorf("probe name is missing")
	}
	if c.GitLab.URL == "" {
		return fmt.Errorf("gitlab url is missing")
	}
	if c.GitLab.Token == "" {
		return fmt.Errorf("gitlab token is missing")
	}
	if len(c.Data.Entities) == 0 {
		return fmt.Errorf("at least one entity must be configured")
	}
	if c.State.Dir == "" {
		return fmt.Errorf("state directory is missing")
	}
	if c.State.CleanupInterval <= 0 {
		c.State.CleanupInterval = 60
	}

	if c.HasEntity("repositories") || c.HasEntity("code_files") {
		if c.Data.Repositories == nil {
			return fmt.Errorf("data.repositories is required when repositories or code_files entities are enabled")
		}
		if c.Data.Repositories.BaseDir == "" {
			c.Data.Repositories.BaseDir = "data/repositories"
		}
		for i := range c.Data.Repositories.Repos {
			r := &c.Data.Repositories.Repos[i]
			if r.URL == "" {
				return fmt.Errorf("repositories.repos[%d].url is required", i)
			}
			if r.IndexFiles {
				if r.MaxFileBytes <= 0 {
					r.MaxFileBytes = 512 * 1024
				}
				if r.MaxFilesPerRepo <= 0 {
					r.MaxFilesPerRepo = 2000
				}
			}
		}
	}

	return nil
}

// HasEntity reports whether name is listed in data.entities.
func (c *Config) HasEntity(name string) bool {
	for _, e := range c.Data.Entities {
		if e.Name == name {
			return true
		}
	}
	return false
}

func (c *Config) resolveEnvironmentVariables() error {
	if c.GitLab.Token != "" {
		if resolved, isEnvVar := resolveEnvVar(c.GitLab.Token); isEnvVar {
			if resolved == "" {
				return fmt.Errorf("GITLAB_TOKEN environment variable is not defined")
			}
			c.GitLab.Token = resolved
		}
	}

	if c.Controlplane != nil {
		if c.Controlplane.BaseURL != "" {
			if resolved, isEnvVar := resolveEnvVar(c.Controlplane.BaseURL); isEnvVar {
				if resolved == "" {
					log.Printf("Warning: environment variable not defined for base_url, controlplane disabled")
					c.Controlplane = nil
					return nil
				}
				c.Controlplane.BaseURL = resolved
			}
		}

		if c.Controlplane.Parameters == nil {
			log.Printf("Warning: controlplane parameters missing, controlplane disabled")
			c.Controlplane = nil
			return nil
		}

		if c.Controlplane.Parameters.OrganizationUUID != "" {
			if resolved, isEnvVar := resolveEnvVar(c.Controlplane.Parameters.OrganizationUUID); isEnvVar && resolved != "" {
				c.Controlplane.Parameters.OrganizationUUID = resolved
			}
		}
		if c.Controlplane.Parameters.Token != "" {
			if resolved, isEnvVar := resolveEnvVar(c.Controlplane.Parameters.Token); isEnvVar && resolved != "" {
				c.Controlplane.Parameters.Token = resolved
			}
		}
		if c.Controlplane.Parameters.OrganizationUUID == "" || c.Controlplane.Parameters.Token == "" {
			log.Printf("Warning: controlplane parameters incomplete, controlplane disabled")
			c.Controlplane = nil
		}
	}

	return nil
}

func resolveEnvVar(value string) (string, bool) {
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value, false
	}
	envVar := strings.TrimPrefix(strings.TrimSuffix(value, "}"), "${")
	return os.Getenv(envVar), true
}

func (c *Config) GetProbeName() string                      { return c.Probe.Name }
func (c *Config) GetProbeVersion() string                   { return c.Probe.Version }
func (c *Config) GetStateDir() string                       { return c.State.Dir }
func (c *Config) GetCleanupInterval() int                   { return c.State.CleanupInterval }
func (c *Config) GetEntities() []core.EntityConfig          { return c.Data.Entities }
func (c *Config) GetControlplane() *core.ControlplaneConfig { return c.Controlplane }
