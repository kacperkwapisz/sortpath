# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2024-12-19

### Added - Configuration & Security Phase
- **Enhanced Configuration Management**: Comprehensive input validation and sanitization for all configuration values
- **Secure File Operations**: Atomic write operations with proper file permissions (600) and directory creation
- **Environment Detection**: Smart detection of CI/CD, container, and non-interactive environments
- **Input Sanitization**: Protection against directory traversal attacks and malicious input
- **Configuration Security**: API key redaction in configuration display and secure value handling

### Security Enhancements
- Path sanitization preventing directory traversal attacks (`../` sequences blocked)
- Model name validation with character restrictions (alphanumeric, hyphens, dots, underscores only)
- URL validation with scheme checking (http/https required)
- Secure file permission enforcement (600) for configuration files
- Atomic file operations preventing corruption during writes

### Environment Awareness
- CI/CD environment detection (GitHub Actions, GitLab CI, Jenkins, etc.)
- Container environment detection (Docker, Kubernetes)
- Non-interactive environment handling with appropriate fallbacks
- Environment-specific error messages and suggestions
- Smart user prompting based on environment context

### Configuration Improvements
- Log level normalization and validation (debug, info, error)
- Configuration value sanitization with whitespace trimming
- Enhanced error messages with specific remediation steps
- Graceful handling of corrupted configuration files with fallback to defaults
- Comprehensive validation for API endpoints, file paths, and model names

### Technical Enhancements
- Added `internal/config/security.go` with comprehensive input validation utilities
- Added `internal/config/environment.go` for environment detection and edge case handling
- Enhanced FileLoader with atomic write operations and secure permissions
- Comprehensive test coverage for security scenarios and edge cases
- Improved CLI configuration commands with validation and sanitization

## [0.2.0] - 2024-12-19

### Added - Foundation Phase
- **Testing Infrastructure**: Complete testing framework with unit, integration, and testdata directories
- **Centralized Error Handling**: New `internal/errors` package with context-aware error types and user-friendly messaging
- **Logging Infrastructure**: Configurable logging system with level control, sensitive data redaction, and context support
- **Test Utilities**: Comprehensive mocking framework and test helpers for filesystem, API, and configuration testing
- **Error Factory Functions**: Specialized error constructors for Config, API, Filesystem, Installation, and Validation errors

### Technical Improvements
- Added 80%+ test coverage for error handling and logging packages
- Implemented secure logging with automatic API key and sensitive data redaction
- Created extensible testing framework supporting multiple test scenarios
- Added environment variable support for log level configuration (`SORTPATH_LOG_LEVEL`, `LOG_LEVEL`)
- Implemented timed operation logging for performance monitoring

### Developer Experience
- Enhanced error messages with emoji indicators and actionable suggestions
- Added context-aware logging with operation timing
- Created comprehensive test fixtures for configuration, filesystem trees, and API responses
- Implemented mock utilities for all major components (API client, filesystem reader, logger, config, installer)

### Foundation for Production Readiness
This release establishes the foundation for the production-ready refactor by implementing:
- Robust error handling with user-friendly messages
- Comprehensive logging infrastructure with security considerations
- Complete testing framework enabling confident refactoring
- Development principles adherence (KISS, YAGNI, DRY, WET, TDD)

## [0.1.1] - Previous Release
- Initial sortpath functionality
- Basic CLI interface
- File organization recommendations via AI