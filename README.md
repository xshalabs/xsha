# XSHA

XSHA is a modern full-stack application development platform that integrates project management, task assignment, AI self-driven development, and Git code management.

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

### Prerequisites

- **Docker & Docker Compose**: For containerized deployment
- **Go 1.21+**: For local development (optional)
- **Node.js 18+**: For frontend development (optional)
- **Make**: For build automation (optional)

### Using Docker Compose (Recommended)

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
- **Database Admin**: http://localhost:8081 (phpMyAdmin, optional)

4. **Default credentials**

- **Username**: admin
- **Password**: admin123

### Local Development

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

3. **Database setup**

```bash
docker-compose up mysql -d  # Start MySQL only
make db-reset               # Reset database (if needed)
```

### Environment Configuration

Copy and customize the environment variables:

```bash
# Backend configuration
export XSHA_PORT=8080
export XSHA_ENVIRONMENT=development
export XSHA_DATABASE_TYPE=mysql
export XSHA_MYSQL_DSN="root:password@tcp(localhost:3306)/xsha?charset=utf8mb4&parseTime=True&loc=Local"
export XSHA_JWT_SECRET="your-secret-key"
export XSHA_WORKSPACE_BASE_DIR="/tmp/workspaces"
```

### Available Commands

```bash
# Development
make dev           # Start development server
make build         # Build production binary
make test          # Run tests
make check         # Run all checks (format, vet, lint, test)

# Database
make db-reset      # Reset database

# Deployment
docker-compose up -d              # Start all services
docker-compose up -d --profile tools  # Include phpMyAdmin
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

### Code Organization

- **Backend**: Clean architecture with Repository → Service → Handler layers
- **Frontend**: Component-based architecture with hooks and contexts
- **Database**: GORM models with automated migrations
- **API**: RESTful API design with OpenAPI documentation

### Testing

```bash
# Backend tests
cd backend
make test              # Run all tests
make test-coverage     # Generate coverage report

# Frontend tests
cd frontend
npm run test          # Run frontend tests
```

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
