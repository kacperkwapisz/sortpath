package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// LogLevelDebug enables all log messages
	LogLevelDebug LogLevel = iota
	// LogLevelInfo enables info and error messages
	LogLevelInfo
	// LogLevelError enables only error messages
	LogLevelError
	// LogLevelSilent disables all log messages
	LogLevelSilent
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelError:
		return "error"
	case LogLevelSilent:
		return "silent"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "error":
		return LogLevelError
	case "silent":
		return LogLevelSilent
	default:
		return LogLevelInfo // default to info
	}
}

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	SetLevel(level LogLevel)
	GetLevel() LogLevel
}

// StandardLogger implements Logger using Go's standard log package
type StandardLogger struct {
	level      LogLevel
	debugLog   *log.Logger
	infoLog    *log.Logger
	errorLog   *log.Logger
	sensitiveKeys []string
}

// NewLogger creates a new StandardLogger with the specified level
func NewLogger(level LogLevel) *StandardLogger {
	return &StandardLogger{
		level:    level,
		debugLog: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		infoLog:  log.New(os.Stdout, "[INFO]  ", log.LstdFlags),
		errorLog: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		sensitiveKeys: []string{
			"api_key", "apikey", "api-key",
			"token", "password", "secret",
			"authorization", "auth",
		},
	}
}

// NewLoggerWithOutput creates a new StandardLogger with custom output writers
func NewLoggerWithOutput(level LogLevel, stdout, stderr io.Writer) *StandardLogger {
	return &StandardLogger{
		level:    level,
		debugLog: log.New(stdout, "[DEBUG] ", log.LstdFlags),
		infoLog:  log.New(stdout, "[INFO]  ", log.LstdFlags),
		errorLog: log.New(stderr, "[ERROR] ", log.LstdFlags),
		sensitiveKeys: []string{
			"api_key", "apikey", "api-key",
			"token", "password", "secret",
			"authorization", "auth",
		},
	}
}

// Debug logs a debug message if the level allows it
func (l *StandardLogger) Debug(msg string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		formatted := l.formatMessage(msg, args...)
		redacted := l.redactSensitiveData(formatted)
		l.debugLog.Print(redacted)
	}
}

// Info logs an info message if the level allows it
func (l *StandardLogger) Info(msg string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		formatted := l.formatMessage(msg, args...)
		redacted := l.redactSensitiveData(formatted)
		l.infoLog.Print(redacted)
	}
}

// Error logs an error message if the level allows it
func (l *StandardLogger) Error(msg string, args ...interface{}) {
	if l.level <= LogLevelError {
		formatted := l.formatMessage(msg, args...)
		redacted := l.redactSensitiveData(formatted)
		l.errorLog.Print(redacted)
	}
}

// SetLevel sets the logging level
func (l *StandardLogger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel returns the current logging level
func (l *StandardLogger) GetLevel() LogLevel {
	return l.level
}

// formatMessage formats the message with arguments
func (l *StandardLogger) formatMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

// redactSensitiveData removes or masks sensitive information from log messages
func (l *StandardLogger) redactSensitiveData(message string) string {
	result := message
	
	for _, key := range l.sensitiveKeys {
		// Look for patterns like "api_key=value" or "api_key: value"
		patterns := []string{
			fmt.Sprintf("%s=", key),
			fmt.Sprintf("%s:", key),
			fmt.Sprintf(`"%s"`, key),
			fmt.Sprintf("'%s'", key),
		}
		
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(result), strings.ToLower(pattern)) {
				// Replace the value part with [REDACTED]
				result = l.maskSensitiveValue(result, key)
			}
		}
	}
	
	return result
}

// maskSensitiveValue masks sensitive values in the message
func (l *StandardLogger) maskSensitiveValue(message, key string) string {
	lower := strings.ToLower(message)
	lowerKey := strings.ToLower(key)
	
	// Find the key in the message
	keyIndex := strings.Index(lower, lowerKey)
	if keyIndex == -1 {
		return message
	}
	
	// Find the start of the value (after = or :)
	valueStart := keyIndex + len(key)
	for valueStart < len(message) && (message[valueStart] == '=' || message[valueStart] == ':' || message[valueStart] == ' ' || message[valueStart] == '"' || message[valueStart] == '\'') {
		valueStart++
	}
	
	if valueStart >= len(message) {
		return message
	}
	
	// Find the end of the value (space, comma, quote, or end of string)
	valueEnd := valueStart
	inQuotes := false
	quoteChar := byte(0)
	
	for valueEnd < len(message) {
		char := message[valueEnd]
		
		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quoteChar = char
		} else if inQuotes && char == quoteChar {
			valueEnd++
			break
		} else if !inQuotes && (char == ' ' || char == ',' || char == '\n' || char == '\t') {
			break
		}
		valueEnd++
	}
	
	// Replace the value with [REDACTED]
	if valueEnd > valueStart {
		return message[:valueStart] + "[REDACTED]" + message[valueEnd:]
	}
	
	return message
}

// NewLoggerFromEnv creates a logger with level from environment variable
func NewLoggerFromEnv() *StandardLogger {
	levelStr := os.Getenv("SORTPATH_LOG_LEVEL")
	if levelStr == "" {
		levelStr = os.Getenv("LOG_LEVEL")
	}
	if levelStr == "" {
		levelStr = "info" // default
	}
	
	level := ParseLogLevel(levelStr)
	return NewLogger(level)
}

// TimedOperation logs the duration of an operation
func (l *StandardLogger) TimedOperation(operation string, fn func() error) error {
	start := time.Now()
	l.Debug("Starting operation: %s", operation)
	
	err := fn()
	duration := time.Since(start)
	
	if err != nil {
		l.Error("Operation failed: %s (took %v): %v", operation, duration, err)
	} else {
		l.Debug("Operation completed: %s (took %v)", operation, duration)
	}
	
	return err
}

// WithContext returns a logger that includes context in all messages
func (l *StandardLogger) WithContext(context string) Logger {
	return &contextLogger{
		logger:  l,
		context: context,
	}
}

// contextLogger wraps a logger to add context to all messages
type contextLogger struct {
	logger  Logger
	context string
}

func (c *contextLogger) Debug(msg string, args ...interface{}) {
	c.logger.Debug("[%s] %s", c.context, fmt.Sprintf(msg, args...))
}

func (c *contextLogger) Info(msg string, args ...interface{}) {
	c.logger.Info("[%s] %s", c.context, fmt.Sprintf(msg, args...))
}

func (c *contextLogger) Error(msg string, args ...interface{}) {
	c.logger.Error("[%s] %s", c.context, fmt.Sprintf(msg, args...))
}

func (c *contextLogger) SetLevel(level LogLevel) {
	c.logger.SetLevel(level)
}

func (c *contextLogger) GetLevel() LogLevel {
	return c.logger.GetLevel()
}