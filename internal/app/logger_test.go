package app

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "debug"},
		{LogLevelInfo, "info"},
		{LogLevelError, "error"},
		{LogLevelSilent, "silent"},
		{LogLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"DEBUG", LogLevelDebug},
		{"info", LogLevelInfo},
		{"INFO", LogLevelInfo},
		{"error", LogLevelError},
		{"ERROR", LogLevelError},
		{"silent", LogLevelSilent},
		{"SILENT", LogLevelSilent},
		{"invalid", LogLevelInfo}, // defaults to info
		{"", LogLevelInfo},        // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseLogLevel(tt.input); got != tt.expected {
				t.Errorf("ParseLogLevel(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStandardLogger_LogLevels(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger := NewLoggerWithOutput(LogLevelInfo, &stdout, &stderr)

	// Test that debug messages are not logged at info level
	logger.Debug("debug message")
	if stdout.String() != "" {
		t.Errorf("Expected no debug output at info level, got: %s", stdout.String())
	}

	// Test that info messages are logged at info level
	stdout.Reset()
	logger.Info("info message")
	if !strings.Contains(stdout.String(), "info message") {
		t.Errorf("Expected info message in output, got: %s", stdout.String())
	}

	// Test that error messages are logged at info level
	stderr.Reset()
	logger.Error("error message")
	if !strings.Contains(stderr.String(), "error message") {
		t.Errorf("Expected error message in stderr, got: %s", stderr.String())
	}
}

func TestStandardLogger_DebugLevel(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger := NewLoggerWithOutput(LogLevelDebug, &stdout, &stderr)

	// Test that all messages are logged at debug level
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Error("error message")

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	if !strings.Contains(stdoutStr, "debug message") {
		t.Errorf("Expected debug message in stdout, got: %s", stdoutStr)
	}

	if !strings.Contains(stdoutStr, "info message") {
		t.Errorf("Expected info message in stdout, got: %s", stdoutStr)
	}

	if !strings.Contains(stderrStr, "error message") {
		t.Errorf("Expected error message in stderr, got: %s", stderrStr)
	}
}

func TestStandardLogger_ErrorLevel(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger := NewLoggerWithOutput(LogLevelError, &stdout, &stderr)

	// Test that only error messages are logged at error level
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Error("error message")

	if stdout.String() != "" {
		t.Errorf("Expected no stdout output at error level, got: %s", stdout.String())
	}

	if !strings.Contains(stderr.String(), "error message") {
		t.Errorf("Expected error message in stderr, got: %s", stderr.String())
	}
}

func TestStandardLogger_SilentLevel(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger := NewLoggerWithOutput(LogLevelSilent, &stdout, &stderr)

	// Test that no messages are logged at silent level
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Error("error message")

	if stdout.String() != "" {
		t.Errorf("Expected no stdout output at silent level, got: %s", stdout.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected no stderr output at silent level, got: %s", stderr.String())
	}
}

func TestStandardLogger_SetGetLevel(t *testing.T) {
	logger := NewLogger(LogLevelInfo)

	if logger.GetLevel() != LogLevelInfo {
		t.Errorf("Expected initial level to be Info, got %v", logger.GetLevel())
	}

	logger.SetLevel(LogLevelDebug)
	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("Expected level to be Debug after setting, got %v", logger.GetLevel())
	}
}

func TestStandardLogger_RedactSensitiveData(t *testing.T) {
	var stdout bytes.Buffer
	logger := NewLoggerWithOutput(LogLevelDebug, &stdout, &stdout)

	tests := []struct {
		name     string
		message  string
		contains string
		notContains string
	}{
		{
			name:        "redact api_key with equals",
			message:     "Config loaded with api_key=sk-1234567890abcdef",
			contains:    "[REDACTED]",
			notContains: "sk-1234567890abcdef",
		},
		{
			name:        "redact api_key with colon",
			message:     "API configuration: api_key: sk-1234567890abcdef",
			contains:    "[REDACTED]",
			notContains: "sk-1234567890abcdef",
		},
		{
			name:        "redact token",
			message:     "Authorization token=bearer_token_12345",
			contains:    "[REDACTED]",
			notContains: "bearer_token_12345",
		},
		{
			name:        "redact password",
			message:     "Login with password: mySecretPassword123",
			contains:    "[REDACTED]",
			notContains: "mySecretPassword123",
		},
		{
			name:        "preserve non-sensitive data",
			message:     "Processing file: /path/to/file.txt",
			contains:    "/path/to/file.txt",
			notContains: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout.Reset()
			logger.Info(tt.message)
			output := stdout.String()

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %s", tt.contains, output)
			}

			if tt.notContains != "" && strings.Contains(output, tt.notContains) {
				t.Errorf("Expected output to NOT contain %q, got: %s", tt.notContains, output)
			}
		})
	}
}

