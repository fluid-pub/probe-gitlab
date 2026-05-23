package config

import (
	"net/url"
	"strings"
)

// DefaultSourceViewTemplate is used when neither global nor repo override is set.
// Placeholders: {path} = GitLab project path (group/sub/project), {ref} = branch.
const DefaultSourceViewTemplate = "https://gitlab.com/{path}/-/blob/{ref}"

// ResolveSourceViewBaseURL expands a template for "View source" links (GitLab blob URL without file path).
// repoOverride wins over globalTemplate when non-empty after trim.
func ResolveSourceViewBaseURL(globalTemplate, repoOverride, repoURL, branch string) string {
	tmpl := strings.TrimSpace(repoOverride)
	if tmpl == "" {
		tmpl = strings.TrimSpace(globalTemplate)
	}
	if tmpl == "" {
		tmpl = DefaultSourceViewTemplate
	}
	path := gitLabProjectPathFromURL(repoURL)
	ref := strings.TrimSpace(branch)
	if ref == "" {
		ref = "main"
	}
	s := strings.ReplaceAll(tmpl, "{path}", path)
	s = strings.ReplaceAll(s, "{ref}", ref)
	return strings.TrimSpace(s)
}

func gitLabProjectPathFromURL(repoURL string) string {
	u, err := url.Parse(strings.TrimSpace(repoURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	p := strings.TrimPrefix(u.Path, "/")
	return strings.TrimSuffix(p, ".git")
}
