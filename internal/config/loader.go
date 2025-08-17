package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration with only essential fields
type Config struct {
	APIKey   string `yaml:"api_key"`
	APIBase  string `yaml:"api_base"`
	Model    string `yaml:"model"`
	TreePath string `yaml:"tree_path"`
	LogLevel string `yaml:"log_level"`
}

// Loader interface for configuration operations
type Loader interface {
	Load() (*Config, error)
	Save(*Config) error
}

// FileLoader implements the Loader interface for file-based configuration
type FileLoader struct {
	ConfigPath string
}

// NewFileLoader creates a new FileLoader with the default config path
func NewFileLoader() *FileLoader {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sortpath", "config.yaml")
	return &FileLoader{ConfigPath: configPath}
}

// Load reads configuration from file, returns empty config if file doesn't exist
func (fl *FileLoader) Load() (*Config, error) {
	f, err := os.Open(fl.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var c Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &c, nil
}

// Save writes configuration to file with secure permissions
func (fl *FileLoader) Save(c *Config) error {
	dir := filepath.Dir(fl.ConfigPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	f, err := os.Create(fl.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	// Set secure file permissions (600)
	if err := f.Chmod(0600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	enc := yaml.NewEncoder(f)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Default configuration values
var defaults = Config{
	APIBase:  "https://api.openai.com/v1",
	Model:    "gpt-3.5-turbo",
	TreePath: ".",
	LogLevel: "info",
}

// Load is a convenience function that uses the default FileLoader
func Load() (*Config, error) {
	loader := NewFileLoader()
	return loader.Load()
}

// Save is a convenience function that uses the default FileLoader
func Save(c *Config) error {
	loader := NewFileLoader()
	return loader.Save(c)
}

// CLIOptions represents command-line configuration options
type CLIOptions struct {
	APIKey   string
	APIBase  string
	Model    string
	TreePath string
	LogLevel string
}

// ResolveConfig resolves configuration with priority: CLI > ENV > file > defaults
func ResolveConfig(opts CLIOptions) (*Config, error) {
	return ResolveConfigWithLoader(opts, NewFileLoader())
}

// ResolveConfigWithLoader resolves configuration using a custom loader (useful for testing)
func ResolveConfigWithLoader(opts CLIOptions, loader Loader) (*Config, error) {
	// Load from file first
	fileConfig, _ := loader.Load()

	// Apply priority resolution: CLI > ENV > file > defaults
	resolved := &Config{
		APIKey:   resolveValue(opts.APIKey, os.Getenv("OPENAI_API_KEY"), fileConfig.APIKey, ""),
		APIBase:  resolveValue(opts.APIBase, os.Getenv("OPENAI_API_BASE"), fileConfig.APIBase, defaults.APIBase),
		Model:    resolveValue(opts.Model, os.Getenv("OPENAI_MODEL"), fileConfig.Model, defaults.Model),
		TreePath: resolveValue(opts.TreePath, os.Getenv("SORTPATH_FOLDER_TREE"), fileConfig.TreePath, defaults.TreePath),
		LogLevel: resolveValue(opts.LogLevel, os.Getenv("SORTPATH_LOG_LEVEL"), fileConfig.LogLevel, defaults.LogLevel),
	}

	// Apply default for TreePath if still empty
	if resolved.TreePath == "." || resolved.TreePath == "" {
		if wd, err := os.Getwd(); err == nil {
			resolved.TreePath = wd
		} else {
			resolved.TreePath = "."
		}
	}

	// Validate required fields
	if resolved.APIKey == "" {
		return nil, fmt.Errorf("API key is required. Set it with: sortpath config set api-key YOUR_KEY")
	}
	if resolved.APIBase == "" {
		return nil, fmt.Errorf("API base URL is required. Set it with: sortpath config set api-base https://api.openai.com/v1")
	}
	if resolved.Model == "" {
		return nil, fmt.Errorf("model is required. Set it with: sortpath config set model gpt-3.5-turbo")
	}

	return resolved, nil
}

// resolveValue applies priority resolution for a single config value
func resolveValue(cli, env, file, defaultVal string) string {
	if cli != "" {
		return cli
	}
	if env != "" {
		return env
	}
	if file != "" {
		return file
	}
	return defaultVal
}