package config

import (
	"os"
	"strings"
)

// EnvironmentDetector provides utilities for detecting the runtime environment
type EnvironmentDetector struct{}

// IsNonInteractive detects if the application is running in a non-interactive environment
func (e *EnvironmentDetector) IsNonInteractive() bool {
	// Check if stdin is not a terminal
	if fi, err := os.Stdin.Stat(); err == nil {
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			return true
		}
	}

	// Check common CI/CD environment variables
	ciEnvVars := []string{
		"CI",           // Generic CI indicator
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_URL",
		"BUILDKITE",
		"CIRCLECI",
		"TRAVIS",
		"DRONE",
		"TEAMCITY_VERSION",
		"TF_BUILD",     // Azure DevOps
		"CODEBUILD_BUILD_ID", // AWS CodeBuild
	}

	for _, envVar := range ciEnvVars {
		if value := os.Getenv(envVar); value != "" && strings.ToLower(value) != "false" {
			return true
		}
	}

	// Check if running in a container
	if e.isRunningInContainer() {
		return true
	}

	// Check if TERM is not set or set to "dumb"
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return true
	}

	return false
}

// isRunningInContainer detects if running inside a container
func (e *EnvironmentDetector) isRunningInContainer() bool {
	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for other container indicators
	containerEnvVars := []string{
		"DOCKER_CONTAINER",
		"KUBERNETES_SERVICE_HOST",
		"CONTAINER",
	}

	for _, envVar := range containerEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// GetEnvironmentType returns a string describing the environment type
func (e *EnvironmentDetector) GetEnvironmentType() string {
	if e.IsNonInteractive() {
		if os.Getenv("CI") != "" {
			return "ci"
		}
		if e.isRunningInContainer() {
			return "container"
		}
		return "non-interactive"
	}
	return "interactive"
}

// ShouldPromptUser determines if the application should prompt for user input
func (e *EnvironmentDetector) ShouldPromptUser() bool {
	return !e.IsNonInteractive()
}

// DefaultEnvironmentDetector provides a default instance
var DefaultEnvironmentDetector = &EnvironmentDetector{}

// EdgeCaseHandler provides utilities for handling edge cases
type EdgeCaseHandler struct {
	envDetector *EnvironmentDetector
}

// NewEdgeCaseHandler creates a new EdgeCaseHandler
func NewEdgeCaseHandler() *EdgeCaseHandler {
	return &EdgeCaseHandler{
		envDetector: DefaultEnvironmentDetector,
	}
}

// HandleMissingConfig provides fallback behavior when config is missing or corrupted
func (h *EdgeCaseHandler) HandleMissingConfig() *Config {
	// Return a config with sensible defaults
	return &Config{
		APIBase:  "https://api.openai.com/v1",
		Model:    "gpt-3.5-turbo",
		TreePath: ".",
		LogLevel: "info",
		// APIKey is intentionally left empty - it must be provided by user
	}
}

// HandleCorruptedConfig attempts to recover from a corrupted config file
func (h *EdgeCaseHandler) HandleCorruptedConfig(configPath string, err error) (*Config, error) {
	// Log the corruption (in a real implementation, this would use the logger)
	// For now, we'll just return defaults and let the user know
	
	if !h.envDetector.IsNonInteractive() {
		// In interactive mode, we could potentially prompt the user
		// For now, just return defaults
	}
	
	// Return defaults and let the validation catch missing required fields
	return h.HandleMissingConfig(), nil
}

// HandlePermissionError provides guidance for permission-related errors
func (h *EdgeCaseHandler) HandlePermissionError(path string, operation string) error {
	envType := h.envDetector.GetEnvironmentType()
	
	switch envType {
	case "ci":
		return &ConfigError{
			Code:    "permission_denied_ci",
			Message: "Permission denied in CI environment",
			Cause:   nil,
			Context: map[string]interface{}{
				"path":      path,
				"operation": operation,
				"suggestion": "Ensure the CI environment has proper permissions or use environment variables for configuration",
			},
		}
	case "container":
		return &ConfigError{
			Code:    "permission_denied_container",
			Message: "Permission denied in container",
			Cause:   nil,
			Context: map[string]interface{}{
				"path":      path,
				"operation": operation,
				"suggestion": "Mount a writable volume or use environment variables for configuration",
			},
		}
	default:
		return &ConfigError{
			Code:    "permission_denied",
			Message: "Permission denied",
			Cause:   nil,
			Context: map[string]interface{}{
				"path":      path,
				"operation": operation,
				"suggestion": "Check file permissions or run with appropriate privileges",
			},
		}
	}
}

// ConfigError represents a configuration-related error with context
type ConfigError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]interface{}
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// DefaultEdgeCaseHandler provides a default instance
var DefaultEdgeCaseHandler = NewEdgeCaseHandler()