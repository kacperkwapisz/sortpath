package errors

import (
	"fmt"
	"strings"
)

// AppError represents an application error with context and user-friendly messaging
type AppError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Cause   error                  `json:"-"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap implements the unwrapper interface for Go 1.13+ error handling
func (e *AppError) Unwrap() error {
	return e.Cause
}

// UserMessage returns a user-friendly error message with emoji
func (e *AppError) UserMessage() string {
	return fmt.Sprintf("❌ %s", e.Message)
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// ConfigError creates a configuration-related error
func ConfigError(msg string, cause error) *AppError {
	return &AppError{
		Code:    "CONFIG_ERROR",
		Message: msg,
		Cause:   cause,
	}
}

// APIError creates an API-related error
func APIError(msg string, cause error) *AppError {
	return &AppError{
		Code:    "API_ERROR",
		Message: msg,
		Cause:   cause,
	}
}

// FSError creates a filesystem-related error
func FSError(msg string, path string, cause error) *AppError {
	err := &AppError{
		Code:    "FS_ERROR",
		Message: msg,
		Cause:   cause,
	}
	if path != "" {
		err.WithContext("path", path)
	}
	return err
}

// InstallError creates an installation-related error
func InstallError(msg string, cause error) *AppError {
	return &AppError{
		Code:    "INSTALL_ERROR",
		Message: msg,
		Cause:   cause,
	}
}

// ValidationError creates a validation-related error
func ValidationError(msg string, field string) *AppError {
	err := &AppError{
		Code:    "VALIDATION_ERROR",
		Message: msg,
		Cause:   nil,
	}
	if field != "" {
		err.WithContext("field", field)
	}
	return err
}

// NetworkError creates a network-related error
func NetworkError(msg string, cause error) *AppError {
	return &AppError{
		Code:    "NETWORK_ERROR",
		Message: msg,
		Cause:   cause,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, msg string) *AppError {
	if err == nil {
		return nil
	}
	
	// If it's already an AppError, preserve the original code
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:    appErr.Code,
			Message: fmt.Sprintf("%s: %s", msg, appErr.Message),
			Cause:   appErr.Cause,
			Context: appErr.Context,
		}
	}
	
	return &AppError{
		Code:    "WRAPPED_ERROR",
		Message: msg,
		Cause:   err,
	}
}

// IsType checks if an error is of a specific type
func IsType(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// GetContext retrieves context value from an AppError
func GetContext(err error, key string) (interface{}, bool) {
	if appErr, ok := err.(*AppError); ok && appErr.Context != nil {
		value, exists := appErr.Context[key]
		return value, exists
	}
	return nil, false
}

// FormatUserError formats an error for user display with helpful suggestions
func FormatUserError(err error) string {
	if err == nil {
		return ""
	}
	
	appErr, ok := err.(*AppError)
	if !ok {
		return fmt.Sprintf("❌ %v", err)
	}
	
	var parts []string
	parts = append(parts, appErr.UserMessage())
	
	// Add context-specific suggestions
	switch appErr.Code {
	case "CONFIG_ERROR":
		if strings.Contains(appErr.Message, "API key") {
			parts = append(parts, "💡 Set your API key with: sortpath config set api-key YOUR_KEY")
		}
		if strings.Contains(appErr.Message, "config file") {
			parts = append(parts, "💡 Create config with: sortpath config init")
		}
	case "API_ERROR":
		if strings.Contains(appErr.Message, "401") || strings.Contains(appErr.Message, "unauthorized") {
			parts = append(parts, "💡 Check your API key with: sortpath config get api-key")
		}
		if strings.Contains(appErr.Message, "network") || strings.Contains(appErr.Message, "timeout") {
			parts = append(parts, "💡 Check your internet connection and try again")
		}
	case "FS_ERROR":
		if path, exists := GetContext(err, "path"); exists {
			if strings.Contains(appErr.Message, "permission") {
				parts = append(parts, fmt.Sprintf("💡 Try: chmod +r %v", path))
			}
			if strings.Contains(appErr.Message, "not found") {
				parts = append(parts, fmt.Sprintf("💡 Check if path exists: %v", path))
			}
		}
	case "INSTALL_ERROR":
		if strings.Contains(appErr.Message, "permission") {
			parts = append(parts, "💡 Try running with sudo or choose a different install path")
		}
	}
	
	return strings.Join(parts, "\n")
}