package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "error without cause",
			err: &AppError{
				Code:    "TEST_ERROR",
				Message: "test message",
			},
			expected: "test message",
		},
		{
			name: "error with cause",
			err: &AppError{
				Code:    "TEST_ERROR",
				Message: "test message",
				Cause:   errors.New("underlying error"),
			},
			expected: "test message: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("AppError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestAppError_UserMessage(t *testing.T) {
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "test message",
	}

	expected := "❌ test message"
	if got := err.UserMessage(); got != expected {
		t.Errorf("AppError.UserMessage() = %v, want %v", got, expected)
	}
}

func TestAppError_WithContext(t *testing.T) {
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "test message",
	}

	err.WithContext("key1", "value1")
	err.WithContext("key2", 42)

	if len(err.Context) != 2 {
		t.Errorf("Expected 2 context items, got %d", len(err.Context))
	}

	if err.Context["key1"] != "value1" {
		t.Errorf("Expected context key1 = 'value1', got %v", err.Context["key1"])
	}

	if err.Context["key2"] != 42 {
		t.Errorf("Expected context key2 = 42, got %v", err.Context["key2"])
	}
}

func TestConfigError(t *testing.T) {
	cause := errors.New("config file not found")
	err := ConfigError("Configuration error", cause)

	if err.Code != "CONFIG_ERROR" {
		t.Errorf("Expected code CONFIG_ERROR, got %s", err.Code)
	}

	if err.Message != "Configuration error" {
		t.Errorf("Expected message 'Configuration error', got %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Expected cause to be preserved")
	}
}

func TestAPIError(t *testing.T) {
	cause := errors.New("network timeout")
	err := APIError("API request failed", cause)

	if err.Code != "API_ERROR" {
		t.Errorf("Expected code API_ERROR, got %s", err.Code)
	}

	if err.Message != "API request failed" {
		t.Errorf("Expected message 'API request failed', got %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Expected cause to be preserved")
	}
}

func TestFSError(t *testing.T) {
	cause := errors.New("permission denied")
	path := "/test/path"
	err := FSError("Cannot read directory", path, cause)

	if err.Code != "FS_ERROR" {
		t.Errorf("Expected code FS_ERROR, got %s", err.Code)
	}

	if err.Message != "Cannot read directory" {
		t.Errorf("Expected message 'Cannot read directory', got %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Expected cause to be preserved")
	}

	if err.Context["path"] != path {
		t.Errorf("Expected path context to be %s, got %v", path, err.Context["path"])
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("Invalid field value", "api_key")

	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("Expected code VALIDATION_ERROR, got %s", err.Code)
	}

	if err.Message != "Invalid field value" {
		t.Errorf("Expected message 'Invalid field value', got %s", err.Message)
	}

	if err.Context["field"] != "api_key" {
		t.Errorf("Expected field context to be 'api_key', got %v", err.Context["field"])
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		msg      string
		expected *AppError
	}{
		{
			name:     "wrap nil error",
			err:      nil,
			msg:      "test message",
			expected: nil,
		},
		{
			name: "wrap regular error",
			err:  errors.New("original error"),
			msg:  "wrapped message",
			expected: &AppError{
				Code:    "WRAPPED_ERROR",
				Message: "wrapped message",
				Cause:   errors.New("original error"),
			},
		},
		{
			name: "wrap AppError",
			err: &AppError{
				Code:    "ORIGINAL_ERROR",
				Message: "original message",
				Cause:   errors.New("root cause"),
			},
			msg: "wrapped message",
			expected: &AppError{
				Code:    "ORIGINAL_ERROR",
				Message: "wrapped message: original message",
				Cause:   errors.New("root cause"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.err, tt.msg)

			if tt.expected == nil {
				if got != nil {
					t.Errorf("Expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("Expected AppError, got nil")
				return
			}

			if got.Code != tt.expected.Code {
				t.Errorf("Expected code %s, got %s", tt.expected.Code, got.Code)
			}

			if got.Message != tt.expected.Message {
				t.Errorf("Expected message %s, got %s", tt.expected.Message, got.Message)
			}
		})
	}
}

func TestIsType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		expected bool
	}{
		{
			name:     "matching AppError",
			err:      &AppError{Code: "TEST_ERROR"},
			code:     "TEST_ERROR",
			expected: true,
		},
		{
			name:     "non-matching AppError",
			err:      &AppError{Code: "TEST_ERROR"},
			code:     "OTHER_ERROR",
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			code:     "TEST_ERROR",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsType(tt.err, tt.code); got != tt.expected {
				t.Errorf("IsType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetContext(t *testing.T) {
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Context: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	// Test existing key
	value, exists := GetContext(err, "key1")
	if !exists {
		t.Errorf("Expected key1 to exist")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test non-existing key
	_, exists = GetContext(err, "nonexistent")
	if exists {
		t.Errorf("Expected nonexistent key to not exist")
	}

	// Test regular error
	regularErr := errors.New("regular error")
	_, exists = GetContext(regularErr, "key1")
	if exists {
		t.Errorf("Expected regular error to have no context")
	}
}

func TestFormatUserError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "nil error",
			err:      nil,
			contains: []string{},
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			contains: []string{"❌", "regular error"},
		},
		{
			name:     "config error with API key",
			err:      ConfigError("API key is required", nil),
			contains: []string{"❌", "API key is required", "💡", "sortpath config set api-key"},
		},
		{
			name:     "API error with 401",
			err:      APIError("Request failed (401 Unauthorized)", nil),
			contains: []string{"❌", "401 Unauthorized", "💡", "sortpath config get api-key"},
		},
		{
			name:     "FS error with permission",
			err:      FSError("Cannot read directory (permission denied)", "/test/path", nil),
			contains: []string{"❌", "permission denied", "💡", "chmod +r", "/test/path"},
		},
		{
			name:     "install error with permission",
			err:      InstallError("Installation failed (permission denied)", nil),
			contains: []string{"❌", "permission denied", "💡", "sudo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatUserError(tt.err)

			if tt.err == nil {
				if got != "" {
					t.Errorf("Expected empty string for nil error, got %s", got)
				}
				return
			}

			for _, expected := range tt.contains {
				if !strings.Contains(got, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, got)
				}
			}
		})
	}
}