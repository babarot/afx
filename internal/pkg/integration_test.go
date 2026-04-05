package pkg

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	log.SetOutput(io.Discard)
}

// TestCommand_GetLink_Install_Unlink tests the full symlink lifecycle:
// create a binary in home dir → GetLink → Install (symlink) → Installed → Unlink
func TestCommand_GetLink_Install_Unlink(t *testing.T) {
	// Setup: create home dir with a binary
	homeDir := t.TempDir()
	binDir := t.TempDir()
	t.Setenv("AFX_COMMAND_PATH", binDir)

	binaryPath := filepath.Join(homeDir, "mytool")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho hello"), 0755); err != nil {
		t.Fatal(err)
	}

	pkg := &Local{
		Name:      "test-local",
		Directory: homeDir,
		Command: &Command{
			Link: []*Link{{From: "mytool"}},
		},
	}

	// Test GetLink
	links, err := pkg.Command.GetLink(pkg)
	if err != nil {
		t.Fatalf("GetLink() error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("GetLink() returned %d links, want 1", len(links))
	}
	if links[0].From != binaryPath {
		t.Errorf("GetLink()[0].From = %q, want %q", links[0].From, binaryPath)
	}
	expectedTo := filepath.Join(binDir, "mytool")
	if links[0].To != expectedTo {
		t.Errorf("GetLink()[0].To = %q, want %q", links[0].To, expectedTo)
	}

	// Test Install (creates symlink)
	if err := pkg.Command.Install(pkg); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Verify symlink was created
	fi, err := os.Lstat(expectedTo)
	if err != nil {
		t.Fatalf("symlink not created at %q: %v", expectedTo, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected symlink at %q, got mode %v", expectedTo, fi.Mode())
	}

	// Verify symlink target
	target, err := os.Readlink(expectedTo)
	if err != nil {
		t.Fatalf("Readlink() error: %v", err)
	}
	if target != binaryPath {
		t.Errorf("symlink target = %q, want %q", target, binaryPath)
	}

	// Test Installed
	if !pkg.Command.Installed(pkg) {
		t.Error("Installed() should be true after Install()")
	}

	// Test Unlink
	if err := pkg.Command.Unlink(pkg); err != nil {
		t.Fatalf("Unlink() error: %v", err)
	}
	if _, err := os.Lstat(expectedTo); !os.IsNotExist(err) {
		t.Error("Unlink() should have removed the symlink")
	}
}

// TestCommand_GetLink_dotFrom tests the special case where From is "."
func TestCommand_GetLink_dotFrom(t *testing.T) {
	homeDir := t.TempDir()
	binDir := t.TempDir()
	t.Setenv("AFX_COMMAND_PATH", binDir)

	pkg := &Local{
		Name:      "test-dot",
		Directory: homeDir,
		Command: &Command{
			Link: []*Link{{From: "."}},
		},
	}

	links, err := pkg.Command.GetLink(pkg)
	if err != nil {
		t.Fatalf("GetLink() error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("GetLink() returned %d links, want 1", len(links))
	}
	// When From is ".", the link From should be the home dir itself
	if links[0].From != homeDir {
		t.Errorf("GetLink()[0].From = %q, want %q (home dir)", links[0].From, homeDir)
	}
}

// TestCommand_GetLink_homeNotExist tests error when home dir doesn't exist
func TestCommand_GetLink_homeNotExist(t *testing.T) {
	pkg := &Local{
		Name:      "test-nodir",
		Directory: "/nonexistent/path",
		Command: &Command{
			Link: []*Link{{From: "tool"}},
		},
	}

	_, err := pkg.Command.GetLink(pkg)
	if err == nil {
		t.Error("GetLink() should error when home dir doesn't exist")
	}
}

// TestCommand_GetLink_noMatches tests error when glob finds no files
func TestCommand_GetLink_noMatches(t *testing.T) {
	homeDir := t.TempDir()

	pkg := &Local{
		Name:      "test-nomatch",
		Directory: homeDir,
		Command: &Command{
			Link: []*Link{{From: "nonexistent_binary"}},
		},
	}

	_, err := pkg.Command.GetLink(pkg)
	if err == nil {
		t.Error("GetLink() should error when no files match")
	}
}

// TestGitHub_Uninstall removes home dir and command links
func TestGitHub_Uninstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	g := GitHub{Name: "test-uninstall", Owner: "owner", Repo: "repo"}
	home := g.GetHome()

	// Create home dir
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a file inside
	if err := os.WriteFile(filepath.Join(home, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := g.Uninstall(t.Context()); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Error("Uninstall() should have removed home dir")
	}
}

// TestGist_Uninstall removes home dir
func TestGist_Uninstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	g := Gist{Name: "test-uninstall", Owner: "owner", ID: "abc123"}
	home := g.GetHome()

	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}

	if err := g.Uninstall(t.Context()); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Error("Uninstall() should have removed home dir")
	}
}

// TestCommand_Install_overwritesExistingSymlink tests that Install replaces an existing symlink
func TestCommand_Install_overwritesExistingSymlink(t *testing.T) {
	homeDir := t.TempDir()
	binDir := t.TempDir()
	t.Setenv("AFX_COMMAND_PATH", binDir)

	// Create old and new binaries
	oldBinary := filepath.Join(homeDir, "old")
	newBinary := filepath.Join(homeDir, "tool")
	if err := os.WriteFile(oldBinary, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newBinary, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}

	linkTo := filepath.Join(binDir, "tool")

	// Create existing symlink pointing to old binary
	if err := os.Symlink(oldBinary, linkTo); err != nil {
		t.Fatal(err)
	}

	pkg := &Local{
		Name:      "test-overwrite",
		Directory: homeDir,
		Command: &Command{
			Link: []*Link{{From: "tool"}},
		},
	}

	// Install should overwrite the existing symlink
	if err := pkg.Command.Install(pkg); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Verify symlink now points to new binary
	target, err := os.Readlink(linkTo)
	if err != nil {
		t.Fatalf("Readlink() error: %v", err)
	}
	if target != newBinary {
		t.Errorf("symlink target = %q, want %q", target, newBinary)
	}
}

// TestCommand_buildRequired_integration verifies build+link workflow
func TestCommand_buildRequired_with_steps(t *testing.T) {
	cmd := Command{
		Build: &Build{
			Steps: []string{"echo building"},
		},
		Link: []*Link{{From: "output"}},
	}
	if !cmd.buildRequired() {
		t.Error("buildRequired() should be true when Build.Steps is non-empty")
	}
}
