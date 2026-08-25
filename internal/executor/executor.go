package executor

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"sync"
)

// Usage contains process execution statistics.
type Usage struct {
	UserTime   float64
	SystemTime float64
	MaxMemory  int64
}

// CommandExecutor defines an interface for executing external commands.
// This allows for dependency injection and testing without actual FFmpeg/FFprobe.
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) ([]byte, *Usage, error)
	ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *Usage, error)
}

// RealCommandExecutor executes actual system commands.
type RealCommandExecutor struct{}

// Execute runs a real command and returns its stdout output.
func (r *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, *Usage, error) {
	return r.ExecuteWithProgress(ctx, nil, name, args...)
}

// ExecuteWithProgress runs a real command and sends progress updates to the provided channel.
func (r *RealCommandExecutor) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *Usage, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if progress == nil {
		cmd.Stdout = &out
	}

	if progress != nil {
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return nil, nil, err
		}

		if err := cmd.Start(); err != nil {
			return nil, nil, err
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(progress)
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Text()
				out.WriteString(line + "\n")
				progress <- line
			}
		}()

		if err := cmd.Wait(); err != nil {
			wg.Wait() // Ensure goroutine finishes even on error
			if stderr.Len() > 0 {
				return nil, nil, &CommandError{
					Command: name,
					Args:    args,
					Err:     err,
					Stderr:  stderr.String(),
				}
			}
			return nil, nil, err
		}
		wg.Wait()
	} else {
		if err := cmd.Run(); err != nil {
			if stderr.Len() > 0 {
				return nil, nil, &CommandError{
					Command: name,
					Args:    args,
					Err:     err,
					Stderr:  stderr.String(),
				}
			}
			return nil, nil, err
		}
	}

	usage := &Usage{
		UserTime:   cmd.ProcessState.UserTime().Seconds(),
		SystemTime: cmd.ProcessState.SystemTime().Seconds(),
		MaxMemory:  extractMaxMemory(cmd.ProcessState),
	}

	return out.Bytes(), usage, nil
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
