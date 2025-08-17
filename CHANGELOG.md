# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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