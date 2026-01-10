package executor

import (
	"context"
	"fmt"
)

// MockCommandExecutor is a mock implementation for testing.
type MockCommandExecutor struct {
	// Responses maps command names to their mock responses
	Responses map[string]MockResponse
	// CallLog records all commands executed
	CallLog []MockCall
}

// MockResponse defines a mock response for a command.
type MockResponse struct {
	Err          error
	Output       []byte
	ProgressData []string
	Usage        *Usage
}

// MockCall records a command execution.
type MockCall struct {
	Name string
	Args []string
}

// Execute records the call and returns the mocked response.
func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, *Usage, error) {
	return m.ExecuteWithProgress(ctx, nil, name, args...)
}

// ExecuteWithProgress records the call and returns the mocked response.
func (m *MockCommandExecutor) ExecuteWithProgress(ctx context.Context, progress chan<- string, name string, args ...string) ([]byte, *Usage, error) {
	if m.CallLog == nil {
		m.CallLog = []MockCall{}
	}

	m.CallLog = append(m.CallLog, MockCall{
		Name: name,
		Args: args,
	})

	if m.Responses == nil {
		return nil, nil, fmt.Errorf("no mock response configured for: %s", name)
	}

	resp, ok := m.Responses[name]
	if !ok {
		return nil, nil, fmt.Errorf("no mock response configured for: %s", name)
	}

	if progress != nil {
		for _, p := range resp.ProgressData {
			progress <- p
		}
		close(progress)
	}

	return resp.Output, resp.Usage, resp.Err
}

// GetCallCount returns the number of times a command was executed.
func (m *MockCommandExecutor) GetCallCount(name string) int {
	count := 0
	for _, call := range m.CallLog {
		if call.Name == name {
			count++
		}
	}
	return count
}

// Reset clears the call log.
func (m *MockCommandExecutor) Reset() {
	m.CallLog = []MockCall{}
}

// NewMockExecutor creates a new mock executor with no responses configured.
func NewMockExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Responses: make(map[string]MockResponse),
		CallLog:   []MockCall{},
	}
}
