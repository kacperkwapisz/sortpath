package config

import (
	"os"
	"testing"
)

func TestEnvironmentDetector_IsNonInteractive(t *testing.T) {
	detector := &EnvironmentDetector{}
	
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no CI environment",
			envVars:  map[string]string{},
			expected: false, // This will depend on the actual test environment
		},
		{
			name: "GitHub Actions",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			expected: true,
		},
		{
			name: "GitLab CI",
			envVars: map[string]string{
				"GITLAB_CI": "true",
			},
			expected: true,
		},
		{
			name: "Generic CI",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: true,
		},
		{
			name: "Jenkins",
			envVars: map[string]string{
				"JENKINS_URL": "http://jenkins.example.com",
			},
			expected: true,
		},
		{
			name: "dumb terminal",
			envVars: map[string]string{
				"TERM": "dumb",
			},
			expected: true,
		},
		{
			name: "CI false should not be detected",
			envVars: map[string]string{
				"CI": "false",
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			originalEnv := make(map[string]string)
			for k, v := range tt.envVars {
				originalEnv[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			
			// Clean up environment variables
			defer func() {
				for k := range tt.envVars {
					if originalValue, exists := originalEnv[k]; exists {
						os.Setenv(k, originalValue)
					} else {
						os.Unsetenv(k)
					}
				}
			}()
			
			result := detector.IsNonInteractive()
			if result != tt.expected {
				t.Errorf("IsNonInteractive() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnvironmentDetector_GetEnvironmentType(t *testing.T) {
	detector := &EnvironmentDetector{}
	
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "CI environment",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: "ci",
		},
		{
			name: "container environment",
			envVars: map[string]string{
				"KUBERNETES_SERVICE_HOST": "10.0.0.1",
			},
			expected: "container",
		},
		{
			name: "dumb terminal",
			envVars: map[string]string{
				"TERM": "dumb",
			},
			expected: "non-interactive",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			originalEnv := make(map[string]string)
			for k, v := range tt.envVars {
				originalEnv[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			
			// Clean up environment variables
			defer func() {
				for k := range tt.envVars {
					if originalValue, exists := originalEnv[k]; exists {
						os.Setenv(k, originalValue)
					} else {
						os.Unsetenv(k)
					}
				}
			}()
			
			result := detector.GetEnvironmentType()
			if result != tt.expected {
				t.Errorf("GetEnvironmentType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnvironmentDetector_ShouldPromptUser(t *testing.T) {
	detector := &EnvironmentDetector{}
	
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "CI environment should not prompt",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: false,
		},
		{
			name: "dumb terminal should not prompt",
			envVars: map[string]string{
				"TERM": "dumb",
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			originalEnv := make(map[string]string)
			for k, v := range tt.envVars {
				originalEnv[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			
			// Clean up environment variables
			defer func() {
				for k := range tt.envVars {
					if originalValue, exists := originalEnv[k]; exists {
						os.Setenv(k, originalValue)
					} else {
						os.Unsetenv(k)
					}
				}
			}()
			
			result := detector.ShouldPromptUser()
			if result != tt.expected {
				t.Errorf("ShouldPromptUser() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEdgeCaseHandler_HandleMissingConfig(t *testing.T) {
	handler := NewEdgeCaseHandler()
	
	config := handler.HandleMissingConfig()
	
	if config == nil {
		t.Fatal("HandleMissingConfig() returned nil")
	}
	
	// Check that defaults are set
	if config.APIBase != "https://api.openai.com/v1" {
		t.Errorf("Expected default APIBase, got %v", config.APIBase)
	}
	if config.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected default Model, got %v", config.Model)
	}
	if config.TreePath != "." {
		t.Errorf("Expected default TreePath, got %v", config.TreePath)
	}
	if config.LogLevel != "info" {
		t.Errorf("Expected default LogLevel, got %v", config.LogLevel)
	}
	
	// APIKey should be empty (must be provided by user)
	if config.APIKey != "" {
		t.Errorf("Expected empty APIKey, got %v", config.APIKey)
	}
}

func TestEdgeCaseHandler_HandleCorruptedConfig(t *testing.T) {
	handler := NewEdgeCaseHandler()
	
	config, err := handler.HandleCorruptedConfig("/path/to/config.yaml", os.ErrNotExist)
	
	if err != nil {
		t.Errorf("HandleCorruptedConfig() unexpected error = %v", err)
	}
	
	if config == nil {
		t.Fatal("HandleCorruptedConfig() returned nil config")
	}
	
	// Should return defaults
	if config.APIBase != "https://api.openai.com/v1" {
		t.Errorf("Expected default APIBase, got %v", config.APIBase)
	}
}

func TestEdgeCaseHandler_HandlePermissionError(t *testing.T) {
	handler := NewEdgeCaseHandler()
	
	tests := []struct {
		name     string
		envVars  map[string]string
		path     string
		op       string
		wantCode string
	}{
		{
			name: "CI environment",
			envVars: map[string]string{
				"CI": "true",
			},
			path:     "/config/path",
			op:       "write",
			wantCode: "permission_denied_ci",
		},
		{
			name: "container environment",
			envVars: map[string]string{
				"KUBERNETES_SERVICE_HOST": "10.0.0.1",
			},
			path:     "/config/path",
			op:       "write",
			wantCode: "permission_denied_container",
		},
		{
			name:     "regular environment",
			envVars:  map[string]string{},
			path:     "/config/path",
			op:       "write",
			wantCode: "permission_denied",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			originalEnv := make(map[string]string)
			for k, v := range tt.envVars {
				originalEnv[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			
			// Clean up environment variables
			defer func() {
				for k := range tt.envVars {
					if originalValue, exists := originalEnv[k]; exists {
						os.Setenv(k, originalValue)
					} else {
						os.Unsetenv(k)
					}
				}
			}()
			
			err := handler.HandlePermissionError(tt.path, tt.op)
			
			if err == nil {
				t.Fatal("HandlePermissionError() expected error but got nil")
			}
			
			configErr, ok := err.(*ConfigError)
			if !ok {
				t.Fatalf("Expected *ConfigError, got %T", err)
			}
			
			if configErr.Code != tt.wantCode {
				t.Errorf("Expected error code %v, got %v", tt.wantCode, configErr.Code)
			}
		})
	}
}

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		expected string
	}{
		{
			name: "error without cause",
			err: &ConfigError{
				Code:    "test_error",
				Message: "test message",
			},
			expected: "test message",
		},
		{
			name: "error with cause",
			err: &ConfigError{
				Code:    "test_error",
				Message: "test message",
				Cause:   os.ErrNotExist,
			},
			expected: "test message: file does not exist",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	cause := os.ErrNotExist
	err := &ConfigError{
		Code:    "test_error",
		Message: "test message",
		Cause:   cause,
	}
	
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}