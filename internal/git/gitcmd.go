package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// ErrNotInstalled is returned when the git binary is not found in PATH.
var ErrNotInstalled = errors.New("git: command not found")

// Runner abstracts git command execution for testability.
type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// GitRunner executes git commands as subprocesses.
type GitRunner struct{}

// NewRunner creates a new GitRunner.
func NewRunner() *GitRunner {
	return &GitRunner{}
}

// Run executes a git command with the given arguments.
func (r *GitRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmdStr := "git " + strings.Join(args, " ")
	log.Printf("[DEBUG] exec: %s", cmdStr)

	cmd := exec.CommandContext(ctx, "git", args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()

	if runErr == nil {
		log.Printf("[DEBUG] ok: %s", cmdStr)
		return outBuf.Bytes(), nil
	}

	if errors.Is(runErr, exec.ErrNotFound) {
		return nil, ErrNotInstalled
	}

	stderr := strings.TrimSpace(errBuf.String())
	exitCode := cmd.ProcessState.ExitCode()

	log.Printf("[WARN] command failed: %s (exit %d): %s", cmdStr, exitCode, stderr)

	return nil, &ExitError{
		Cmd:      cmdStr,
		ExitCode: exitCode,
		Stderr:   stderr,
	}
}

// RunInDir executes a git command in the specified directory.
func (r *GitRunner) RunInDir(ctx context.Context, dir string, args ...string) ([]byte, error) {
	fullArgs := append([]string{"-C", dir}, args...)
	return r.Run(ctx, fullArgs...)
}

// ExitError is returned when the git command exits with a non-zero status.
type ExitError struct {
	Cmd      string
	ExitCode int
	Stderr   string
}

func (e *ExitError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("git: %s (exit %d)", e.Stderr, e.ExitCode)
	}
	return fmt.Sprintf("git: exited with code %d", e.ExitCode)
}

// IsAuthError returns true if the error looks like an authentication failure.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "authentication") ||
		strings.Contains(s, "could not read username") ||
		strings.Contains(s, "terminal prompts disabled") ||
		strings.Contains(s, "401") ||
		strings.Contains(s, "403")
}
