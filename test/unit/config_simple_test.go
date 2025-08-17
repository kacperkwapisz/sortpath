package unit

import (
	"testing"

	"github.com/kacperkwapisz/sortpath/internal/config"
)

func TestConfigStruct(t *testing.T) {
	cfg := config.Config{
		APIKey:   "test-key",
		APIBase:  "https://api.example.com/v1",
		Model:    "gpt-4",
		TreePath: "/test/path",
		LogLevel: "debug",
	}

	if cfg.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got '%s'", cfg.APIKey)
	}
	if cfg.APIBase != "https://api.example.com/v1" {
		t.Errorf("Expected APIBase 'https://api.example.com/v1', got '%s'", cfg.APIBase)
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("Expected Model 'gpt-4', got '%s'", cfg.Model)
	}
	if cfg.TreePath != "/test/path" {
		t.Errorf("Expected TreePath '/test/path', got '%s'", cfg.TreePath)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestNewFileLoader(t *testing.T) {
	loader := config.NewFileLoader()
	if loader == nil {
		t.Error("NewFileLoader() returned nil")
	}
}