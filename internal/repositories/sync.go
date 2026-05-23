package repositories

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"fluid/probes/gitlab/internal/config"
)

// Sync runs clone or pull for each configured repository under baseDir.
// gitlabBaseURL and token are used to build authenticated HTTPS URLs for private repos (e.g. https://oauth2:TOKEN@host/...).
// If token is empty, URLs are used as-is (public repos or SSH).
func Sync(baseDir string, repos []config.RepoConfig, gitlabBaseURL, token string) error {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("create base dir %s: %w", baseDir, err)
	}

	auth := gitHTTPAuth(token)

	for _, r := range repos {
		if err := syncOne(baseDir, r, gitlabBaseURL, token, auth); err != nil {
			log.Printf("repositories: sync %s: %v", r.URL, err)
			// continue with other repos
		}
	}

	return nil
}

func gitHTTPAuth(token string) *githttp.BasicAuth {
	if token == "" {
		return nil
	}
	return &githttp.BasicAuth{Username: "oauth2", Password: token}
}

func syncOne(baseDir string, r config.RepoConfig, gitlabBaseURL, token string, auth *githttp.BasicAuth) error {
	cloneURL := r.URL
	if token != "" && isGitLabURL(r.URL, gitlabBaseURL) {
		cloneURL = injectToken(r.URL, token)
	}

	dirName := r.Path
	if dirName == "" {
		dirName = repoNameFromURL(r.URL)
	}
	dest := filepath.Join(baseDir, dirName)

	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		return pull(dest, r.Branch, auth)
	}
	return clone(cloneURL, dest, r.Branch, auth)
}

func isGitLabURL(repoURL, gitlabBaseURL string) bool {
	base := strings.TrimSuffix(gitlabBaseURL, "/")
	return strings.HasPrefix(strings.TrimSuffix(repoURL, "/"), base) ||
		strings.HasPrefix(strings.TrimSuffix(repoURL, ".git"), base)
}

func injectToken(repoURL, token string) string {
	u, err := url.Parse(repoURL)
	if err != nil {
		return repoURL
	}
	u.User = url.UserPassword("oauth2", token)
	return u.String()
}

func redactURLForLog(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "[url]"
	}
	u.User = nil
	out := u.String()
	if out == "" {
		return "[url]"
	}
	return out
}

func repoNameFromURL(repoURL string) string {
	s := strings.TrimSuffix(repoURL, ".git")
	if idx := strings.LastIndex(s, "/"); idx >= 0 {
		return s[idx+1:]
	}
	return s
}

func clone(cloneURL, dest, branch string, auth *githttp.BasicAuth) error {
	opts := &git.CloneOptions{
		URL:          cloneURL,
		Auth:         auth,
		SingleBranch: true,
		Depth:        1,
	}
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	if _, err := git.PlainClone(dest, false, opts); err != nil {
		return fmt.Errorf("clone repository: %w", err)
	}
	log.Printf("repositories: cloned %s -> %s", redactURLForLog(cloneURL), dest)
	return nil
}

func pull(dest, branch string, auth *githttp.BasicAuth) error {
	repo, err := git.PlainOpen(dest)
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("remote origin: %w", err)
	}
	if err := remote.Fetch(&git.FetchOptions{Auth: auth}); err != nil {
		return fmt.Errorf("fetch origin: %w", err)
	}

	hash, err := resolvePullCommitHash(repo, branch)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}
	if err := worktree.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: hash}); err != nil {
		return fmt.Errorf("reset worktree: %w", err)
	}

	log.Printf("repositories: updated %s", dest)
	return nil
}

func resolvePullCommitHash(repo *git.Repository, branch string) (plumbing.Hash, error) {
	if branch != "" {
		ref, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("resolve origin/%s: %w", branch, err)
		}
		return ref.Hash(), nil
	}
	head, err := repo.Head()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("resolve HEAD: %w", err)
	}
	return head.Hash(), nil
}
