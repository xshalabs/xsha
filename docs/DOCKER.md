# Docker Usage Guide

This project has been configured with complete Docker support, including configurations for both production and development environments.

## Quick Start

### 1. Using Makefile (Recommended)

```bash
# View all available commands
make help

# Setup development environment
make setup

# Build Docker image
make docker-build

# Start application (single container)
make docker-run

# Start complete services (including database)
make docker-compose-up

# View logs
make docker-compose-logs

# Stop services
make docker-compose-down
```

### 2. Using Docker Commands Directly

```bash
# Build image
docker build -t sleep0-backend:latest .

# Run container
docker run --rm -p 8080:8080 sleep0-backend:latest
```

### 3. Using Docker Compose

#### Production Environment
```bash
# Start all services (app + MySQL + phpMyAdmin)
docker-compose up -d

# Start only app and database
docker-compose up -d app mysql

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

#### Development Environment
```bash
# Start development environment (SQLite + Redis + hot reload)
docker-compose -f docker-compose.dev.yml up -d

# Start development environment with MySQL
docker-compose -f docker-compose.dev.yml --profile mysql up -d

# View development environment logs
docker-compose -f docker-compose.dev.yml logs -f app
```

## Configuration

### Environment Variables

Create a `.env` file (based on `.env.example`) to configure the application:

```bash
cp .env.example .env
# Then edit the .env file
```

### Database Selection

- **SQLite** (default): Suitable for development and small deployments
- **MySQL**: Suitable for production environments

### Service Ports

- **8080**: Main application port
- **3306**: MySQL database
- **8081**: phpMyAdmin (optional)
- **6379**: Redis (development environment)

## File Descriptions

- `Dockerfile`: Production environment image build
- `Dockerfile.dev`: Development environment image build (with hot reload)
- `docker-compose.yml`: Production environment service orchestration
- `docker-compose.dev.yml`: Development environment service orchestration
- `.dockerignore`: Docker build ignore file
- `scripts/init.sql`: MySQL initialization script

## Development Workflow

### Local Development
```bash
# 1. Setup environment
make setup

# 2. Start development server
make dev

# Or use Docker development environment
make docker-compose -f docker-compose.dev.yml up -d
```

### Build and Test
```bash
# Code check
make check

# Run tests
make test

# Build application
make build

# Build Docker image
make docker-build
```

### Production Deployment
```bash
# 1. Build production image
make docker-build

# 2. Start production services
make docker-compose-up

# 3. Check health status
make health
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check port usage
   lsof -i :8080
   
   # Modify port (in .env or docker-compose.yml)
   SLEEP0_PORT=8081
   ```

2. **Permission Issues**
   ```bash
   # Fix file permissions
   sudo chown -R $USER:$USER .
   ```

3. **Database Connection Failure**
   ```bash
   # Check database service status
   docker-compose ps mysql
   
   # View database logs
   docker-compose logs mysql
   ```

4. **Clean Environment**
   ```bash
   # Clean Docker resources
   make docker-clean
   
   # Reset database
   make db-reset
   ```

### Log Viewing

```bash
# Application logs
docker-compose logs -f app

# Database logs
docker-compose logs -f mysql

# All service logs
docker-compose logs -f
```

### Database Management

- **phpMyAdmin**: http://localhost:8081 (Username: root, Password: password)
- **Direct MySQL Connection**:
  ```bash
  docker-compose exec mysql mysql -u root -p sleep0
  ```

## Performance Optimization

### Image Optimization
- Use multi-stage builds to reduce image size
- Use Alpine Linux base images
- Enable Go compiler optimizations (`-ldflags="-w -s"`)

### Runtime Optimization
- Run with non-root user
- Configure health checks
- Set resource limits (can be added in docker-compose.yml)

```yaml
services:
  app:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
``` 