# Docker in Docker (DinD) 解决方案

## 问题描述

当应用运行在 Docker 容器内时，需要执行其他 Docker 命令来启动 AI 工具容器。这就是典型的 "Docker in Docker" (DinD) 场景。

虽然可以访问 `/var/run/docker.sock`，但容器内没有 Docker CLI 工具，导致无法执行 Docker 命令。

## 解决方案

### 1. 安装 Docker CLI

在 Dockerfile 中安装 Docker CLI：

```dockerfile
# 在 runtime 阶段安装 Docker CLI
RUN apk --no-cache add docker-cli

# 将用户添加到 docker 组
RUN addgroup appuser docker
```

### 2. 挂载 Docker Socket

在 `docker-compose.yml` 中挂载 Docker socket：

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock
  - ./workspaces:/app/workspaces
```

### 3. 配置工作空间

设置环境变量指定工作空间目录：

```yaml
environment:
  - SLEEP0_WORKSPACE_BASE_DIR=/app/workspaces
```

## 使用步骤

### 1. 构建 AI 工具镜像

首先构建所需的 AI 工具 Docker 镜像：

```bash
./scripts/build-ai-images.sh
```

这将构建以下镜像：
- `claude-code:latest`
- `opencode:latest`
- `gemini-cli:latest`

### 2. 启动应用

使用 Docker Compose 启动应用：

```bash
# 生产环境
docker-compose up -d

# 开发环境
docker-compose -f docker-compose.dev.yml up -d
```

### 3. 验证 Docker 可用性

应用启动后，AI 任务执行器会自动检查 Docker 是否可用：

- ✅ 如果 Docker 可用，任务将正常执行
- ❌ 如果 Docker 不可用，将记录错误信息

## 安全考虑

### 权限管理

- 应用以非 root 用户运行
- 用户被添加到 docker 组以获得 Docker 访问权限
- 通过资源限制控制容器使用

### 网络隔离

- AI 任务在独立的容器中执行
- 使用 `--rm` 参数确保容器执行后自动清理
- 通过卷挂载共享代码，避免网络传输

## 故障排除

### 1. Docker 命令不可用

错误信息：`Docker 命令不可用或 Docker 守护进程未运行`

解决方案：
- 确保容器内安装了 Docker CLI
- 检查 `/var/run/docker.sock` 是否正确挂载
- 验证用户是否在 docker 组中

### 2. 权限拒绝

错误信息：`permission denied while trying to connect to the Docker daemon socket`

解决方案：
- 检查 Docker socket 的权限
- 确保应用用户在 docker 组中
- 在 macOS/Windows 上，可能需要调整 Docker Desktop 设置

### 3. 工作空间权限问题

错误信息：`无法创建工作目录` 或 `无法写入文件`

解决方案：
- 确保工作空间目录存在且有正确权限
- 检查卷挂载配置
- 验证容器内外的用户 ID 映射

## 监控和日志

### 执行日志

AI 任务执行器提供详细的执行日志：

- Docker 可用性检查结果
- 工作空间创建和清理过程
- Docker 命令执行的实时输出
- 错误信息和堆栈跟踪

### 性能监控

- CPU 和内存使用限制
- 执行超时控制（默认 30 分钟）
- 并发任务数量控制

## 最佳实践

1. **定期清理**: 虽然使用 `--rm` 参数，但建议定期清理不必要的镜像和卷
2. **资源限制**: 为每个 AI 任务设置合适的 CPU 和内存限制
3. **超时控制**: 设置合理的任务执行超时时间
4. **日志管理**: 定期清理或轮转执行日志
5. **安全更新**: 定期更新 Docker 镜像和依赖项 