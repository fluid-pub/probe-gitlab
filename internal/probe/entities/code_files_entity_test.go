package entities

import (
	"testing"

	"fluid/probes/core"
	"fluid/probes/gitlab/internal/config"
	"fluid/probes/gitlab/internal/models"
)

type stubConfig struct {
	entities []core.EntityConfig
}

func (s *stubConfig) GetEntities() []core.EntityConfig { return s.entities }
func (s *stubConfig) GetProbeName() string             { return "test" }
func (s *stubConfig) GetProbeVersion() string          { return "0.0.0" }
func (s *stubConfig) GetStateDir() string              { return "state" }
func (s *stubConfig) GetCleanupInterval() int          { return 60 }

func TestCodeFilesRAGPayload(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Data: config.DataConfig{
			Entities: []core.EntityConfig{
				{
					Name: "code_files",
					Fields: map[string]core.EntityFieldConfig{
						"content": {RAG: true},
					},
				},
			},
		},
	}

	files := []models.CodeFile{{Content: "resource \"x\" {}"}}
	ragEnabled := codeFilesRAGEnabled(cfg)
	for i := range files {
		if ragEnabled {
			files[i].RagForContent = files[i].Content
			files[i].Content = ""
		} else {
			files[i].Content = ""
			files[i].RagForContent = ""
		}
	}

	if files[0].RagForContent != `resource "x" {}` {
		t.Fatalf("expected rag_for_content payload, got %q", files[0].RagForContent)
	}
	if files[0].Content != "" {
		t.Fatal("expected content omitted when RAG is enabled")
	}
}
