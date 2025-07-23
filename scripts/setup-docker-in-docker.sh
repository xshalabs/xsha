#!/bin/bash

# Setup Docker in Docker environment for Sleep0 Backend

set -e

echo "🚀 开始设置 Docker in Docker 环境..."

# 检查 Docker 是否已安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查 Docker Compose 是否已安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

echo "✅ Docker 和 Docker Compose 已安装"

# 创建工作空间目录
echo "📁 创建工作空间目录..."
mkdir -p workspaces
chmod 755 workspaces

# 构建 AI 工具镜像
echo "🔨 构建 AI 工具 Docker 镜像..."
make docker-build-ai

# 构建主应用镜像
echo "🔨 构建主应用 Docker 镜像..."
make docker-build

echo "✅ Docker in Docker 环境设置完成！"

echo ""
echo "📋 下一步操作："
echo "1. 启动开发环境: make docker-compose-up-dev"
echo "2. 启动生产环境: make docker-compose-up"
echo "3. 查看日志: make docker-compose-logs"
echo "4. 停止服务: make docker-compose-down"
echo ""
echo "📖 详细文档请查看: docs/DOCKER_IN_DOCKER.md" 