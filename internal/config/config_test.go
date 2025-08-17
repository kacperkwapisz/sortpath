package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				APIKey:   "test-key",
				APIBase:  "https://api.openai.com/v1",
				Model:    "gpt-3.5-turbo",
				TreePath: "/tmp",
				LogLevel: "info",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: Config{
				APIBase:  "https://api.openai.com/v1",
				Model:    "gpt-3.5-turbo",
				TreePath: "/tmp",
				LogLevel: "info",
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "invalid log level",
			config: Config{
				APIKey:   "test-key",
				APIBase:  "https://api.openai.com/v1",
				Model:    "gpt-3.5-turbo",
				TreePath: "/tmp",
				LogLevel: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestResolveConfig_Priority(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	// Write test config file
	configContent := fmt.Sprintf(`api_key: file-key
api_base: https://file.example.com
model: file-model
tree_path: %s
log_level: debug
`, tmpDir)
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	loader := &FileLoader{ConfigPath: configPath}

	tests := []struct {
		name     string
		opts     CLIOptions
		envVars  map[string]string
		expected Config
	}{
		{
			name: "CLI overrides everything",
			opts: CLIOptions{
				APIKey:   "cli-key",
				APIBase:  "https://cli.example.com",
				Model:    "cli-model",
				TreePath: tmpDir,
				LogLevel: "error",
			},
			envVars: map[string]string{
				"OPENAI_API_KEY":        "env-key",
				"OPENAI_API_BASE":       "https://env.example.com",
				"OPENAI_MODEL":          "env-model",
				"SORTPATH_FOLDER_TREE":  tmpDir,
				"SORTPATH_LOG_LEVEL":    "info",
			},
			expected: Config{
				APIKey:   "cli-key",
				APIBase:  "https://cli.example.com",
				Model:    "cli-model",
				TreePath: tmpDir,
				LogLevel: "error",
			},
		},
		{
			name: "ENV overrides file",
			opts: CLIOptions{}, // No CLI options
			envVars: map[string]string{
				"OPENAI_API_KEY":        "env-key",
				"OPENAI_API_BASE":       "https://env.example.com",
				"OPENAI_MODEL":          "env-model",
				"SORTPATH_FOLDER_TREE":  tmpDir,
				"SORTPATH_LOG_LEVEL":    "info",
			},
			expected: Config{
				APIKey:   "env-key",
				APIBase:  "https://env.example.com",
				Model:    "env-model",
				TreePath: tmpDir,
				LogLevel: "info",
			},
		},
		{
			name:    "File values used when no CLI or ENV",
			opts:    CLIOptions{},
			envVars: map[string]string{},
			expected: Config{
				APIKey:   "file-key",
				APIBase:  "https://file.example.com",
				Model:    "file-model",
				TreePath: tmpDir,
				LogLevel: "debug",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			config, err := ResolveConfigWithLoader(tt.opts, loader)
			if err != nil {
				t.Fatalf("ResolveConfigWithLoader() error = %v", err)
			}

			if config.APIKey != tt.expected.APIKey {
				t.Errorf("APIKey = %v, want %v", config.APIKey, tt.expected.APIKey)
			}
			if config.APIBase != tt.expected.APIBase {
				t.Errorf("APIBase = %v, want %v", config.APIBase, tt.expected.APIBase)
			}
			if config.Model != tt.expected.Model {
				t.Errorf("Model = %v, want %v", config.Model, tt.expected.Model)
			}
			if config.TreePath != tt.expected.TreePath {
				t.Errorf("TreePath = %v, want %v", config.TreePath, tt.expected.TreePath)
			}
			if config.LogLevel != tt.expected.LogLevel {
				t.Errorf("LogLevel = %v, want %v", config.LogLevel, tt.expected.LogLevel)
			}
		})
	}
}

func TestFileLoader_LoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	loader := &FileLoader{ConfigPath: configPath}

	// Test loading non-existent file returns empty config
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.APIKey != "" {
		t.Errorf("Expected empty config, got APIKey = %v", config.APIKey)
	}

	// Test saving and loading
	testConfig := &Config{
		APIKey:   "test-key",
		APIBase:  "https://api.openai.com/v1",
		Model:    "gpt-3.5-turbo",
		TreePath: "/test/path",
		LogLevel: "debug",
	}

	if err := loader.Save(testConfig); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %v", info.Mode().Perm())
	}

	// Load and verify
	loadedConfig, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loadedConfig.APIKey != testConfig.APIKey {
		t.Errorf("APIKey = %v, want %v", loadedConfig.APIKey, testConfig.APIKey)
	}
	if loadedConfig.APIBase != testConfig.APIBase {
		t.Errorf("APIBase = %v, want %v", loadedConfig.APIBase, testConfig.APIBase)
	}
	if loadedConfig.Model != testConfig.Model {
		t.Errorf("Model = %v, want %v", loadedConfig.Model, testConfig.Model)
	}
	if loadedConfig.TreePath != testConfig.TreePath {
		t.Errorf("TreePath = %v, want %v", loadedConfig.TreePath, testConfig.TreePath)
	}
	if loadedConfig.LogLevel != testConfig.LogLevel {
		t.Errorf("LogLevel = %v, want %v", loadedConfig.LogLevel, testConfig.LogLevel)
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

func TestResolveConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (Loader, func())
		opts        CLIOptions
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "corrupted config file",
			setupFunc: func() (Loader, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				// Write invalid YAML
				os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0600)
				loader := &FileLoader{ConfigPath: configPath}
				return loader, func() {}
			},
			opts: CLIOptions{
				APIKey:  "test-key",
				APIBase: "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
			},
			expectError: false, // Should still work with CLI options
		},
		{
			name: "missing config file with defaults",
			setupFunc: func() (Loader, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "nonexistent", "config.yaml")
				loader := &FileLoader{ConfigPath: configPath}
				return loader, func() {}
			},
			opts: CLIOptions{
				APIKey: "test-key",
			},
			expectError: false, // Should use defaults
		},
		{
			name: "working directory fallback for tree path",
			setupFunc: func() (Loader, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				loader := &FileLoader{ConfigPath: configPath}
				return loader, func() {}
			},
			opts: CLIOptions{
				APIKey:  "test-key",
				APIBase: "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
				// TreePath not specified, should use working directory
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader, cleanup := tt.setupFunc()
			defer cleanup()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			config, err := ResolveConfigWithLoader(tt.opts, loader)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error = %v, want to contain %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error = %v", err)
					return
				}
				if config == nil {
					t.Errorf("Expected config but got nil")
				}
			}
		})
	}
}

