package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kacperkwapisz/sortpath/internal/config"
)

func TestFileLoader_LoadNonExistentFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	loader := &config.FileLoader{ConfigPath: configPath}
	cfg, err := loader.Load()

	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
	}

	// Should return empty config
	expected := &config.Config{}
	if *cfg != *expected {
		t.Errorf("Expected empty config, got: %+v", cfg)
	}
}

func TestFileLoader_LoadValidFile(t *testing.T) {
	// Create a temporary directory and config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Write a valid config file
	configContent := `api_key: test-key
api_base: https://api.example.com/v1
model: gpt-4
tree_path: /test/path
log_level: debug
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	loader := &config.FileLoader{ConfigPath: configPath}
	cfg, err := loader.Load()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := &config.Config{
		APIKey:   "test-key",
		APIBase:  "https://api.example.com/v1",
		Model:    "gpt-4",
		TreePath: "/test/path",
		LogLevel: "debug",
	}

	if *cfg != *expected {
		t.Errorf("Expected config %+v, got %+v", expected, cfg)
	}
}

func TestFileLoader_LoadInvalidFile(t *testing.T) {
	// Create a temporary directory and invalid config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Write an invalid YAML file
	invalidContent := `api_key: test-key
invalid_yaml: [unclosed bracket
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	loader := &config.FileLoader{ConfigPath: configPath}
	cfg, err := loader.Load()

	// With the new edge case handler, corrupted files should return defaults, not error
	if err != nil {
		t.Errorf("Unexpected error for invalid YAML: %v", err)
	}
	if cfg == nil {
		t.Error("Expected default config for invalid YAML, got nil")
	}
	// Should return defaults
	if cfg.APIBase != "https://api.openai.com/v1" {
		t.Errorf("Expected default APIBase for invalid YAML, got: %v", cfg.APIBase)
	}
}

func TestFileLoader_Save(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	loader := &config.FileLoader{ConfigPath: configPath}
	cfg := &config.Config{
		APIKey:   "test-key",
		APIBase:  "https://api.example.com/v1",
		Model:    "gpt-4",
		TreePath: "/test/path",
		LogLevel: "debug",
	}

	err := loader.Save(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
	}

	// Verify content by loading it back
	loadedConfig, err := loader.Load()
	if err != nil {
		t.Errorf("Failed to load saved config: %v", err)
	}

	if *loadedConfig != *cfg {
		t.Errorf("Loaded config %+v doesn't match saved config %+v", loadedConfig, cfg)
	}
}

func TestFileLoader_SaveCreateDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nested", "dir", "config.yaml")

	loader := &config.FileLoader{ConfigPath: configPath}
	cfg := &config.Config{
		APIKey: "test-key",
	}

	err := loader.Save(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	// Verify directory permissions
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Failed to stat config directory: %v", err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("Expected directory permissions 0700, got %o", info.Mode().Perm())
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test that convenience functions work
	// This is mainly to ensure they don't panic and use the default loader
	
	// Load should not error on non-existent file (except for permission issues)
	_, err := config.Load()
	if err != nil && !os.IsPermission(err) && !os.IsNotExist(err) {
		t.Errorf("Load() returned unexpected error: %v", err)
	}
	
	// Save should work with a valid config (though it might fail due to permissions in test env)
	cfg := &config.Config{APIKey: "test"}
	err = config.Save(cfg)
	if err != nil && !os.IsPermission(err) {
		t.Errorf("Save() returned unexpected error: %v", err)
	}
}