package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	afxpkg "github.com/babarot/afx/internal/pkg"
)

func init() {
	log.SetOutput(io.Discard)
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		pkgs    []afxpkg.Package
		wantErr bool
	}{
		"no duplicates": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "pkg1", Owner: "o", Repo: "r1"},
				&afxpkg.GitHub{Name: "pkg2", Owner: "o", Repo: "r2"},
			},
			wantErr: false,
		},
		"with duplicates": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "pkg1", Owner: "o", Repo: "r1"},
				&afxpkg.GitHub{Name: "pkg1", Owner: "o", Repo: "r2"},
			},
			wantErr: true,
		},
		"empty": {
			pkgs:    []afxpkg.Package{},
			wantErr: false,
		},
		"mixed types no duplicate": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "pkg1", Owner: "o", Repo: "r"},
				&afxpkg.Local{Name: "pkg2", Directory: "/tmp"},
			},
			wantErr: false,
		},
		"mixed types with duplicate": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "same", Owner: "o", Repo: "r"},
				&afxpkg.Local{Name: "same", Directory: "/tmp"},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := Validate(tt.pkgs)
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestConfig_Get(t *testing.T) {
	cfg := Config{
		GitHub: []*afxpkg.GitHub{
			{Name: "foo", Owner: "o", Repo: "foo"},
			{Name: "bar", Owner: "o", Repo: "bar"},
		},
		Local: []*afxpkg.Local{
			{Name: "baz", Directory: "/tmp/baz"},
		},
	}

	tests := map[string]struct {
		args      []string
		wantCount int
	}{
		"exact match github": {
			args:      []string{"foo"},
			wantCount: 1,
		},
		"no match": {
			args:      []string{"nonexistent"},
			wantCount: 0,
		},
		"match local": {
			args:      []string{"baz"},
			wantCount: 1,
		},
		"multiple args": {
			args:      []string{"foo", "baz"},
			wantCount: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := cfg.Get(tt.args...)
			total := len(got.GitHub) + len(got.Gist) + len(got.Local) + len(got.HTTP)
			if total != tt.wantCount {
				t.Errorf("Get() returned %d packages, want %d", total, tt.wantCount)
			}
		})
	}
}

func TestConfig_Contains(t *testing.T) {
	cfg := Config{
		GitHub: []*afxpkg.GitHub{
			{Name: "my-tool", Owner: "o", Repo: "r1"},
			{Name: "other-lib", Owner: "o", Repo: "r2"},
		},
		Local: []*afxpkg.Local{
			{Name: "my-local", Directory: "/tmp"},
		},
	}

	tests := map[string]struct {
		args      []string
		wantCount int
	}{
		"substring match": {
			args:      []string{"my"},
			wantCount: 2, // my-tool + my-local
		},
		"no match": {
			args:      []string{"xyz"},
			wantCount: 0,
		},
		"full name": {
			args:      []string{"other-lib"},
			wantCount: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := cfg.Contains(tt.args...)
			total := len(got.GitHub) + len(got.Gist) + len(got.Local) + len(got.HTTP)
			if total != tt.wantCount {
				t.Errorf("Contains() returned %d packages, want %d", total, tt.wantCount)
			}
		})
	}
}

func TestVisitYAML(t *testing.T) {
	dir, err := os.MkdirTemp("", "visitYAML")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create test files
	for _, name := range []string{"a.yaml", "b.yml", "c.txt", "d.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var files []string
	fn := visitYAML(&files)
	if err := filepath.Walk(dir, fn); err != nil {
		t.Fatal(err)
	}

	want := 2 // a.yaml + b.yml
	if len(files) != want {
		t.Errorf("visitYAML() found %d files, want %d", len(files), want)
	}

	for _, f := range files {
		ext := filepath.Ext(f)
		if ext != ".yaml" && ext != ".yml" {
			t.Errorf("visitYAML() included non-YAML file: %s", f)
		}
	}
}

func TestParse(t *testing.T) {
	cfg := Config{
		GitHub: []*afxpkg.GitHub{
			{Name: "gh1", Owner: "o", Repo: "r1"},
			{Name: "gh2", Owner: "o", Repo: "r2"},
		},
		Gist: []*afxpkg.Gist{
			{Name: "gist1", Owner: "o", ID: "abc"},
		},
		Local: []*afxpkg.Local{
			{Name: "local1", Directory: "/tmp"},
		},
	}

	pkgs := parse(cfg)
	if len(pkgs) != 4 {
		t.Errorf("parse() returned %d packages, want 4", len(pkgs))
	}

	// Verify names
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.GetName()
	}
	for _, want := range []string{"gh1", "gh2", "gist1", "local1"} {
		if !slices.Contains(names, want) {
			t.Errorf("parse() missing package %q in %v", want, names)
		}
	}
}

