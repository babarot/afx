package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

const validConfigYAML = `github:
  - name: test-pkg
    owner: testowner
    repo: testrepo
    command:
      link:
        - from: "test"
`

func TestRead_validConfig(t *testing.T) {
	dir, err := os.MkdirTemp("", "afx-read-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(validConfigYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Read(path)
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}

	if len(cfg.GitHub) != 1 {
		t.Fatalf("Read() GitHub count = %d, want 1", len(cfg.GitHub))
	}
	pkg := cfg.GitHub[0]
	if pkg.Name != "test-pkg" {
		t.Errorf("Name = %q, want %q", pkg.Name, "test-pkg")
	}
	if pkg.Owner != "testowner" {
		t.Errorf("Owner = %q, want %q", pkg.Owner, "testowner")
	}
	if pkg.Repo != "testrepo" {
		t.Errorf("Repo = %q, want %q", pkg.Repo, "testrepo")
	}
}

func TestRead_invalidYAML(t *testing.T) {
	dir, err := os.MkdirTemp("", "afx-read-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(":\tbad: yaml: [\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = Read(path)
	if err == nil {
		t.Error("Read() expected error for invalid YAML, got nil")
	}
}

func TestRead_nonexistentFile(t *testing.T) {
	_, err := Read("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Read() expected error for nonexistent file, got nil")
	}
}

func TestWalkDir_directory(t *testing.T) {
	dir, err := os.MkdirTemp("", "afx-walkdir-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, name := range []string{"a.yaml", "b.yml", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := WalkDir(dir)
	if err != nil {
		t.Fatalf("WalkDir() error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("WalkDir() found %d files, want 2 (yaml/yml only)", len(files))
	}

	for _, f := range files {
		ext := filepath.Ext(f)
		if ext != ".yaml" && ext != ".yml" {
			t.Errorf("WalkDir() included non-YAML file: %s", f)
		}
	}
}

func TestWalkDir_singleYAMLFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "afx-walkdir-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := WalkDir(path)
	if err != nil {
		t.Fatalf("WalkDir() error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("WalkDir() found %d files, want 1", len(files))
	}
	if len(files) == 1 && files[0] != path {
		t.Errorf("WalkDir() file = %q, want %q", files[0], path)
	}
}

func TestWalkDir_singleNonYAMLFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "afx-walkdir-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "config.txt")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := WalkDir(path)
	if err != nil {
		t.Fatalf("WalkDir() error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("WalkDir() found %d files for non-yaml path, want 0", len(files))
	}
}

func TestWalkDir_nonexistentPath(t *testing.T) {
	_, err := WalkDir("/nonexistent/path")
	if err == nil {
		t.Error("WalkDir() expected error for nonexistent path, got nil")
	}
}

func TestWalkDir_nestedDirectories(t *testing.T) {
	dir, err := os.MkdirTemp("", "afx-walkdir-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join(dir, "top.yaml"),
		filepath.Join(subdir, "nested.yml"),
	} {
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := WalkDir(dir)
	if err != nil {
		t.Fatalf("WalkDir() error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("WalkDir() found %d files, want 2 (including nested)", len(files))
	}
}
