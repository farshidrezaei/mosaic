package executor

import (
	"bytes"
	"context"
	"os/exec"
)

// CommandExecutor defines an interface for executing external commands.
// This allows for dependency injection and testing without actual FFmpeg/FFprobe.
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)
}

// RealCommandExecutor executes actual system commands.
type RealCommandExecutor struct{}

// Execute runs a real command and returns its stdout output.
func (r *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Return stderr for debugging
		if stderr.Len() > 0 {
			return nil, &CommandError{
				Command: name,
				Args:    args,
				Err:     err,
				Stderr:  stderr.String(),
			}
		}
		return nil, err
	}

	return out.Bytes(), nil
}

// CommandError wraps command execution errors with additional context.
type CommandError struct {
	Command string
	Stderr  string
	Err     error
	Args    []string
}

func (e *CommandError) Error() string {
	if e.Stderr != "" {
		return e.Err.Error() + ": " + e.Stderr
	}
	return e.Err.Error()
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// DefaultExecutor is the default command executor used by the package.
var DefaultExecutor CommandExecutor = &RealCommandExecutor{}
