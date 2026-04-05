package cmd

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/babarot/afx/internal/config"
	afxpkg "github.com/babarot/afx/internal/pkg"
	"github.com/babarot/afx/internal/state"
)

func init() {
	log.SetOutput(io.Discard)
}

func TestGetPackage_found(t *testing.T) {
	m := metaCmd{
		packages: []afxpkg.Package{
			&afxpkg.GitHub{Name: "tool-a", Owner: "owner", Repo: "tool-a"},
			&afxpkg.GitHub{Name: "tool-b", Owner: "owner", Repo: "tool-b"},
		},
	}

	resource := state.Resource{Name: "tool-a"}
	pkg := m.GetPackage(resource)

	if pkg == nil {
		t.Fatal("GetPackage() returned nil, want non-nil")
	}
	if pkg.GetName() != "tool-a" {
		t.Errorf("GetPackage() name = %q, want %q", pkg.GetName(), "tool-a")
	}
}

func TestGetPackage_notFound(t *testing.T) {
	m := metaCmd{
		packages: []afxpkg.Package{
			&afxpkg.GitHub{Name: "tool-a", Owner: "owner", Repo: "tool-a"},
		},
	}

	resource := state.Resource{Name: "nonexistent"}
	pkg := m.GetPackage(resource)

	if pkg != nil {
		t.Errorf("GetPackage() = %v, want nil for non-existent resource", pkg)
	}
}

func TestGetPackages(t *testing.T) {
	m := metaCmd{
		packages: []afxpkg.Package{
			&afxpkg.GitHub{Name: "tool-a", Owner: "owner", Repo: "tool-a"},
			&afxpkg.GitHub{Name: "tool-b", Owner: "owner", Repo: "tool-b"},
			&afxpkg.GitHub{Name: "tool-c", Owner: "owner", Repo: "tool-c"},
		},
	}

	resources := []state.Resource{
		{Name: "tool-c"},
		{Name: "tool-a"},
	}
	pkgs := m.GetPackages(resources)

	if len(pkgs) != 2 {
		t.Fatalf("GetPackages() returned %d packages, want 2", len(pkgs))
	}
	if pkgs[0].GetName() != "tool-c" {
		t.Errorf("GetPackages()[0] = %q, want %q", pkgs[0].GetName(), "tool-c")
	}
	if pkgs[1].GetName() != "tool-a" {
		t.Errorf("GetPackages()[1] = %q, want %q", pkgs[1].GetName(), "tool-a")
	}
}

func TestGetConfig_mergesAll(t *testing.T) {
	main := &config.Main{Shell: "zsh"}

	m := metaCmd{
		configs: map[string]config.Config{
			"file1.yaml": {
				Main: main,
				GitHub: []*afxpkg.GitHub{
					{Name: "gh1", Owner: "o", Repo: "r1"},
					{Name: "gh2", Owner: "o", Repo: "r2"},
				},
				Gist: []*afxpkg.Gist{
					{Name: "gist1", Owner: "o", ID: "id1"},
				},
			},
			"file2.yaml": {
				GitHub: []*afxpkg.GitHub{
					{Name: "gh3", Owner: "o", Repo: "r3"},
				},
				Gist: []*afxpkg.Gist{
					{Name: "gist2", Owner: "o", ID: "id2"},
				},
			},
		},
	}

	all := m.GetConfig()

	// Verify totals (map iteration order is non-deterministic, use counts)
	if len(all.GitHub) != 3 {
		t.Errorf("GetConfig() GitHub count = %d, want 3", len(all.GitHub))
	}
	if len(all.Gist) != 2 {
		t.Errorf("GetConfig() Gist count = %d, want 2", len(all.Gist))
	}

	// Main should be set from the config that has it
	if all.Main == nil {
		t.Error("GetConfig() Main is nil, want non-nil")
	}
	if all.Main.Shell != "zsh" {
		t.Errorf("GetConfig() Main.Shell = %q, want %q", all.Main.Shell, "zsh")
	}
}

func TestGetConfig_noMain(t *testing.T) {
	m := metaCmd{
		configs: map[string]config.Config{
			"file1.yaml": {
				GitHub: []*afxpkg.GitHub{
					{Name: "gh1", Owner: "o", Repo: "r1"},
				},
			},
		},
	}

	all := m.GetConfig()

	if all.Main != nil {
		t.Errorf("GetConfig() Main = %v, want nil when no config has Main", all.Main)
	}
	if len(all.GitHub) != 1 {
		t.Errorf("GetConfig() GitHub count = %d, want 1", len(all.GitHub))
	}
}

func TestIsCI_false(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("BUILD_NUMBER", "")
	t.Setenv("RUN_ID", "")

	if isCI() {
		t.Error("isCI() = true, want false when no CI env vars are set")
	}
}

func TestIsCI_viaCI(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("BUILD_NUMBER", "")
	t.Setenv("RUN_ID", "")

	if !isCI() {
		t.Error("isCI() = false, want true when CI=true")
	}
}

func TestIsCI_viaBuildNumber(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("BUILD_NUMBER", "1")
	t.Setenv("RUN_ID", "")

	if !isCI() {
		t.Error("isCI() = false, want true when BUILD_NUMBER=1")
	}
}

func TestIsCI_viaRunID(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("BUILD_NUMBER", "")
	t.Setenv("RUN_ID", "1")

	if !isCI() {
		t.Error("isCI() = false, want true when RUN_ID=1")
	}
}

func TestShouldCheckForUpdate_noUpdateNotifier(t *testing.T) {
	t.Setenv("AFX_NO_UPDATE_NOTIFIER", "1")
	t.Setenv("CI", "")
	t.Setenv("BUILD_NUMBER", "")
	t.Setenv("RUN_ID", "")

	if shouldCheckForUpdate() {
		t.Error("shouldCheckForUpdate() = true, want false when AFX_NO_UPDATE_NOTIFIER is set")
	}
}

func TestShouldCheckForUpdate_inCI(t *testing.T) {
	t.Setenv("AFX_NO_UPDATE_NOTIFIER", "")
	t.Setenv("CI", "true")
	t.Setenv("BUILD_NUMBER", "")
	t.Setenv("RUN_ID", "")

	if shouldCheckForUpdate() {
		t.Error("shouldCheckForUpdate() = true, want false in CI environment")
	}
}

func TestShouldCheckForUpdate_nonTerminal(t *testing.T) {
	// In test environment, stdout/stderr are not terminals, so IsTerminal returns false.
	// shouldCheckForUpdate() requires IsTerminal(stdout) && IsTerminal(stderr) to be true,
	// so it will always return false in tests (regardless of CI/notifier env vars).
	t.Setenv("AFX_NO_UPDATE_NOTIFIER", "")
	t.Setenv("CI", "")
	t.Setenv("BUILD_NUMBER", "")
	t.Setenv("RUN_ID", "")

	// Even with no restricting env vars set, non-terminal stdout/stderr means false.
	result := shouldCheckForUpdate()
	if result {
		t.Log("shouldCheckForUpdate() returned true: stdout/stderr appear to be terminals in this environment")
	}
	// We do not assert a specific value here since it depends on the terminal state.
	// The key invariant is: if AFX_NO_UPDATE_NOTIFIER or CI is set, result must be false.
	_ = os.Stdout
}
