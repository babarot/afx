package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

// New returns Shell instance
func New(command string, args ...string) Shell {
	return Shell{
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
		env:     map[string]string{},
		command: command,
		args:    args,
	}
}

// Shell represents shell command
type Shell struct {
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	env     map[string]string
	command string
	args    []string
	dir     string
}

// Run runs shell command
func (s Shell) Run(ctx context.Context) error {
	command := s.command
	if _, err := exec.LookPath(command); err != nil {
		return err
	}
	for _, arg := range s.args {
		command += " " + arg
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Stderr = s.stderr
	cmd.Stdout = s.stdout
	cmd.Stdin = s.stdin
	cmd.Dir = s.dir
	for k, v := range s.env {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	return cmd.Run()
}

// RunCommand runs command with given arguments
func RunCommand(command string, args ...string) error {
	return New(command, args...).Run(context.Background())
}
