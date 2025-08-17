package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kacperkwapisz/sortpath/internal/config"
)

func TestSetConfigValue_Validation(t *testing.T) {
	// Create temporary config directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid api-key",
			key:     "api-key",
			value:   "test-key-123",
			wantErr: false,
		},
		{
			name:    "empty api-key",
			key:     "api-key",
			value:   "",
			wantErr: true,
			errMsg:  "API key cannot be empty",
		},
		{
			name:    "valid api-base",
			key:     "api-base",
			value:   "https://api.openai.com/v1",
			wantErr: false,
		},
		{
			name:    "invalid api-base URL",
			key:     "api-base",
			value:   "://invalid-url",
			wantErr: true,
			errMsg:  "invalid API base URL",
		},
		{
			name:    "empty api-base",
			key:     "api-base",
			value:   "",
			wantErr: true,
			errMsg:  "API base URL cannot be empty",
		},
		{
			name:    "valid model",
			key:     "model",
			value:   "gpt-4",
			wantErr: false,
		},
		{
			name:    "empty model",
			key:     "model",
			value:   "",
			wantErr: true,
			errMsg:  "model cannot be empty",
		},
		{
			name:    "valid tree-path",
			key:     "tree-path",
			value:   tmpDir,
			wantErr: false,
		},
		{
			name:    "nonexistent tree-path",
			key:     "tree-path",
			value:   "/nonexistent/path",
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name:    "valid log-level",
			key:     "log-level",
			value:   "debug",
			wantErr: false,
		},
		{
			name:    "invalid log-level",
			key:     "log-level",
			value:   "invalid",
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name:    "unknown key",
			key:     "unknown-key",
			value:   "value",
			wantErr: true,
			errMsg:  "unknown config key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setConfigValue(tt.key, tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("setConfigValue() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("setConfigValue() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("setConfigValue() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestGetConfigValue(t *testing.T) {
	// Create temporary config directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Set up a test config
	configDir := filepath.Join(tmpDir, ".config", "sortpath")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yaml")
	
	testConfig := &config.Config{
		APIKey:   "test-key",
		APIBase:  "https://api.openai.com/v1",
		Model:    "gpt-3.5-turbo",
		TreePath: tmpDir,
		LogLevel: "info",
	}
	
	loader := &config.FileLoader{ConfigPath: configPath}
	if err := loader.Save(testConfig); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		key      string
		expected string
		wantErr  bool
	}{
		{
			name:     "get api-key",
			key:      "api-key",
			expected: "test-key",
			wantErr:  false,
		},
		{
			name:     "get api-base",
			key:      "api-base",
			expected: "https://api.openai.com/v1",
			wantErr:  false,
		},
		{
			name:     "get model",
			key:      "model",
			expected: "gpt-3.5-turbo",
			wantErr:  false,
		},
		{
			name:     "get tree-path",
			key:      "tree-path",
			expected: tmpDir,
			wantErr:  false,
		},
		{
			name:     "get log-level",
			key:      "log-level",
			expected: "info",
			wantErr:  false,
		},
		{
			name:    "get unknown key",
			key:     "unknown-key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := getConfigValue(tt.key)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("getConfigValue() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("getConfigValue() unexpected error = %v", err)
					return
				}
				if value != tt.expected {
					t.Errorf("getConfigValue() = %v, want %v", value, tt.expected)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}