package models

// CodeFile is a text file under a cloned repository, pushed for RAG indexing on the control plane.
type CodeFile struct {
	ID        string `json:"id"`
	RepoURL   string `json:"repo_url"`
	FilePath  string `json:"file_path"`
	Content   string `json:"content"`
	Title     string `json:"title,omitempty"`
	SourceURL string `json:"source_url,omitempty"`
}
