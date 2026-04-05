package gh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// Runner abstracts gh command execution for testability.
type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// GHRunner executes gh commands as subprocesses.
type GHRunner struct{}

// NewRunner creates a new GHRunner.
func NewRunner() *GHRunner {
	return &GHRunner{}
}

// Run executes a gh command with the given arguments.
func (r *GHRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmdStr := "gh " + strings.Join(args, " ")
	log.Printf("[DEBUG] exec: %s", cmdStr)

	cmd := exec.CommandContext(ctx, "gh", args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()

	if runErr == nil {
		log.Printf("[DEBUG] ok: %s (%d bytes)", cmdStr, outBuf.Len())
		return outBuf.Bytes(), nil
	}

	if errors.Is(runErr, exec.ErrNotFound) {
		return nil, ErrNotInstalled
	}

	stderr := strings.TrimSpace(errBuf.String())
	exitCode := cmd.ProcessState.ExitCode()

	log.Printf("[WARN] command failed: %s (exit %d): %s", cmdStr, exitCode, stderr)

	if strings.Contains(stderr, "not logged in") ||
		strings.Contains(stderr, "gh auth login") {
		return nil, fmt.Errorf("%w: %s", ErrNotAuthed, stderr)
	}

	return nil, &ExitError{
		Cmd:      cmdStr,
		ExitCode: exitCode,
		Stderr:   stderr,
	}
}
