# Multi-stage build Dockerfile
# Stage 1: Build environment
FROM golang:1.23.1-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Shanghai

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY backend/ .

# Build application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

# Stage 2: Runtime environment
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata curl

# Set timezone
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy executable from builder stage
COPY --from=builder /app/main .

# Copy internationalization files
COPY --from=builder /app/i18n/locales ./i18n/locales/

# Set file permissions
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1

# Start application
CMD ["./main"] 