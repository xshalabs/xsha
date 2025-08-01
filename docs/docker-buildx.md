# Multi-Platform Docker Build Guide

This document explains how to build Docker images for multiple architectures.

## Supported Platforms

- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, Apple Silicon, AWS Graviton)
- `linux/arm/v7` (ARM 32-bit, Raspberry Pi)

## Prerequisites

1. **Docker Buildx**: Ensure Docker Buildx is installed and enabled
   ```bash
   docker buildx version
   ```

2. **Setup Buildx Builder**: Create a multi-platform builder instance
   ```bash
   make docker-setup-buildx
   ```

## Build Commands

### 1. Single Platform Builds

```bash
# Build for current platform (default)
make docker-build

# Build specifically for AMD64
make docker-build-amd64

# Build specifically for ARM64
make docker-build-arm64
```

### 2. Multi-Platform Build

```bash
# Build and push to registry (requires Docker Hub or other registry)
make docker-build-multiplatform
```

**Note**: Multi-platform builds require pushing to a registry. Make sure you're logged in:
```bash
docker login
```

### 3. Local Multi-Platform Testing

To build for multiple platforms locally without pushing:

```bash
# Build for multiple platforms and load to local Docker
docker buildx build --platform linux/amd64,linux/arm64 -t xsha-backend:latest .
```

## Environment Variables

You can override build settings:

```bash
# Custom image name
DOCKER_IMAGE=myregistry/xsha-backend:v1.0.0 make docker-build-multiplatform

# Build with custom platforms
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t xsha-backend:latest --push .
```

## Docker Compose with Multi-Platform

Update your `docker-compose.yml` to specify platform:

```yaml
services:
  app:
    image: xsha-backend:latest
    platform: linux/amd64  # or linux/arm64
    # ... rest of config
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build Multi-Platform Docker Image

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
        
      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
          
      - name: Build and push multi-platform image
        run: |
          make docker-setup-buildx
          make docker-build-multiplatform
```

## Troubleshooting

### 1. Builder Not Found
```bash
# Remove existing builder and recreate
docker buildx rm multiplatform-builder
make docker-setup-buildx
```

### 2. Platform Not Supported
Some base images might not support all platforms. Check:
```bash
docker manifest inspect node:20-alpine
docker manifest inspect golang:1.23.1-alpine
```

### 3. Local Registry Testing
For local testing with registry:
```bash
# Start local registry
docker run -d -p 5000:5000 --name registry registry:2

# Build and push to local registry
docker buildx build --platform linux/amd64,linux/arm64 -t localhost:5000/xsha-backend:latest --push .
```

## Performance Tips

1. **Use Build Cache**: Enable BuildKit cache for faster builds
   ```bash
   export DOCKER_BUILDKIT=1
   ```

2. **Parallel Builds**: BuildX automatically builds platforms in parallel

3. **Layer Optimization**: Multi-stage builds reduce final image size across all platforms

## Verification

Check built images:
```bash
# List images
docker images | grep xsha-backend

# Inspect multi-platform manifest
docker manifest inspect xsha-backend:latest
```