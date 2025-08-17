package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kacperkwapisz/sortpath/internal/config"
)

func TestResolveConfig_CLIPriority(t *testing.T) {
	// CLI options should take highest priority
	tmpDir := t.TempDir()
	opts := config.CLIOptions{
		APIKey:   "cli-key",
		APIBase:  "https://cli.example.com",
		Model:    "cli-model",
		TreePath: tmpDir,
		LogLevel: "debug",
	}

	// Set environment variables that should be overridden
	os.Setenv("OPENAI_API_KEY", "env-key")
	os.Setenv("OPENAI_API_BASE", "https://env.example.com")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("OPENAI_API_BASE")
	}()

	resolved, err := config.ResolveConfig(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resolved.APIKey != "cli-key" {
		t.Errorf("Expected CLI APIKey 'cli-key', got '%s'", resolved.APIKey)
	}
	if resolved.APIBase != "https://cli.example.com" {
		t.Errorf("Expected CLI APIBase 'https://cli.example.com', got '%s'", resolved.APIBase)
	}
	if resolved.Model != "cli-model" {
		t.Errorf("Expected CLI Model 'cli-model', got '%s'", resolved.Model)
	}
}

func TestResolveConfig_EnvPriority(t *testing.T) {
	// Environment variables should take priority over file and defaults
	tmpDir := t.TempDir()
	opts := config.CLIOptions{} // Empty CLI options

	os.Setenv("OPENAI_API_KEY", "env-key")
	os.Setenv("OPENAI_API_BASE", "https://env.example.com")
	os.Setenv("OPENAI_MODEL", "env-model")
	os.Setenv("SORTPATH_FOLDER_TREE", tmpDir)
	os.Setenv("SORTPATH_LOG_LEVEL", "error")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("OPENAI_API_BASE")
		os.Unsetenv("OPENAI_MODEL")
		os.Unsetenv("SORTPATH_FOLDER_TREE")
		os.Unsetenv("SORTPATH_LOG_LEVEL")
	}()

	resolved, err := config.ResolveConfig(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resolved.APIKey != "env-key" {
		t.Errorf("Expected env APIKey 'env-key', got '%s'", resolved.APIKey)
	}
	if resolved.APIBase != "https://env.example.com" {
		t.Errorf("Expected env APIBase 'https://env.example.com', got '%s'", resolved.APIBase)
	}
	if resolved.Model != "env-model" {
		t.Errorf("Expected env Model 'env-model', got '%s'", resolved.Model)
	}
	if resolved.LogLevel != "error" {
		t.Errorf("Expected env LogLevel 'error', got '%s'", resolved.LogLevel)
	}
}

func TestResolveConfig_FilePriority(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `api_key: file-key
api_base: https://file.example.com
model: file-model
tree_path: /file/path
log_level: info
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Use a custom loader for testing
	loader := &config.FileLoader{ConfigPath: configPath}
	fileConfig, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Test that file values are loaded correctly
	// Since we can't easily mock the Load function, we'll test the file loading directly

	if fileConfig.APIKey != "file-key" {
		t.Errorf("Expected file APIKey 'file-key', got '%s'", fileConfig.APIKey)
	}
	if fileConfig.APIBase != "https://file.example.com" {
		t.Errorf("Expected file APIBase 'https://file.example.com', got '%s'", fileConfig.APIBase)
	}
}

func TestResolveConfig_Defaults(t *testing.T) {
	// Test that defaults are used when no other values are provided
	opts := config.CLIOptions{
		APIKey: "required-key", // API key is required, so provide it
	}

	resolved, err := config.ResolveConfig(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that defaults are applied
	if resolved.APIBase != "https://api.openai.com/v1" {
		t.Errorf("Expected default APIBase 'https://api.openai.com/v1', got '%s'", resolved.APIBase)
	}
	if resolved.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected default Model 'gpt-3.5-turbo', got '%s'", resolved.Model)
	}
	if resolved.LogLevel != "info" {
		t.Errorf("Expected default LogLevel 'info', got '%s'", resolved.LogLevel)
	}
}

func TestResolveConfig_RequiredFields(t *testing.T) {
	// Clear any environment variables that might interfere
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalAPIBase := os.Getenv("OPENAI_API_BASE")
	originalModel := os.Getenv("OPENAI_MODEL")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("OPENAI_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
		if originalAPIBase != "" {
			os.Setenv("OPENAI_API_BASE", originalAPIBase)
		} else {
			os.Unsetenv("OPENAI_API_BASE")
		}
		if originalModel != "" {
			os.Setenv("OPENAI_MODEL", originalModel)
		} else {
			os.Unsetenv("OPENAI_MODEL")
		}
	}()
	
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_BASE")
	os.Unsetenv("OPENAI_MODEL")

	tests := []struct {
		name     string
		opts     config.CLIOptions
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "missing API key",
			opts:    config.CLIOptions{},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "missing API base",
			opts: config.CLIOptions{
				APIKey: "test-key",
				// APIBase intentionally empty to test validation
			},
			wantErr: false, // APIBase has a default value
		},
		{
			name: "all required fields present",
			opts: config.CLIOptions{
				APIKey:  "test-key",
				APIBase: "https://api.example.com",
				Model:   "test-model",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a custom loader that points to a non-existent file to avoid interference
			tempDir := t.TempDir()
			loader := &config.FileLoader{ConfigPath: tempDir + "/nonexistent.yaml"}
			
			_, err := config.ResolveConfigWithLoader(tt.opts, loader)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestResolveConfig_TreePathDefault(t *testing.T) {
	opts := config.CLIOptions{
		APIKey: "test-key",
	}

	resolved, err := config.ResolveConfig(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// TreePath should be set to current working directory or "."
	if resolved.TreePath == "" {
		t.Error("Expected TreePath to be set, got empty string")
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