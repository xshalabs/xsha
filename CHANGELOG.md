# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.4.0

### Added

- Add model selection to task conversation
- Added scheduled label and execution time to the task conversation list
- Added image support for chat conversations
- Added PDF support for chat conversations
- Refactored backend logging system

### Fixed

- Fixed task conversation status abnormality caused by container execution input logs exceeding 64KB per line
- Optimized task conversation retry mechanism - only the latest failed or cancelled task conversation can be retried under the new mechanism

## v0.3.0 - 2025-08-20

### Added

- Added kanban page
- Added real-time log viewing

## v0.2.0 - 2025-08-15

### Added

- Scheduled task execution system for automated project workflows
- System prompt configuration for projects to customize AI behavior
- System prompt configuration for development environments
- Enhanced timezone handling for better international support

### Changed

- Complete UI reconstruction with improved user experience
- Refactored date and timezone handling across the application

### Fixed

- Database initialization bug that caused duplicate initialization processes

## v0.1.0 - 2025-08-05

### Added

- Initial release of XSHA AI-powered project task automation platform
- Core project management functionality
- Basic task execution capabilities
- User authentication and authorization system
- RESTful API with comprehensive documentation
