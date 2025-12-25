package executor

import (
	"bytes"
	"context"
	"os/exec"
	"sync"
)

// CommandExecutor defines an interface for executing external commands.
// This allows for dependency injection and testing without actual FFmpeg/FFprobe.
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)
	ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, error)
}

// RealCommandExecutor executes actual system commands.
type RealCommandExecutor struct{}

// Execute runs a real command and returns its stdout output.
func (r *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	return r.ExecuteWithProgress(ctx, nil, name, args...)
}

// ExecuteWithProgress runs a real command and sends progress updates to the provided channel.
func (r *RealCommandExecutor) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if progress == nil {
		cmd.Stdout = &out
	}

	if progress != nil {
		// This is a simplified implementation. In a real scenario, we'd use cmd.StdoutPipe()
		// and read from it in a goroutine.
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		if err := cmd.Start(); err != nil {
			return nil, err
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(progress)
			buf := make([]byte, 1024)
			for {
				n, err := stdoutPipe.Read(buf)
				if n > 0 {
					data := buf[:n]
					out.Write(data)
					progress <- string(data)
				}
				if err != nil {
					break
				}
			}
		}()

		if err := cmd.Wait(); err != nil {
			wg.Wait() // Ensure goroutine finishes even on error
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
		wg.Wait()
	} else {
		if err := cmd.Run(); err != nil {
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
