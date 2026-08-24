package executor

import (
	"context"
	"fmt"
)

// MockCommandExecutor is a mock implementation for testing.
type MockCommandExecutor struct {
	// Responses maps command names to their mock responses
	Responses map[string]MockResponse
	// SequentialResponses maps command names to ordered lists of responses consumed sequentially
	SequentialResponses map[string][]MockResponse
	// CallLog records all commands executed
	CallLog []MockCall
}

// MockResponse defines a mock response for a command.
type MockResponse struct {
	Usage        *Usage
	Err          error
	Output       []byte
	ProgressData []string
}

// MockCall records a command execution.
type MockCall struct {
	Name string
	Args []string
}

// AddSequentialResponse queues a response for a specific command name.
func (m *MockCommandExecutor) AddSequentialResponse(name string, resp MockResponse) {
	if m.SequentialResponses == nil {
		m.SequentialResponses = make(map[string][]MockResponse)
	}
	m.SequentialResponses[name] = append(m.SequentialResponses[name], resp)
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

	var resp MockResponse
	var ok bool

	if m.SequentialResponses != nil && len(m.SequentialResponses[name]) > 0 {
		resp = m.SequentialResponses[name][0]
		m.SequentialResponses[name] = m.SequentialResponses[name][1:]
		ok = true
	} else if m.Responses != nil {
		resp, ok = m.Responses[name]
	}

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

// Reset clears the call log and sequential responses.
func (m *MockCommandExecutor) Reset() {
	m.CallLog = []MockCall{}
	m.SequentialResponses = make(map[string][]MockResponse)
}

// NewMockExecutor creates a new mock executor with no responses configured.
func NewMockExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Responses:           make(map[string]MockResponse),
		SequentialResponses: make(map[string][]MockResponse),
		CallLog:             []MockCall{},
	}
}
