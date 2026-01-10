package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	output, _, err := mock.Execute(context.Background(), "test", "arg1", "arg2")

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

	_, _, err := mock.Execute(context.Background(), "fail", "ffmpeg", "-i", "input.mp4")

	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestMockExecutorExecuteNoResponse(t *testing.T) {
	mock := NewMockExecutor()

	_, _, err := mock.Execute(context.Background(), "unknown")

	if err == nil {
		t.Error("expected error for unconfigured command")
	}
}

func TestMockExecutorGetCallCount(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["cmd"] = MockResponse{Output: []byte("ok"), Err: nil}
	mock.Responses["other"] = MockResponse{Output: []byte("ok"), Err: nil}

	_, _, _ = mock.Execute(context.Background(), "cmd")
	_, _, _ = mock.Execute(context.Background(), "cmd")
	_, _, _ = mock.Execute(context.Background(), "other")

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
	_, _, _ = mock.Execute(context.Background(), "cmd")

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

func TestMockCommandExecutor_Execute(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["test"] = MockResponse{
		Output: []byte("output"),
		Usage:  &Usage{UserTime: 1.0, SystemTime: 0.5, MaxMemory: 1024},
	}

	out, usage, err := mock.Execute(context.Background(), "test", "arg1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "output" {
		t.Errorf("expected output 'output', got '%s'", out)
	}
	if usage == nil {
		t.Fatal("expected usage stats, got nil")
	}
	if usage.UserTime != 1.0 || usage.SystemTime != 0.5 || usage.MaxMemory != 1024 {
		t.Errorf("unexpected usage stats: %+v", usage)
	}

	if mock.GetCallCount("test") != 1 {
		t.Errorf("expected 1 call, got %d", mock.GetCallCount("test"))
	}
}

func TestMockCommandExecutor_Execute_Error(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["test"] = MockResponse{
		Err: fmt.Errorf("command failed"),
	}

	_, _, err := mock.Execute(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "command failed" {
		t.Errorf("expected error 'command failed', got '%v'", err)
	}
}

func TestMockCommandExecutor_ExecuteWithProgress(t *testing.T) {
	mock := NewMockExecutor()
	mock.Responses["test"] = MockResponse{
		Output:       []byte("output"),
		ProgressData: []string{"p1", "p2"},
		Usage:        &Usage{UserTime: 2.0},
	}

	progress := make(chan string, 2)
	out, usage, err := mock.ExecuteWithProgress(context.Background(), progress, "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "output" {
		t.Errorf("expected output 'output', got '%s'", out)
	}
	if usage == nil {
		t.Fatal("expected usage stats, got nil")
	}
	if usage.UserTime != 2.0 {
		t.Errorf("unexpected usage stats: %+v", usage)
	}

	// Read from progress channel
	var p []string
	for s := range progress {
		p = append(p, s)
	}

	if len(p) != 2 {
		t.Errorf("expected 2 progress updates, got %d", len(p))
	}

	// Test with progress == nil to cover that branch
	out, usage, err = mock.ExecuteWithProgress(context.Background(), nil, "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(out) != "output" {
		t.Errorf("expected 'output', got '%s'", string(out))
	}
	if usage == nil {
		t.Fatal("expected usage stats, got nil")
	}
	if usage.UserTime != 2.0 {
		t.Errorf("unexpected usage stats: %+v", usage)
	}

	// Test with unconfigured command
	_, _, err = mock.ExecuteWithProgress(context.Background(), nil, "unknown")
	if err == nil {
		t.Error("expected error for unconfigured command")
	}

	// Test with nil Responses
	mock.Responses = nil
	_, _, err = mock.ExecuteWithProgress(context.Background(), nil, "test")
	if err == nil {
		t.Error("expected error for nil Responses")
	}
}

func TestRealCommandExecutor_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	exec := &RealCommandExecutor{}
	out, usage, err := exec.Execute(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(bytes.TrimSpace(out)) != "hello" {
		t.Errorf("expected output 'hello', got '%s'", out)
	}
	if usage == nil {
		t.Fatal("expected usage stats, got nil")
	}
	// Usage stats might be 0 for such a fast command, but shouldn't be nil.
}

func TestRealCommandExecutor_ExecuteWithProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	exec := &RealCommandExecutor{}
	progress := make(chan string, 10)

	// Use a command that prints multiple lines
	out, usage, err := exec.ExecuteWithProgress(context.Background(), progress, "echo", "-e", "line1\nline2")
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}

	if string(bytes.TrimSpace(out)) != "line1\nline2" {
		t.Errorf("expected 'line1\\nline2', got '%s'", string(bytes.TrimSpace(out)))
	}
	if usage == nil {
		t.Fatal("expected usage stats, got nil")
	}

	var totalOutput string
	for line := range progress {
		totalOutput += line
	}
	if totalOutput != "line1\nline2\n" { // The progress channel gets lines with their newlines
		t.Errorf("expected 'line1\\nline2\\n', got '%s'", totalOutput)
	}
}

func TestRealCommandExecutor_Execute_Error(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	exec := &RealCommandExecutor{}
	_, _, err := exec.Execute(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRealCommandExecutor_ExecuteWithProgress_Error(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	exec := &RealCommandExecutor{}
	progress := make(chan string)
	_, _, err := exec.ExecuteWithProgress(context.Background(), progress, "false")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRealCommandExecutor_Execute_NotFound(t *testing.T) {
	exec := &RealCommandExecutor{}
	_, _, err := exec.Execute(context.Background(), "nonexistentcommand")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRealCommandExecutorStartError(t *testing.T) {
	exec := &RealCommandExecutor{}
	_, _, err := exec.ExecuteWithProgress(context.Background(), make(chan string), "")
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestRealCommandExecutorStderrError(t *testing.T) {
	exec := &RealCommandExecutor{}
	// 'ls nonexistent' prints to stderr and returns exit code 2
	_, _, err := exec.Execute(context.Background(), "ls", "nonexistent_file_12345")
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
