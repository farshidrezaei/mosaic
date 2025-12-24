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
	Output []byte
	Err    error
}

// MockCall records a command execution.
type MockCall struct {
	Name string
	Args []string
}

// Execute records the call and returns the mocked response.
func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.CallLog == nil {
		m.CallLog = []MockCall{}
	}

	m.CallLog = append(m.CallLog, MockCall{
		Name: name,
		Args: args,
	})

	if m.Responses == nil {
		return nil, fmt.Errorf("no mock response configured for: %s", name)
	}

	resp, ok := m.Responses[name]
	if !ok {
		return nil, fmt.Errorf("no mock response configured for: %s", name)
	}

	return resp.Output, resp.Err
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
