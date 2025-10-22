# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.12.0 - 20251022

- Added large language model provider configuration support
- Removed the restriction that plan mode must use the Opus model
- Fixed page title display error bug
- Fixed project edit repo_url field editing bug

## v0.11.0 - 20251012

- Removed the mandatory requirement to use Opus model in plan mode, allowing users to select their preferred model

## v0.10.0 - 20251011

- MCP (Model Context Protocol) support for enhanced AI integration
- Optimized container directory mapping for improved workspace management
- Upgraded Claude Code to V2.X

## v0.9.0 - 20250917

- Email notification system for task completion alerts
- Notification system supporting WeChat Work, DingTalk, Feishu, Slack, Discord, and Webhook integrations
- Refactored system configuration logic with optimized multi-language environment display

## v0.8.0 - 20250916

- Added task title inline editing functionality in kanban board
- Improved kanban page UI layout (moved return button to top-left, removed logo)
- Enabled push branch button for completed and cancelled tasks in kanban board
- Hide new message input form for cancelled tasks in task detail panel
- Fixed title edit button overlap with close button in TaskDetailSheet

## v0.7.0 - 20250911

- Multi-user support with role-based access control (Developer, Admin, Super Admin)
- Fixed git diff permission vulnerability

## v0.6.0 - 20250901

- Added HTTP protocol support for Git repositories

## v0.5.0 - 20250826

- Added plan mode feature for enhanced task planning and execution
- Optimized workspace directory mapping for host environment execution

## v0.4.0 - 20250825

- Add model selection to task conversation
- Added scheduled label and execution time to the task conversation list
- Added image support for chat conversations
- Added PDF support for chat conversations
- Refactored backend logging system
- Fixed task conversation status abnormality caused by container execution input logs exceeding 64KB per line
- Optimized task conversation retry mechanism - only the latest failed or cancelled task conversation can be retried under the new mechanism

## v0.3.0 - 2025-08-20

- Added kanban page
- Added real-time log viewing

## v0.2.0 - 2025-08-15

- Scheduled task execution system for automated project workflows
- System prompt configuration for projects to customize AI behavior
- System prompt configuration for development environments
- Enhanced timezone handling for better international support
- Complete UI reconstruction with improved user experience
- Refactored date and timezone handling across the application
- Database initialization bug that caused duplicate initialization processes

## v0.1.0 - 2025-08-05

- Initial release of XSHA AI-powered project task automation platform
- Core project management functionality
- Basic task execution capabilities
- User authentication and authorization system
- RESTful API with comprehensive documentation
