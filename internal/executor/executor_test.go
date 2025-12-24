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
