# XSHA

[Chinese🇨🇳](README_CN.md)

XSHA is an AI-powered (currently supporting `Claude Code`) project task automation development platform.

## Key Features

### 🤖 **AI-Powered Task Automation**

- **Intelligent Task Execution**: AI-driven task processing and automation
- **Conversation-based Workflows**: Natural language task descriptions and execution
- **Task Scheduling**: Automated task execution with flexible scheduling options

### 🎯 **Project Management**

- **Multi-project Support**: Manage multiple projects with individual configurations and settings
- **Project Templates**: Quick project setup with predefined templates and configurations
- **Project Analytics**: Track project metrics and development progress

### 🔐 **Git Credentials Management**

- **Secure Credential Storage**: Encrypted storage of Git credentials with role-based access
- **Multiple Provider Support**: Support for GitHub, GitLab, Bitbucket, and custom Git servers
- **Credential Sharing**: Team-based credential sharing with granular permissions

### 🚀 **Development Environment Orchestration**

- **Environment Templates**: Predefined development environment configurations
- **Docker Integration**: Containerized development environments for consistency
- **Environment Provisioning**: Automated setup and teardown of development environments

### 📊 **Admin & Monitoring**

- **Operation Logging**: Comprehensive audit trails for all system operations
- **User Management**: Role-based access control and user administration
- **System Configuration**: Flexible system-wide configuration management

### 🌐 **Modern Tech Stack**

- **Backend**: Go with Gin framework, GORM ORM, JWT authentication
- **Frontend**: React 18+, TypeScript, Vite, shadcn/ui, Tailwind CSS
- **Database**: MySQL with automated migrations
- **Deployment**: Docker & Docker Compose with health checks
- **API Documentation**: OpenAPI/Swagger integration

## Quick Start

1. **Clone the repository**

```bash
git clone https://github.com/XshaLabs/xsha.git
cd xsha
```

2. **Start the application**

```bash
docker-compose up -d
```

3. **Access the application**

- **Frontend**: http://localhost:8080

4. **Default credentials**

- **Username**: xshauser
- **Password**: xshapass

## Local Development

### Prerequisites

- **Docker & Docker Compose**: For containerized deployment
- **Go 1.21+**: For local development (optional)
- **Node.js 18+**: For frontend development (optional)
- **Make**: For build automation (optional)

### Getting Started

1. **Backend setup**

```bash
cd backend
make deps          # Download dependencies
make dev           # Start development server
```

2. **Frontend setup**

```bash
cd frontend
npm install        # Install dependencies
npm run dev        # Start development server
```

## Contributing

We welcome contributions from the community! Here's how you can get involved:

### Development Setup

1. **Fork the repository** and clone your fork
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Follow the coding standards**:
   - Go: Follow standard Go conventions and run `make check`
   - TypeScript: Use ESLint and Prettier configurations
   - Commit messages: Use conventional commit format

### Pull Request Process

1. **Ensure tests pass** and coverage is maintained
2. **Update documentation** for any API changes
3. **Follow the PR template** and provide clear descriptions
4. **Request review** from maintainers

### Issues and Bug Reports

- Use the issue templates provided
- Include reproduction steps and environment details
- Label issues appropriately (bug, enhancement, question, etc.)

## License

This repository is licensed under the [XSHA Open Source License](LICENSE), based on Apache 2.0 with additional conditions.

---

**Built with ❤️ by the XSHA team**
