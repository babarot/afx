package shell

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	s := New("echo", "hello")

	if s.Command != "echo" {
		t.Errorf("New() Command = %q, want %q", s.Command, "echo")
	}
	if len(s.Args) != 1 || s.Args[0] != "hello" {
		t.Errorf("New() Args = %v, want [hello]", s.Args)
	}
	if s.Stdin != os.Stdin {
		t.Error("New() Stdin should default to os.Stdin")
	}
	if s.Stdout != os.Stdout {
		t.Error("New() Stdout should default to os.Stdout")
	}
	if s.Stderr != os.Stderr {
		t.Error("New() Stderr should default to os.Stderr")
	}
	if s.Env == nil {
		t.Error("New() Env should not be nil")
	}
	if len(s.Env) != 0 {
		t.Errorf("New() Env should be empty, got %v", s.Env)
	}
}

func TestNew_noArgs(t *testing.T) {
	s := New("ls")
	if s.Command != "ls" {
		t.Errorf("New() Command = %q, want %q", s.Command, "ls")
	}
	if len(s.Args) != 0 {
		t.Errorf("New() Args = %v, want empty", s.Args)
	}
}

func TestShell_Run_echo(t *testing.T) {
	var stdout bytes.Buffer
	s := Shell{
		Stdin:   &bytes.Buffer{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Env:     map[string]string{},
		Command: "echo",
		Args:    []string{"hello"},
	}

	err := s.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	got := stdout.String()
	if got != "hello\n" {
		t.Errorf("Run() stdout = %q, want %q", got, "hello\n")
	}
}

func TestShell_Run_withEnv(t *testing.T) {
	var stdout bytes.Buffer
	s := Shell{
		Stdin:   &bytes.Buffer{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Env:     map[string]string{"TEST_VAR": "myvalue"},
		Command: "sh",
		Args:    []string{"-c", "echo $TEST_VAR"},
	}

	err := s.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}

func TestShell_Run_withDir(t *testing.T) {
	var stdout bytes.Buffer
	dir := t.TempDir()
	s := Shell{
		Stdin:   &bytes.Buffer{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Env:     map[string]string{},
		Command: "pwd",
		Dir:     dir,
	}

	err := s.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}

func TestShell_Run_invalidCommand(t *testing.T) {
	s := Shell{
		Stdin:   &bytes.Buffer{},
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
		Env:     map[string]string{},
		Command: "nonexistent_command_12345",
	}

	err := s.Run(context.Background())
	if err == nil {
		t.Error("Run() expected error for invalid command")
	}
}

func TestShell_Run_contextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	s := Shell{
		Stdin:   &bytes.Buffer{},
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
		Env:     map[string]string{},
		Command: "sleep",
		Args:    []string{"10"},
	}

	err := s.Run(ctx)
	if err == nil {
		t.Error("Run() expected error for canceled context")
	}
}

func TestRunCommand(t *testing.T) {
	err := RunCommand("true")
	if err != nil {
		t.Fatalf("RunCommand() error: %v", err)
	}
}

func TestRunCommand_failure(t *testing.T) {
	err := RunCommand("false")
	if err == nil {
		t.Error("RunCommand() expected error for failing command")
	}
}
