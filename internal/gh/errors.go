package gh

import (
	"errors"
	"fmt"
)

var (
	// ErrNotInstalled is returned when the gh CLI binary is not found in PATH.
	ErrNotInstalled = errors.New("gh: command not found")

	// ErrNotAuthed is returned when the gh CLI reports the user is not logged in.
	ErrNotAuthed = errors.New("gh: not authenticated, run 'gh auth login'")
)

// ExitError is returned when the gh command exits with a non-zero status.
type ExitError struct {
	Cmd      string
	ExitCode int
	Stderr   string
}

func (e *ExitError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("gh: %s (exit %d)", e.Stderr, e.ExitCode)
	}
	return fmt.Sprintf("gh: exited with code %d", e.ExitCode)
}
