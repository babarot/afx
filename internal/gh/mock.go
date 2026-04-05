package gh

import (
	"context"
	"strings"
)

// MockRunner is a test double for Runner.
type MockRunner struct {
	Responses map[string][]byte
	Errors    map[string]error
	Called    [][]string
}

func (m *MockRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	m.Called = append(m.Called, args)
	if err, ok := m.Errors[key]; ok {
		return nil, err
	}
	if resp, ok := m.Responses[key]; ok {
		return resp, nil
	}
	return nil, nil
}
