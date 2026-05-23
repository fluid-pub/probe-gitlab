package repositories

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"fluid/probes/gitlab/internal/config"
	"fluid/probes/gitlab/internal/models"
)

var errWalkStop = errors.New("walk: file limit reached")

var textSuffixes = map[string]struct{}{
	".tf": {}, ".tfvars": {}, ".md": {}, ".markdown": {},
	".yml": {}, ".yaml": {}, ".json": {}, ".sh": {}, ".bash": {},
	".ex": {}, ".exs": {}, ".go": {}, ".rs": {},
	".js": {}, ".ts": {}, ".tsx": {}, ".jsx": {},
	".py": {}, ".toml": {}, ".hcl": {},
}

var skipDirNames = map[string]struct{}{
	".git": {}, "node_modules": {}, ".terraform": {}, "vendor": {}, "dist": {}, "build": {},
}

// ScanCodeFiles walks a cloned repo and returns text files for RAG (bounded by size/count).
// blobBaseURL is the resolved GitLab blob prefix (…/-/blob/{ref}); if empty, SourceURL is left empty.
func ScanCodeFiles(baseDir string, repo config.RepoConfig, blobBaseURL string, maxBytes, maxFiles int) ([]models.CodeFile, error) {
	dirName := repo.Path
	if dirName == "" {
		dirName = repoNameFromURL(repo.URL)
	}
	root := filepath.Join(baseDir, dirName)

	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("repo directory %s: %w", root, err)
	}

	repoURL := strings.TrimSpace(repo.URL)
	out := make([]models.CodeFile, 0, 64)
	n := 0

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if _, skip := skipDirNames[name]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if n >= maxFiles {
			return errWalkStop
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !wantFile(rel, maxBytes) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if len(data) > maxBytes {
			return nil
		}
		if !looksText(data) {
			return nil
		}
		id := codeFileID(repoURL, rel)
		sourceURL := ""
		if blobBaseURL != "" {
			sourceURL = config.JoinBlobDisplayURL(blobBaseURL, rel)
		}
		out = append(out, models.CodeFile{
			ID:        id,
			RepoURL:   repoURL,
			FilePath:  rel,
			Content:   string(data),
			Title:     filepath.Base(rel),
			SourceURL: sourceURL,
		})
		n++
		return nil
	})

	if err != nil && !errors.Is(err, errWalkStop) {
		return nil, err
	}
	return out, nil
}

func wantFile(rel string, maxBytes int) bool {
	if maxBytes <= 0 {
		maxBytes = 512 * 1024
	}
	ext := strings.ToLower(filepath.Ext(rel))
	if ext == "" {
		return false
	}
	_, ok := textSuffixes[ext]
	return ok
}

func looksText(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	// Reject obvious binary: NUL in first 8k
	n := len(b)
	if n > 8192 {
		n = 8192
	}
	for i := 0; i < n; i++ {
		if b[i] == 0 {
			return false
		}
	}
	return true
}

func codeFileID(repoURL, filePath string) string {
	h := sha256.Sum256([]byte(repoURL + "\x00" + filePath))
	return hex.EncodeToString(h[:])
}
