package config

import "testing"

func TestJoinBlobDisplayURL(t *testing.T) {
	base := "https://gitlab.com/foo/bar/-/blob/main"
	got := JoinBlobDisplayURL(base, "infra/aws/main.tf")
	want := "https://gitlab.com/foo/bar/-/blob/main/infra/aws/main.tf"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