func TestSort(t *testing.T) {
	tests := map[string]struct {
		pkgs    []afxpkg.Package
		wantErr bool
		wantLen int
	}{
		"no dependencies": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "a"},
				&afxpkg.GitHub{Name: "b", Owner: "o", Repo: "b"},
			},
			wantLen: 2,
		},
		"with dependency": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "a", DependsOn: []string{"b"}},
				&afxpkg.GitHub{Name: "b", Owner: "o", Repo: "b"},
			},
			wantLen: 2,
		},
		"invalid dependency": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "a", DependsOn: []string{"nonexistent"}},
			},
			wantErr: true,
		},
		"circular dependency": {
			pkgs: []afxpkg.Package{
				&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "a", DependsOn: []string{"b"}},
				&afxpkg.GitHub{Name: "b", Owner: "o", Repo: "b", DependsOn: []string{"a"}},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Sort(tt.pkgs)
			if tt.wantErr {
				if err == nil {
					t.Error("Sort() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Sort() unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("Sort() returned %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestSort_order(t *testing.T) {
	// b depends on a, so a must come first
	pkgs := []afxpkg.Package{
		&afxpkg.GitHub{Name: "b", Owner: "o", Repo: "b", DependsOn: []string{"a"}},
		&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "a"},
	}
	sorted, err := Sort(pkgs)
	if err != nil {
		t.Fatalf("Sort() error: %v", err)
	}

	pos := make(map[string]int)
	for i, p := range sorted {
		pos[p.GetName()] = i
	}
	if pos["a"] >= pos["b"] {
		t.Errorf("expected a before b, got a=%d b=%d", pos["a"], pos["b"])
	}
}

func TestConfig_Get_allTypes(t *testing.T) {
	cfg := Config{
		GitHub: []*afxpkg.GitHub{{Name: "x", Owner: "o", Repo: "r"}},
		Gist:   []*afxpkg.Gist{{Name: "x", Owner: "o", ID: "id"}},
		Local:  []*afxpkg.Local{{Name: "x", Directory: "/tmp"}},
		HTTP:   []*afxpkg.HTTP{{Name: "x", URL: "https://example.com"}},
	}

	got := cfg.Get("x")
	total := len(got.GitHub) + len(got.Gist) + len(got.Local) + len(got.HTTP)
	if total != 4 {
		t.Errorf("Get('x') returned %d, want 4 (one per type)", total)
	}
}

func TestValidate_errorMessage(t *testing.T) {
	pkgs := []afxpkg.Package{
		&afxpkg.GitHub{Name: "dup", Owner: "o", Repo: "r1"},
		&afxpkg.GitHub{Name: "dup", Owner: "o", Repo: "r2"},
	}
	err := Validate(pkgs)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "dup") {
		t.Errorf("error should contain duplicate name, got: %s", err.Error())
	}
}

func TestHasGitHubReleaseBlock(t *testing.T) {
	tests := map[string]struct {
		pkgs []afxpkg.Package
		want bool
	}{
		"no release": {
			pkgs: []afxpkg.Package{&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "r"}},
			want: false,
		},
		"with release": {
			pkgs: []afxpkg.Package{&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "r", Release: &afxpkg.GitHubRelease{}}},
			want: true,
		},
		"non-github": {
			pkgs: []afxpkg.Package{&afxpkg.Local{Name: "a", Directory: "/tmp"}},
			want: false,
		},
		"empty": {
			pkgs: []afxpkg.Package{},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := afxpkg.HasGitHubReleaseBlock(tt.pkgs)
			if got != tt.want {
				t.Errorf("afxpkg.HasGitHubReleaseBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasSudoInCommandBuildSteps(t *testing.T) {
	tests := map[string]struct {
		pkgs []afxpkg.Package
		want bool
	}{
		"no command block": {
			pkgs: []afxpkg.Package{&afxpkg.GitHub{Name: "a", Owner: "o", Repo: "r"}},
			want: false,
		},
		"with sudo": {
			pkgs: []afxpkg.Package{&afxpkg.GitHub{
				Name: "a", Owner: "o", Repo: "r",
				Command: &afxpkg.Command{Build: &afxpkg.Build{Steps: []string{"sudo make install"}}},
			}},
			want: true,
		},
		"without sudo": {
			pkgs: []afxpkg.Package{&afxpkg.GitHub{
				Name: "a", Owner: "o", Repo: "r",
				Command: &afxpkg.Command{Build: &afxpkg.Build{Steps: []string{"make install"}}},
			}},
			want: false,
		},
		"no build": {
			pkgs: []afxpkg.Package{&afxpkg.GitHub{
				Name: "a", Owner: "o", Repo: "r",
				Command: &afxpkg.Command{},
			}},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := afxpkg.HasSudoInCommandBuildSteps(tt.pkgs)
			if got != tt.want {
				t.Errorf("afxpkg.HasSudoInCommandBuildSteps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateDirIfNotExist(t *testing.T) {
	dir, err := os.MkdirTemp("", "afx-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	t.Run("creates new dir", func(t *testing.T) {
		path := filepath.Join(dir, "new", "nested")
		if err := CreateDirIfNotExist(path); err != nil {
			t.Fatalf("CreateDirIfNotExist() error: %v", err)
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("directory was not created")
		}
	})

	t.Run("existing dir ok", func(t *testing.T) {
		if err := CreateDirIfNotExist(dir); err != nil {
			t.Fatalf("CreateDirIfNotExist() error: %v", err)
		}
	})
}

// Verify GitHubRelease type exists (needed for HasGitHubReleaseBlock)
var _ = &afxpkg.GitHubRelease{}
