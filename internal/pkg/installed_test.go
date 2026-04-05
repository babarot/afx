package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitHub_GetHome(t *testing.T) {
	t.Setenv("HOME", "/test/home")
	g := GitHub{Owner: "owner", Repo: "repo"}
	want := filepath.Join("/test/home", ".afx", "github.com", "owner", "repo")
	got := g.GetHome()
	if got != want {
		t.Errorf("GetHome() = %q, want %q", got, want)
	}
}

func TestGitHub_GetHome_GHExtension(t *testing.T) {
	g := GitHub{
		Owner: "owner",
		Repo:  "gh-test",
		As: &GitHubAs{
			GHExtension: &GHExtension{Name: "gh-test"},
		},
	}
	got := g.GetHome()
	// GH extension home is managed by gh CLI, not under .afx
	if got == "" {
		t.Error("GetHome() for GH extension should not be empty")
	}
}

func TestGist_GetHome(t *testing.T) {
	t.Setenv("HOME", "/test/home")
	g := Gist{Owner: "owner", ID: "abc123"}
	want := filepath.Join("/test/home", ".afx", "gist.github.com", "owner", "abc123")
	got := g.GetHome()
	if got != want {
		t.Errorf("GetHome() = %q, want %q", got, want)
	}
}

func TestLocal_GetHome(t *testing.T) {
	t.Setenv("HOME", "/test/home")
	l := Local{Directory: "~/mydir"}
	got := l.GetHome()
	want := "/test/home/mydir"
	if got != want {
		t.Errorf("GetHome() = %q, want %q", got, want)
	}
}

func TestLocal_GetHome_absolute(t *testing.T) {
	l := Local{Directory: "/absolute/path"}
	got := l.GetHome()
	if got != "/absolute/path" {
		t.Errorf("GetHome() = %q, want '/absolute/path'", got)
	}
}

func TestHTTP_GetHome(t *testing.T) {
	t.Setenv("HOME", "/test/home")
	h := HTTP{URL: "https://example.com/releases/tool.tar.gz"}
	got := h.GetHome()
	if got == "" {
		t.Error("GetHome() should not be empty")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("GetHome() = %q, should be absolute path", got)
	}
}

func TestHTTP_ParseURL(t *testing.T) {
	h := &HTTP{
		Name: "test",
		URL:  "https://example.com/{{ .OS }}/tool",
	}
	h.ParseURL()
	// After parsing, URL should have OS substituted
	if h.URL == "https://example.com/{{ .OS }}/tool" {
		t.Error("ParseURL() did not apply template")
	}
}

func TestGitHub_HasPluginBlock(t *testing.T) {
	tests := map[string]struct {
		github GitHub
		want   bool
	}{
		"nil plugin": {
			github: GitHub{},
			want:   false,
		},
		"with plugin": {
			github: GitHub{Plugin: &Plugin{}},
			want:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.github.HasPluginBlock(); got != tt.want {
				t.Errorf("HasPluginBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitHub_HasCommandBlock(t *testing.T) {
	tests := map[string]struct {
		github GitHub
		want   bool
	}{
		"nil command": {
			github: GitHub{},
			want:   false,
		},
		"with command": {
			github: GitHub{Command: &Command{}},
			want:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.github.HasCommandBlock(); got != tt.want {
				t.Errorf("HasCommandBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitHub_HasReleaseBlock(t *testing.T) {
	tests := map[string]struct {
		github GitHub
		want   bool
	}{
		"nil release": {
			github: GitHub{},
			want:   false,
		},
		"with release": {
			github: GitHub{Release: &GitHubRelease{Tag: "v1.0"}},
			want:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.github.HasReleaseBlock(); got != tt.want {
				t.Errorf("HasReleaseBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitHub_GetReleaseTag(t *testing.T) {
	tests := map[string]struct {
		github GitHub
		want   string
	}{
		"no release": {
			github: GitHub{},
			want:   "latest",
		},
		"with tag": {
			github: GitHub{Release: &GitHubRelease{Tag: "v2.0"}},
			want:   "v2.0",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.github.GetReleaseTag(); got != tt.want {
				t.Errorf("GetReleaseTag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGitHub_Installed_noPluginNoCommand(t *testing.T) {
	// Without plugin or command, Installed checks if home dir exists
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	g := GitHub{Name: "test", Owner: "owner", Repo: "repo"}

	// Home doesn't exist yet
	if g.Installed() {
		t.Error("Installed() should be false when home dir doesn't exist")
	}

	// Create home dir
	home := g.GetHome()
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}

	if !g.Installed() {
		t.Error("Installed() should be true when home dir exists")
	}
}

func TestGist_Installed_noPluginNoCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	g := Gist{Name: "test", Owner: "owner", ID: "abc123"}

	if g.Installed() {
		t.Error("Installed() should be false when home dir doesn't exist")
	}

	home := g.GetHome()
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}

	if !g.Installed() {
		t.Error("Installed() should be true when home dir exists")
	}
}

func TestHTTP_Installed_noPluginNoCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	h := HTTP{Name: "test", URL: "https://example.com/tool"}

	if h.Installed() {
		t.Error("Installed() should be false when home dir doesn't exist")
	}

	home := h.GetHome()
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}

	if !h.Installed() {
		t.Error("Installed() should be true when home dir exists")
	}
}

func TestLocal_Installed(t *testing.T) {
	l := Local{Name: "test", Directory: "/tmp"}
	// Local.Installed() always returns true
	if !l.Installed() {
		t.Error("Installed() should always be true for Local")
	}
}

func TestGitHub_GetPluginBlock(t *testing.T) {
	t.Run("nil plugin returns empty", func(t *testing.T) {
		g := GitHub{}
		p := g.GetPluginBlock()
		if p.Sources != nil {
			t.Error("expected empty Plugin")
		}
	})
	t.Run("with plugin returns it", func(t *testing.T) {
		g := GitHub{Plugin: &Plugin{Sources: []string{"*.zsh"}}}
		p := g.GetPluginBlock()
		if len(p.Sources) != 1 {
			t.Errorf("expected 1 source, got %d", len(p.Sources))
		}
	})
}

func TestGitHub_GetCommandBlock(t *testing.T) {
	t.Run("nil command returns empty", func(t *testing.T) {
		g := GitHub{}
		c := g.GetCommandBlock()
		if c.Link != nil {
			t.Error("expected empty Command")
		}
	})
	t.Run("with command returns it", func(t *testing.T) {
		g := GitHub{Command: &Command{Link: []*Link{{From: "bin"}}}}
		c := g.GetCommandBlock()
		if len(c.Link) != 1 {
			t.Errorf("expected 1 link, got %d", len(c.Link))
		}
	})
}

func TestGitHub_IsGHExtension(t *testing.T) {
	tests := map[string]struct {
		github GitHub
		want   bool
	}{
		"no As": {
			github: GitHub{},
			want:   false,
		},
		"As without GHExtension": {
			github: GitHub{As: &GitHubAs{}},
			want:   false,
		},
		"with GHExtension": {
			github: GitHub{As: &GitHubAs{GHExtension: &GHExtension{Name: "gh-test"}}},
			want:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.github.IsGHExtension(); got != tt.want {
				t.Errorf("IsGHExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}
