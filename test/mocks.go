package test

import (
	"context"
)

// MockAPIClient provides a mock implementation for API client testing
type MockAPIClient struct {
	QueryFunc func(ctx context.Context, prompt string) (*APIResponse, error)
	CallCount int
	LastPrompt string
}

// APIResponse represents a mock API response
type APIResponse struct {
	Path   string
	Reason string
}

// Query implements the API client interface
func (m *MockAPIClient) Query(ctx context.Context, prompt string) (*APIResponse, error) {
	m.CallCount++
	m.LastPrompt = prompt
	
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, prompt)
	}
	
	return &APIResponse{
		Path:   "src/components/",
		Reason: "Default mock response",
	}, nil
}

// MockFSReader provides a mock implementation for filesystem reading
type MockFSReader struct {
	ReadTreeFunc func(path string) (string, error)
	CallCount    int
	LastPath     string
}

// ReadTree implements the filesystem reader interface
func (m *MockFSReader) ReadTree(path string) (string, error) {
	m.CallCount++
	m.LastPath = path
	
	if m.ReadTreeFunc != nil {
		return m.ReadTreeFunc(path)
	}
	
	return "mock-tree-structure", nil
}

// MockLogger provides a mock implementation for logging
type MockLogger struct {
	InfoCalls  []LogCall
	ErrorCalls []LogCall
	DebugCalls []LogCall
}

// LogCall represents a logged message
type LogCall struct {
	Message string
	Args    []interface{}
}

// Info logs an info message
func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.InfoCalls = append(m.InfoCalls, LogCall{Message: msg, Args: args})
}

// Error logs an error message
func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.ErrorCalls = append(m.ErrorCalls, LogCall{Message: msg, Args: args})
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.DebugCalls = append(m.DebugCalls, LogCall{Message: msg, Args: args})
}

// Reset clears all logged calls
func (m *MockLogger) Reset() {
	m.InfoCalls = nil
	m.ErrorCalls = nil
	m.DebugCalls = nil
}

// MockConfig provides a mock configuration for testing
type MockConfig struct {
	APIKey   string
	APIBase  string
	Model    string
	TreePath string
	LogLevel string
}

// NewMockConfig creates a default mock configuration
func NewMockConfig() *MockConfig {
	return &MockConfig{
		APIKey:   "test-api-key",
		APIBase:  "https://api.openai.com/v1",
		Model:    "gpt-4",
		TreePath: "/test/path",
		LogLevel: "info",
	}
}

// MockInstaller provides a mock implementation for installation operations
type MockInstaller struct {
	InstallFunc     func(opts InstallOptions) error
	IsInstalledFunc func() bool
	CallCount       int
}

// InstallOptions represents installation options
type InstallOptions struct {
	DestPath string
	Force    bool
}

// Install implements the installer interface
func (m *MockInstaller) Install(opts InstallOptions) error {
	m.CallCount++
	
	if m.InstallFunc != nil {
		return m.InstallFunc(opts)
	}
	
	return nil
}

// IsInstalled implements the installer interface
func (m *MockInstaller) IsInstalled() bool {
	if m.IsInstalledFunc != nil {
		return m.IsInstalledFunc()
	}
	
	return true
}