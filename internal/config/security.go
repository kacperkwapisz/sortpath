package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SanitizePath validates and sanitizes file paths to prevent directory traversal attacks
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path contains directory traversal sequences: %s", path)
	}

	// Check for absolute paths that might be dangerous
	if filepath.IsAbs(cleanPath) {
		// Allow absolute paths but warn about potential security implications
		// This is needed for legitimate use cases like /usr/local/bin
		return cleanPath, nil
	}

	return cleanPath, nil
}

// ValidateConfigKey ensures the configuration key is one of the allowed values
func ValidateConfigKey(key string) error {
	allowedKeys := map[string]bool{
		"api-key":   true,
		"api-base":  true,
		"model":     true,
		"tree-path": true,
		"log-level": true,
	}

	if !allowedKeys[key] {
		return fmt.Errorf("unknown config key: %s. Valid keys: api-key, api-base, model, tree-path, log-level", key)
	}

	return nil
}

// SanitizeConfigValue sanitizes configuration values based on their type
func SanitizeConfigValue(key, value string) (string, error) {
	// Trim whitespace
	value = strings.TrimSpace(value)

	switch key {
	case "api-key":
		// API keys should not contain newlines or control characters
		if strings.ContainsAny(value, "\n\r\t") {
			return "", fmt.Errorf("API key contains invalid characters")
		}
		return value, nil

	case "api-base":
		// URL validation is handled in Config.Validate()
		return value, nil

	case "model":
		// Model names should be alphanumeric with hyphens and dots
		if value != "" && !isValidModelName(value) {
			return "", fmt.Errorf("model name contains invalid characters. Use alphanumeric characters, hyphens, and dots only")
		}
		return value, nil

	case "tree-path":
		// Path sanitization
		return SanitizePath(value)

	case "log-level":
		// Normalize to lowercase
		normalized := strings.ToLower(value)
		
		// Validate log level
		if normalized != "" {
			validLogLevels := []string{"debug", "info", "error"}
			valid := false
			for _, level := range validLogLevels {
				if normalized == level {
					valid = true
					break
				}
			}
			if !valid {
				return "", fmt.Errorf("invalid log level '%s'. Valid options: %s", value, strings.Join(validLogLevels, ", "))
			}
		}
		
		return normalized, nil

	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// isValidModelName checks if a model name contains only allowed characters
func isValidModelName(name string) bool {
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || r == '-' || r == '.' || r == '_') {
			return false
		}
	}
	return true
}

// RedactSensitiveValue masks sensitive configuration values for display
func RedactSensitiveValue(key, value string) string {
	switch key {
	case "api-key":
		if len(value) <= 8 {
			return "***"
		}
		// Show first 4 and last 4 characters
		return value[:4] + "..." + value[len(value)-4:]
	default:
		return value
	}
}

// SecureFileOperations provides utilities for secure file operations
type SecureFileOperations struct{}

// EnsureSecurePermissions ensures a file has secure permissions (600)
func (s *SecureFileOperations) EnsureSecurePermissions(path string) error {
	return os.Chmod(path, 0600)
}

// CreateSecureFile creates a file with secure permissions
func (s *SecureFileOperations) CreateSecureFile(path string) (*os.File, error) {
	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", path, err)
	}

	// Set secure permissions
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to set secure permissions on %s: %w", path, err)
	}

	return file, nil
}

// AtomicWrite performs an atomic write operation to prevent corruption
func (s *SecureFileOperations) AtomicWrite(path string, data []byte) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create a temporary file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp-sortpath-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on failure
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	// Set secure permissions on temp file
	if err := tmpFile.Chmod(0600); err != nil {
		return fmt.Errorf("failed to set permissions on temporary file: %w", err)
	}

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close temp file before rename
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically move temp file to final location
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to move temporary file to final location: %w", err)
	}

	// Clear the defer cleanup since we successfully moved the file
	tmpPath = ""
	return nil
}

// ValidateFilePermissions checks if a file has secure permissions
func (s *SecureFileOperations) ValidateFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	mode := info.Mode()
	if mode.Perm() != 0600 {
		return fmt.Errorf("file %s has insecure permissions %o, expected 0600", path, mode.Perm())
	}

	return nil
}

// DefaultSecureFileOps provides a default instance of SecureFileOperations
var DefaultSecureFileOps = &SecureFileOperations{}