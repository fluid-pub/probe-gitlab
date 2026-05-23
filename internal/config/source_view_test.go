package config

import "testing"

func TestResolveSourceViewBaseURL(t *testing.T) {
	u := "https://gitlab.com/acme-corp/platform/documentation"
	got := ResolveSourceViewBaseURL("", "", u, "develop")
	want := "https://gitlab.com/acme-corp/platform/documentation/-/blob/develop"
	if got != want {
		t.Fatalf("default template: got %q want %q", got, want)
	}

	custom := "https://gitlab.com/{path}/-/tree/{ref}"
	got2 := ResolveSourceViewBaseURL(custom, "", u, "main")
	want2 := "https://gitlab.com/acme-corp/platform/documentation/-/tree/main"
	if got2 != want2 {
		t.Fatalf("custom template: got %q want %q", got2, want2)
	}

	override := "https://example.com/{path}/raw/{ref}"
	got3 := ResolveSourceViewBaseURL("ignored", override, u, "x")
	want3 := "https://example.com/acme-corp/platform/documentation/raw/x"
	if got3 != want3 {
		t.Fatalf("repo override: got %q want %q", got3, want3)
	}
}