func TestStandardLogger_FormatMessage(t *testing.T) {
	var stdout bytes.Buffer
	logger := NewLoggerWithOutput(LogLevelInfo, &stdout, &stdout)

	// Test message without arguments
	logger.Info("simple message")
	if !strings.Contains(stdout.String(), "simple message") {
		t.Errorf("Expected simple message in output")
	}

	// Test message with arguments
	stdout.Reset()
	logger.Info("formatted message: %s %d", "test", 42)
	output := stdout.String()
	if !strings.Contains(output, "formatted message: test 42") {
		t.Errorf("Expected formatted message in output, got: %s", output)
	}
}

func TestNewLoggerFromEnv(t *testing.T) {
	// Test with SORTPATH_LOG_LEVEL
	os.Setenv("SORTPATH_LOG_LEVEL", "debug")
	defer os.Unsetenv("SORTPATH_LOG_LEVEL")

	logger := NewLoggerFromEnv()
	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("Expected debug level from SORTPATH_LOG_LEVEL, got %v", logger.GetLevel())
	}

	// Test with LOG_LEVEL fallback
	os.Unsetenv("SORTPATH_LOG_LEVEL")
	os.Setenv("LOG_LEVEL", "error")
	defer os.Unsetenv("LOG_LEVEL")

	logger = NewLoggerFromEnv()
	if logger.GetLevel() != LogLevelError {
		t.Errorf("Expected error level from LOG_LEVEL, got %v", logger.GetLevel())
	}

	// Test with no environment variables (should default to info)
	os.Unsetenv("LOG_LEVEL")
	logger = NewLoggerFromEnv()
	if logger.GetLevel() != LogLevelInfo {
		t.Errorf("Expected info level as default, got %v", logger.GetLevel())
	}
}

func TestStandardLogger_TimedOperation(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger := NewLoggerWithOutput(LogLevelDebug, &stdout, &stderr)

	// Test successful operation
	err := logger.TimedOperation("test operation", func() error {
		time.Sleep(1 * time.Millisecond) // Small delay to ensure measurable duration
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error from successful operation, got: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Starting operation: test operation") {
		t.Errorf("Expected start message in output")
	}
	if !strings.Contains(output, "Operation completed: test operation") {
		t.Errorf("Expected completion message in output")
	}

	// Test failed operation
	stdout.Reset()
	stderr.Reset()
	err = logger.TimedOperation("failing operation", func() error {
		return errors.New("test error")
	})

	if err == nil {
		t.Errorf("Expected error from failing operation")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "Operation failed: failing operation") {
		t.Errorf("Expected failure message in stderr, got: %s", errorOutput)
	}
}

func TestContextLogger(t *testing.T) {
	var stdout, stderr bytes.Buffer
	baseLogger := NewLoggerWithOutput(LogLevelDebug, &stdout, &stderr)
	contextLogger := baseLogger.WithContext("TEST")

	contextLogger.Debug("debug message")
	contextLogger.Info("info message")
	contextLogger.Error("error message")

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	if !strings.Contains(stdoutStr, "[TEST] debug message") {
		t.Errorf("Expected context in debug message, got: %s", stdoutStr)
	}

	if !strings.Contains(stdoutStr, "[TEST] info message") {
		t.Errorf("Expected context in info message, got: %s", stdoutStr)
	}

	if !strings.Contains(stderrStr, "[TEST] error message") {
		t.Errorf("Expected context in error message, got: %s", stderrStr)
	}

	// Test that context logger preserves level operations
	contextLogger.SetLevel(LogLevelError)
	if contextLogger.GetLevel() != LogLevelError {
		t.Errorf("Expected context logger to preserve level operations")
	}
}