package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

// Validate checks if the configuration is valid and returns helpful error messages
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required. Set it with: sortpath config set api-key YOUR_KEY")
	}

	if c.APIBase == "" {
		return fmt.Errorf("API base URL is required. Set it with: sortpath config set api-base https://api.openai.com/v1")
	}

	// Validate API base URL format
	parsedURL, err := url.Parse(c.APIBase)
	if err != nil {
		return fmt.Errorf("invalid API base URL '%s': %v. Use format: https://api.openai.com/v1", c.APIBase, err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("API base URL must use http or https scheme, got '%s'. Use format: https://api.openai.com/v1", c.APIBase)
	}

	if c.Model == "" {
		return fmt.Errorf("model is required. Set it with: sortpath config set model gpt-3.5-turbo")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "error"}
	if c.LogLevel != "" {
		valid := false
		for _, level := range validLogLevels {
			if strings.ToLower(c.LogLevel) == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log level '%s'. Valid options: %s", c.LogLevel, strings.Join(validLogLevels, ", "))
		}
	}

	// Validate tree path exists and is readable
	if c.TreePath != "" && c.TreePath != "." {
		if _, err := os.Stat(c.TreePath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("tree path '%s' does not exist. Use an existing directory path", c.TreePath)
			}
			return fmt.Errorf("cannot access tree path '%s': %v", c.TreePath, err)
		}
	}

	return nil
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
			// File doesn't exist, return empty config (will use defaults)
			return &Config{}, nil
		}
		if os.IsPermission(err) {
			// Handle permission errors based on environment
			return nil, DefaultEdgeCaseHandler.HandlePermissionError(fl.ConfigPath, "read")
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var c Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&c); err != nil {
		// Handle corrupted config file
		recoveredConfig, recoverErr := DefaultEdgeCaseHandler.HandleCorruptedConfig(fl.ConfigPath, err)
		if recoverErr != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
		// Return recovered config (with defaults) but log the original error
		return recoveredConfig, nil
	}
	return &c, nil
}

// Save writes configuration to file with secure permissions using atomic operations
func (fl *FileLoader) Save(c *Config) error {
	// Marshal the config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Use atomic write for safety
	if err := DefaultSecureFileOps.AtomicWrite(fl.ConfigPath, data); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
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
	if fileConfig == nil {
		fileConfig = &Config{} // Use empty config if loading failed
	}

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

	// Validate the resolved configuration
	if err := resolved.Validate(); err != nil {
		return nil, err
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