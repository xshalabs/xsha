# Multi-stage build Dockerfile
# Stage 1: Frontend build environment
FROM node:20-alpine AS frontend-builder

# Install pnpm
RUN npm install -g pnpm

# Set working directory for frontend
WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install frontend dependencies
RUN pnpm install --frozen-lockfile

# Copy frontend source code
COPY frontend/ .

# Build frontend application (outputs to ../backend/static)
RUN pnpm run build

# Stage 2: Backend build environment
FROM golang:1.23.1-alpine AS backend-builder

# Multi-platform build arguments
ARG TARGETOS
ARG TARGETARCH

# Set working directory
WORKDIR /app

# Install necessary tools including build dependencies for CGO
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# Set timezone
ENV TZ=Asia/Shanghai

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy backend source code
COPY backend/ .

# Copy frontend build output from previous stage
COPY --from=frontend-builder /app/backend/static ./static

# Build application with CGO enabled for SQLite support, supporting multi-platform
RUN CGO_ENABLED=1 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -ldflags="-w -s" -o main .

# Stage 3: Runtime environment
FROM alpine:latest

# Install necessary packages including Docker CLI and SQLite runtime
RUN apk --no-cache add ca-certificates tzdata curl docker-cli sqlite

# Set timezone
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Create non-root user and add to docker group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    addgroup appuser docker

# Set working directory
WORKDIR /app

# Copy executable from backend builder stage
COPY --from=backend-builder /app/main .

# Copy internationalization files
COPY --from=backend-builder /app/i18n/locales ./i18n/locales/

# Copy frontend static files
COPY --from=backend-builder /app/static ./static/

# Set file permissions
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Start application
CMD ["./main"] 