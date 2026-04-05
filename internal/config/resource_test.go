package config

import (
	"testing"
)

func TestGetResource(t *testing.T) {
	t.Setenv("HOME", "/test/home")

	tests := map[string]struct {
		pkg      Package
		wantType string
		wantID   string
		wantName string
	}{
		"GitHub basic": {
			pkg:      GitHub{Name: "test", Owner: "owner", Repo: "repo"},
			wantType: "GitHub",
			wantID:   "github.com/owner/repo",
			wantName: "test",
		},
		"GitHub with Release": {
			pkg: GitHub{
				Name:    "test",
				Owner:   "o",
				Repo:    "r",
				Release: &GitHubRelease{Name: "r", Tag: "v1.0"},
			},
			wantType: "GitHub Release",
			wantID:   "github.com/release/o/r",
			wantName: "test",
		},
		"GitHub GH Extension": {
			pkg: GitHub{
				Name:  "test",
				Owner: "o",
				Repo:  "r",
				As: &GitHubAs{
					GHExtension: &GHExtension{Name: "gh-test"},
				},
			},
			wantType: "GitHub (gh extension)",
			wantID:   "github.com/o/r",
			wantName: "test",
		},
		"Gist": {
			pkg:      Gist{Name: "test", Owner: "owner", ID: "abc123"},
			wantType: "Gist",
			wantID:   "gist.github.com/owner/abc123",
			wantName: "test",
		},
		"Local": {
			pkg:      Local{Name: "test", Directory: "/tmp/test"},
			wantType: "Local",
			wantID:   "local//tmp/test",
			wantName: "test",
		},
		"HTTP": {
			pkg:      HTTP{Name: "test", URL: "https://example.com/tool"},
			wantType: "HTTP",
			wantID:   "https://example.com/tool",
			wantName: "test",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := getResource(tt.pkg)

			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Paths == nil {
				t.Error("Paths is nil, want non-nil slice")
			}
			if len(got.Paths) < 1 {
				t.Errorf("Paths has %d elements, want at least 1 (home path)", len(got.Paths))
			}
		})
	}
}

func TestGetResource_GitHubReleaseVersion(t *testing.T) {
	t.Setenv("HOME", "/test/home")

	pkg := GitHub{
		Name:    "test",
		Owner:   "o",
		Repo:    "r",
		Release: &GitHubRelease{Name: "r", Tag: "v1.0"},
	}
	got := getResource(pkg)
	if got.Version != "v1.0" {
		t.Errorf("Version = %q, want %q", got.Version, "v1.0")
	}
}

func TestGetResource_GitHubBasicNoVersion(t *testing.T) {
	t.Setenv("HOME", "/test/home")

	pkg := GitHub{Name: "test", Owner: "owner", Repo: "repo"}
	got := getResource(pkg)
	if got.Version != "" {
		t.Errorf("Version = %q, want empty string for non-release GitHub package", got.Version)
	}
}
