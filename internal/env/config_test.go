package env

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNew_nonExistentPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg := New(path)

	if cfg == nil {
		t.Fatal("New() returned nil")
	}
	if cfg.Path != path {
		t.Errorf("New() Path = %q, want %q", cfg.Path, path)
	}
	if cfg.Env == nil {
		t.Error("New() Env should be initialized, got nil")
	}
	if len(cfg.Env) != 0 {
		t.Errorf("New() Env should be empty for non-existent path, got %d entries", len(cfg.Env))
	}
}

func TestNew_existingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	initial := Config{
		Path: path,
		Env: map[string]Variable{
			"FOO": {Value: "bar"},
		},
	}
	data, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New(path)

	if cfg.Env == nil {
		t.Fatal("New() Env is nil after loading existing file")
	}
	v, ok := cfg.Env["FOO"]
	if !ok {
		t.Error("New() did not load FOO from existing file")
	}
	if v.Value != "bar" {
		t.Errorf("New() loaded FOO.Value = %q, want %q", v.Value, "bar")
	}
}

func TestAdd_withVariables(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cfg := New(path)

	vars := Variables{
		"MY_VAR": {Value: "hello"},
	}
	err := cfg.Add(vars)
	if err != nil {
		t.Fatalf("Add(Variables{}) error: %v", err)
	}
	if _, ok := cfg.Env["MY_VAR"]; !ok {
		t.Error("Add(Variables{}) did not add MY_VAR to Env")
	}
}

func TestAdd_zeroArgs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cfg := New(path)

	err := cfg.Add()
	if err == nil {
		t.Error("Add() with 0 args should return error")
	}
}

func TestAdd_tooManyArgs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cfg := New(path)

	err := cfg.Add("A", Variable{}, "extra")
	if err == nil {
		t.Error("Add() with 3 args should return error")
	}
	if err.Error() != "too many arguments" {
		t.Errorf("Add() with 3 args error = %q, want %q", err.Error(), "too many arguments")
	}
}

func TestAdd_wrongType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cfg := New(path)

	err := cfg.Add(12345)
	if err == nil {
		t.Error("Add() with wrong type should return error")
	}
}

func TestAdd_twoArgs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cfg := New(path)

	err := cfg.Add("MYKEY", Variable{Value: "myval"})
	if err != nil {
		t.Fatalf("Add(string, Variable) error: %v", err)
	}
	v, ok := cfg.Env["MYKEY"]
	if !ok {
		t.Error("Add(string, Variable) did not add MYKEY to Env")
	}
	if v.Value != "myval" {
		t.Errorf("Add(string, Variable) Value = %q, want %q", v.Value, "myval")
	}
}

func TestAdd_priority_existingValue(t *testing.T) {
	// Priority 1: existing value in Env takes precedence over the incoming variable value.
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	cfg := New(path)
	cfg.Env["PRIO_VAR"] = Variable{Value: "existing"}

	// Ensure the env var is absent so priority 2 doesn't fire
	t.Setenv("PRIO_VAR", "")

	err := cfg.Add("PRIO_VAR", Variable{Value: "incoming"})
	if err != nil {
		t.Fatal(err)
	}

	v := cfg.Env["PRIO_VAR"]
	if v.Value != "existing" {
		t.Errorf("add() priority: expected existing value %q, got %q", "existing", v.Value)
	}
}

func TestAdd_priority_osEnvOverrides(t *testing.T) {
	// Priority 2: if os.Getenv differs AND is non-empty, use os.Getenv.
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	cfg := New(path)
	cfg.Env["PRIO_VAR2"] = Variable{Value: "stored"}
	t.Setenv("PRIO_VAR2", "fromenv")

	err := cfg.Add("PRIO_VAR2", Variable{Value: "incoming"})
	if err != nil {
		t.Fatal(err)
	}

	v := cfg.Env["PRIO_VAR2"]
	if v.Value != "fromenv" {
		t.Errorf("add() priority: expected env value %q, got %q", "fromenv", v.Value)
	}
}

func TestAdd_priority_fallbackToOsEnv(t *testing.T) {
	// Priority 3: if value is still empty after existing check, use os.Getenv.
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	cfg := New(path)
	t.Setenv("PRIO_VAR3", "envvalue")

	err := cfg.Add("PRIO_VAR3", Variable{})
	if err != nil {
		t.Fatal(err)
	}

	v := cfg.Env["PRIO_VAR3"]
	if v.Value != "envvalue" {
		t.Errorf("add() priority: expected os.Getenv value %q, got %q", "envvalue", v.Value)
	}
}

func TestAdd_priority_fallbackToDefault(t *testing.T) {
	// Priority 4: if value is still empty, use Default.
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	cfg := New(path)
	t.Setenv("PRIO_VAR4", "")

	err := cfg.Add("PRIO_VAR4", Variable{Default: "defaultval"})
	if err != nil {
		t.Fatal(err)
	}

	v := cfg.Env["PRIO_VAR4"]
	if v.Value != "defaultval" {
		t.Errorf("add() priority: expected default value %q, got %q", "defaultval", v.Value)
	}
}

func TestReadSave_roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cfg := New(path)

	t.Setenv("ROUND_VAR", "")

	err := cfg.Add("ROUND_VAR", Variable{Value: "round_value"})
	if err != nil {
		t.Fatal(err)
	}

	cfg2 := &Config{Path: path, Env: map[string]Variable{}}
	if err := cfg2.read(); err != nil {
		t.Fatalf("read() error: %v", err)
	}

	v, ok := cfg2.Env["ROUND_VAR"]
	if !ok {
		t.Fatal("read() did not restore ROUND_VAR")
	}
	if v.Value != "round_value" {
		t.Errorf("roundtrip: got Value = %q, want %q", v.Value, "round_value")
	}
}

func TestSave_filtersEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	cfg := New(path)

	// Manually add a variable with empty Value and Default (bypassing add() logic)
	cfg.Env["EMPTY_VAR"] = Variable{Value: "", Default: ""}
	cfg.Env["NONEMPTY_VAR"] = Variable{Value: "present"}

	if err := cfg.save(); err != nil {
		t.Fatalf("save() error: %v", err)
	}

	cfg2 := &Config{Path: path, Env: map[string]Variable{}}
	if err := cfg2.read(); err != nil {
		t.Fatalf("read() error: %v", err)
	}

	if _, ok := cfg2.Env["EMPTY_VAR"]; ok {
		t.Error("save() should have filtered out EMPTY_VAR (empty Value and Default)")
	}
	if _, ok := cfg2.Env["NONEMPTY_VAR"]; !ok {
		t.Error("save() should have kept NONEMPTY_VAR")
	}
}

func TestDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "to_delete.json")

	if err := os.WriteFile(path, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Path: path, Env: map[string]Variable{}}
	if err := cfg.delete(); err != nil {
		t.Fatalf("delete() error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("delete() did not remove the file")
	}
}

func TestRefresh(t *testing.T) {
	path := filepath.Join(t.TempDir(), "refresh.json")

	if err := os.WriteFile(path, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Path: path, Env: map[string]Variable{}}
	if err := cfg.Refresh(); err != nil {
		t.Fatalf("Refresh() error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Refresh() did not remove the file")
	}
}
