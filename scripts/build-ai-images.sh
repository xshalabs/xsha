#!/bin/bash

# 构建AI工具Docker镜像的脚本

set -e

echo "🚀 开始构建 AI 工具 Docker 镜像..."

# 构建 Claude Code 镜像
echo "📦 构建 Claude Code 镜像..."
docker build -t claude-code:latest -f docker/Dockerfile.claude-code .

# 构建 OpenCode 镜像
echo "📦 构建 OpenCode 镜像..."
docker build -t opencode:latest -f docker/Dockerfile.opencode .

# 构建 Gemini CLI 镜像
echo "📦 构建 Gemini CLI 镜像..."
docker build -t gemini-cli:latest -f docker/Dockerfile.gemini-cli .

echo "✅ 所有 AI 工具镜像构建完成！"

# 显示构建的镜像
echo "📋 构建的镜像列表："
docker images | grep -E "(claude-code|opencode|gemini-cli)" 