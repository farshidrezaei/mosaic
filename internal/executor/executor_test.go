package executor

import (
	"context"
	"errors"
	"testing"
)

func TestNewMockExecutor(t *testing.T) {
	mock := NewMockExecutor()

	if mock.Responses == nil {
		t.Error("expected Responses to be initialized")
	}
	if mock.CallLog == nil {
		t.Error("expected CallLog to be initialized")
	}
}

func TestMockExecutorExecute(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["test"] = MockResponse{
		Output: []byte("output"),
		Err:    nil,
	}

	// Assuming the Execute method signature is now func (m *MockExecutor) Execute(ctx context.Context, name string, args ...string)
	output, err := mock.Execute(context.Background(), "test", "arg1", "arg2")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(output) != "output" {
		t.Errorf("expected 'output', got '%s'", string(output))
	}

	if len(mock.CallLog) != 1 {
		t.Errorf("expected 1 call in log, got %d", len(mock.CallLog))
	}
}

func TestMockExecutorExecuteError(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["fail"] = MockResponse{
		Output: nil,
		Err:    errors.New("failed"),
	}

	_, err := mock.Execute(context.Background(), "fail", "ffmpeg", "-i", "input.mp4")

	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestMockExecutorExecuteNoResponse(t *testing.T) {
	mock := NewMockExecutor()

	_, err := mock.Execute(context.Background(), "unknown")

	if err == nil {
		t.Error("expected error for unconfigured command")
	}
}

func TestMockExecutorGetCallCount(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["cmd"] = MockResponse{Output: []byte("ok"), Err: nil}
	mock.Responses["other"] = MockResponse{Output: []byte("ok"), Err: nil}

	_, _ = mock.Execute(context.Background(), "cmd")
	_, _ = mock.Execute(context.Background(), "cmd")
	_, _ = mock.Execute(context.Background(), "other")

	count := mock.GetCallCount("cmd")
	if count != 2 {
		t.Errorf("expected 2 calls to 'cmd', got %d", count)
	}

	otherCount := mock.GetCallCount("other")
	if otherCount != 1 {
		t.Errorf("expected 1 call to 'other', got %d", otherCount)
	}
}

func TestMockExecutorReset(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["cmd"] = MockResponse{Output: []byte("ok"), Err: nil}
	_, _ = mock.Execute(context.Background(), "cmd")

	if len(mock.CallLog) == 0 {
		t.Error("expected calllog to have entries")
	}

	mock.Reset()

	if len(mock.CallLog) != 0 {
		t.Errorf("expected empty calllog after reset, got %d entries", len(mock.CallLog))
	}
}

func TestCommandErrorError(t *testing.T) {
	err := &CommandError{
		Command: "ffmpeg",
		Args:    []string{"-i", "input.mp4"},
		Err:     errors.New("command failed"),
		Stderr:  "error details",
	}

	errStr := err.Error()
	if errStr != "command failed: error details" {
		t.Errorf("unexpected error string: %s", errStr)
	}
}

func TestCommandErrorErrorNoStderr(t *testing.T) {
	err := &CommandError{
		Command: "ffmpeg",
		Args:    []string{"-i", "input.mp4"},
		Err:     errors.New("command failed"),
		Stderr:  "",
	}

	errStr := err.Error()
	if errStr != "command failed" {
		t.Errorf("unexpected error string: %s", errStr)
	}
}

func TestCommandErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &CommandError{
		Command: "test",
		Err:     originalErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap did not return original error")
	}
}
func TestMockExecutorExecuteWithProgress(t *testing.T) {
	mock := &MockCommandExecutor{Responses: make(map[string]MockResponse)} // CallLog is nil
	mock.Responses["test"] = MockResponse{Output: []byte("ok"), Err: nil}

	progress := make(chan string, 1)
	output, err := mock.ExecuteWithProgress(context.Background(), progress, "test")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(output) != "ok" {
		t.Errorf("expected 'ok', got '%s'", string(output))
	}

	// Mock closes the channel immediately
	_, ok := <-progress
	if ok {
		t.Error("expected progress channel to be closed")
	}

	// Test with progress == nil to cover that branch
	output, err = mock.ExecuteWithProgress(context.Background(), nil, "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(output) != "ok" {
		t.Errorf("expected 'ok', got '%s'", string(output))
	}

	// Test with unconfigured command
	_, err = mock.ExecuteWithProgress(context.Background(), nil, "unknown")
	if err == nil {
		t.Error("expected error for unconfigured command")
	}

	// Test with nil Responses
	mock.Responses = nil
	_, err = mock.ExecuteWithProgress(context.Background(), nil, "test")
	if err == nil {
		t.Error("expected error for nil Responses")
	}
}

func TestRealCommandExecutorExecute(t *testing.T) {
	exec := &RealCommandExecutor{}
	output, err := exec.Execute(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if string(output) != "hello\n" {
		t.Errorf("expected 'hello\\n', got '%s'", string(output))
	}
}

func TestRealCommandExecutorExecuteWithProgress(t *testing.T) {
	exec := &RealCommandExecutor{}
	progress := make(chan string, 10)

	// Use a command that prints multiple lines
	output, err := exec.ExecuteWithProgress(context.Background(), progress, "echo", "-e", "line1\nline2")
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}

	if string(output) != "line1\nline2\n" {
		t.Errorf("expected 'line1\\nline2\\n', got '%s'", string(output))
	}

	var totalOutput string
	for line := range progress {
		totalOutput += line
	}
	if totalOutput != "line1\nline2\n" {
		t.Errorf("expected 'line1\\nline2\\n', got '%s'", totalOutput)
	}
}

func TestRealCommandExecutorStartError(t *testing.T) {
	exec := &RealCommandExecutor{}
	_, err := exec.ExecuteWithProgress(context.Background(), make(chan string), "")
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestRealCommandExecutorError(t *testing.T) {
	exec := &RealCommandExecutor{}
	_, err := exec.Execute(context.Background(), "false")
	if err == nil {
		t.Error("expected error for 'false' command")
	}
}

func TestRealCommandExecutorExecuteWithProgressError(t *testing.T) {
	exec := &RealCommandExecutor{}
	progress := make(chan string, 10)
	_, err := exec.ExecuteWithProgress(context.Background(), progress, "false")
	if err == nil {
		t.Error("expected error for 'false' command")
	}
}
func TestRealCommandExecutorStderrError(t *testing.T) {
	exec := &RealCommandExecutor{}
	// 'ls nonexistent' prints to stderr and returns exit code 2
	_, err := exec.Execute(context.Background(), "ls", "nonexistent_file_12345")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	var cmdErr *CommandError
	if errors.As(err, &cmdErr) {
		if cmdErr.Stderr == "" {
			t.Error("expected stderr in CommandError")
		}
	} else {
		t.Errorf("expected CommandError, got %T", err)
	}
}
