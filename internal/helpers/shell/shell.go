package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

// Shell represents shell command
type Shell struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Env     map[string]string
	Command string
	Args    []string
	Dir     string
}

// New returns Shell instance
func New(command string, args ...string) Shell {
	return Shell{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Env:     map[string]string{},
		Command: command,
		Args:    args,
	}
}

// Run runs shell command based on given command and args
func (s Shell) Run(ctx context.Context) error {
	command := s.Command
	if _, err := exec.LookPath(command); err != nil {
		return err
	}
	for _, arg := range s.Args {
		command += " " + arg
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Stderr = s.Stderr
	cmd.Stdout = s.Stdout
	cmd.Stdin = s.Stdin
	cmd.Dir = s.Dir
	for k, v := range s.Env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	return cmd.Run()
}

// RunCommand runs command with given arguments
func RunCommand(command string, args ...string) error {
	return New(command, args...).Run(context.Background())
}
