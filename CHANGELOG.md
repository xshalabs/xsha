# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.8.0 - 20250916

### Changed

- Added task title inline editing functionality in kanban board
- Improved kanban page UI layout (moved return button to top-left, removed logo)
- Enabled push branch button for completed and cancelled tasks in kanban board

### Fixed

- Fixed title edit button overlap with close button in TaskDetailSheet

## v0.7.0 - 20250911

### Added

- Multi-user support with role-based access control (Developer, Admin, Super Admin)

### Fixed

- Fixed git diff permission vulnerability

## v0.6.0 - 20250901

### Added

- Added HTTP protocol support for Git repositories

## v0.5.0 - 20250826

### Added

- Added plan mode feature for enhanced task planning and execution
- Optimized workspace directory mapping for host environment execution

## v0.4.0 - 20250825

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
