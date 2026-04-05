package manager

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func init() {
	log.SetOutput(io.Discard)
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		pkgs    []Package
		wantErr bool
	}{
		"no duplicates": {
			pkgs: []Package{
				&GitHub{Name: "pkg1", Owner: "o", Repo: "r1"},
				&GitHub{Name: "pkg2", Owner: "o", Repo: "r2"},
			},
			wantErr: false,
		},
		"with duplicates": {
			pkgs: []Package{
				&GitHub{Name: "pkg1", Owner: "o", Repo: "r1"},
				&GitHub{Name: "pkg1", Owner: "o", Repo: "r2"},
			},
			wantErr: true,
		},
		"empty": {
			pkgs:    []Package{},
			wantErr: false,
		},
		"mixed types no duplicate": {
			pkgs: []Package{
				&GitHub{Name: "pkg1", Owner: "o", Repo: "r"},
				&Local{Name: "pkg2", Directory: "/tmp"},
			},
			wantErr: false,
		},
		"mixed types with duplicate": {
			pkgs: []Package{
				&GitHub{Name: "same", Owner: "o", Repo: "r"},
				&Local{Name: "same", Directory: "/tmp"},
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
		GitHub: []*GitHub{
			{Name: "foo", Owner: "o", Repo: "foo"},
			{Name: "bar", Owner: "o", Repo: "bar"},
		},
		Local: []*Local{
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
		GitHub: []*GitHub{
			{Name: "my-tool", Owner: "o", Repo: "r1"},
			{Name: "other-lib", Owner: "o", Repo: "r2"},
		},
		Local: []*Local{
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
		GitHub: []*GitHub{
			{Name: "gh1", Owner: "o", Repo: "r1"},
			{Name: "gh2", Owner: "o", Repo: "r2"},
		},
		Gist: []*Gist{
			{Name: "gist1", Owner: "o", ID: "abc"},
		},
		Local: []*Local{
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
		pkgs    []Package
		wantErr bool
		wantLen int
	}{
		"no dependencies": {
			pkgs: []Package{
				&GitHub{Name: "a", Owner: "o", Repo: "a"},
				&GitHub{Name: "b", Owner: "o", Repo: "b"},
			},
			wantLen: 2,
		},
		"with dependency": {
			pkgs: []Package{
				&GitHub{Name: "a", Owner: "o", Repo: "a", DependsOn: []string{"b"}},
				&GitHub{Name: "b", Owner: "o", Repo: "b"},
			},
			wantLen: 2,
		},
		"invalid dependency": {
			pkgs: []Package{
				&GitHub{Name: "a", Owner: "o", Repo: "a", DependsOn: []string{"nonexistent"}},
			},
			wantErr: true,
		},
		"circular dependency": {
			pkgs: []Package{
				&GitHub{Name: "a", Owner: "o", Repo: "a", DependsOn: []string{"b"}},
				&GitHub{Name: "b", Owner: "o", Repo: "b", DependsOn: []string{"a"}},
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
	pkgs := []Package{
		&GitHub{Name: "b", Owner: "o", Repo: "b", DependsOn: []string{"a"}},
		&GitHub{Name: "a", Owner: "o", Repo: "a"},
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
		GitHub: []*GitHub{{Name: "x", Owner: "o", Repo: "r"}},
		Gist:   []*Gist{{Name: "x", Owner: "o", ID: "id"}},
		Local:  []*Local{{Name: "x", Directory: "/tmp"}},
		HTTP:   []*HTTP{{Name: "x", URL: "https://example.com"}},
	}

	got := cfg.Get("x")
	total := len(got.GitHub) + len(got.Gist) + len(got.Local) + len(got.HTTP)
	if total != 4 {
		t.Errorf("Get('x') returned %d, want 4 (one per type)", total)
	}
}

func TestValidate_errorMessage(t *testing.T) {
	pkgs := []Package{
		&GitHub{Name: "dup", Owner: "o", Repo: "r1"},
		&GitHub{Name: "dup", Owner: "o", Repo: "r2"},
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
		pkgs []Package
		want bool
	}{
		"no release": {
			pkgs: []Package{&GitHub{Name: "a", Owner: "o", Repo: "r"}},
			want: false,
		},
		"with release": {
			pkgs: []Package{&GitHub{Name: "a", Owner: "o", Repo: "r", Release: &GitHubRelease{}}},
			want: true,
		},
		"non-github": {
			pkgs: []Package{&Local{Name: "a", Directory: "/tmp"}},
			want: false,
		},
		"empty": {
			pkgs: []Package{},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := HasGitHubReleaseBlock(tt.pkgs)
			if got != tt.want {
				t.Errorf("HasGitHubReleaseBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasSudoInCommandBuildSteps(t *testing.T) {
	tests := map[string]struct {
		pkgs []Package
		want bool
	}{
		"no command block": {
			pkgs: []Package{&GitHub{Name: "a", Owner: "o", Repo: "r"}},
			want: false,
		},
		"with sudo": {
			pkgs: []Package{&GitHub{
				Name: "a", Owner: "o", Repo: "r",
				Command: &Command{Build: &Build{Steps: []string{"sudo make install"}}},
			}},
			want: true,
		},
		"without sudo": {
			pkgs: []Package{&GitHub{
				Name: "a", Owner: "o", Repo: "r",
				Command: &Command{Build: &Build{Steps: []string{"make install"}}},
			}},
			want: false,
		},
		"no build": {
			pkgs: []Package{&GitHub{
				Name: "a", Owner: "o", Repo: "r",
				Command: &Command{},
			}},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := HasSudoInCommandBuildSteps(tt.pkgs)
			if got != tt.want {
				t.Errorf("HasSudoInCommandBuildSteps() = %v, want %v", got, tt.want)
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
var _ = &GitHubRelease{}