func TestFileLoader_ErrorHandling(t *testing.T) {
	t.Run("permission denied on save", func(t *testing.T) {
		// Create a directory where we can't write
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		os.Mkdir(readOnlyDir, 0400) // Read-only directory
		defer os.Chmod(readOnlyDir, 0755) // Cleanup

		configPath := filepath.Join(readOnlyDir, "config.yaml")
		loader := &FileLoader{ConfigPath: configPath}

		config := &Config{
			APIKey:   "test-key",
			APIBase:  "https://api.openai.com/v1",
			Model:    "gpt-3.5-turbo",
			TreePath: tmpDir,
			LogLevel: "info",
		}

		err := loader.Save(config)
		if err == nil {
			t.Errorf("Expected error when saving to read-only directory")
		}
	})

	t.Run("load corrupted file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		
		// Write invalid YAML
		os.WriteFile(configPath, []byte("invalid: yaml: [content"), 0600)
		
		loader := &FileLoader{ConfigPath: configPath}
		config, err := loader.Load()
		
		// With the new edge case handler, corrupted files should return defaults, not error
		if err != nil {
			t.Errorf("Unexpected error when loading corrupted file: %v", err)
		}
		if config == nil {
			t.Errorf("Expected default config when loading corrupted file, got nil")
		}
		// Should return defaults
		if config.APIBase != "https://api.openai.com/v1" {
			t.Errorf("Expected default APIBase when loading corrupted file, got: %v", config.APIBase)
		}
	})
}

func TestEnvironmentVariableNames(t *testing.T) {
	// Test that we're using the correct environment variable names
	envVars := map[string]string{
		"OPENAI_API_KEY":        "env-key",
		"OPENAI_API_BASE":       "https://env.example.com",
		"OPENAI_MODEL":          "env-model",
		"SORTPATH_FOLDER_TREE":  ".",
		"SORTPATH_LOG_LEVEL":    "debug",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	tmpDir := t.TempDir()
	loader := &FileLoader{ConfigPath: filepath.Join(tmpDir, "config.yaml")}
	
	config, err := ResolveConfigWithLoader(CLIOptions{}, loader)
	if err != nil {
		t.Fatalf("ResolveConfigWithLoader() error = %v", err)
	}

	if config.APIKey != "env-key" {
		t.Errorf("OPENAI_API_KEY not read correctly, got %v", config.APIKey)
	}
	if config.APIBase != "https://env.example.com" {
		t.Errorf("OPENAI_API_BASE not read correctly, got %v", config.APIBase)
	}
	if config.Model != "env-model" {
		t.Errorf("OPENAI_MODEL not read correctly, got %v", config.Model)
	}
	if config.LogLevel != "debug" {
		t.Errorf("SORTPATH_LOG_LEVEL not read correctly, got %v", config.LogLevel)
	}
}

func TestConfig_URLValidation(t *testing.T) {
	tests := []struct {
		name    string
		apiBase string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid https URL",
			apiBase: "https://api.openai.com/v1",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			apiBase: "http://localhost:8080/v1",
			wantErr: false,
		},
		{
			name:    "invalid scheme",
			apiBase: "ftp://api.openai.com/v1",
			wantErr: true,
			errMsg:  "must use http or https scheme",
		},
		{
			name:    "no scheme",
			apiBase: "api.openai.com/v1",
			wantErr: true,
			errMsg:  "must use http or https scheme",
		},
		{
			name:    "malformed URL",
			apiBase: "://invalid-url",
			wantErr: true,
			errMsg:  "invalid API base URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				APIKey:   "test-key",
				APIBase:  tt.apiBase,
				Model:    "gpt-3.5-turbo",
				TreePath: ".",
				LogLevel: "info",
			}

			err := config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}