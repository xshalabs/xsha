# 🚀 XSha - Cursor Background 开源替代 | Cluade Code 远程|并发解决方案 🌟

XSha 是一款将项目管理、Git、基于 AI 驱动的需求开发结合的软件系统。现基于 Claude Code 实现了通过任务对话即可完成项目任务开发，可同时多个任务并发执行，每个任务拥有独立的工作空间。通过 Docker 容器分配每个任务的执行环境从而保证了安全性。基于 ENV 环境变量的配置，可以轻松简答的接入 Kimi2、GLM 4.5、Qwen Coder 等更有性价比的大模型。 ✨

## 🔥 核心特性

- **🧠 AI 驱动的项目任务自动开发：** 于 Claude Code 封装，开发能力上线取决于 Claude Code 的上限。同时支持 Kimi2/GLM 4.5/Qwen Coder 等模型。
- **🛡️ 执行环境隔离：** 基于 Docker 的容器运行方案，每个 Claude Code 的执行都在独立的容器内部，保证安全性。
- **⚡ 并发执行任务：** 可控制的并发数量让项目任务开发更快速。
- **🔄 Git 接入：** 直接导入 Git 仓库，项目任务开发完成后一键推送到仓库，还可以在线查看 Git Diff 。

## 🏃‍♂️ 快速开始

1. 📥 **克隆仓库**

```bash
git clone https://github.com/XShaLabs/xsha.git
cd xsha
```

2. 🚀 **启动应用程序**

```bash
sudo chmod 666 /var/run/docker.sock
docker compose -f docker-compose.cn.yml up -d
```

3. 🌍 **访问应用程序**

- **前端**：http://localhost:8080

4. 🔑 **默认凭证**

- **用户名**：xshauser
- **密码**：xshapass

## 💻 本地开发

### 📋 前置要求

- **Docker & Docker Compose**：用于容器化部署 🐳
- **Git**：用于克隆仓库和推送分支 📂
- **Go 1.21+**：用于本地开发 🐹
- **Node.js 20+**：用于前端开发 📦

### 🚀 快速上手

1. 🗄️ **后端设置**

```bash
cd backend
make deps          # 下载依赖
make dev           # 启动开发服务器
```

2. 🎨 **前端设置**

```bash
cd frontend
npm install        # 安装依赖
npm run dev        # 启动开发服务器
```

## 🤝 参与贡献

我们欢迎社区的贡献！以下是参与方式：🎉

### 🛠️ 开发设置

1. 🍴 **Fork 仓库**并克隆您的 fork
2. 🌿 **创建功能分支**：`git checkout -b feature/amazing-feature`

### 📝 Pull Request 流程

1. ✅ **确保测试通过**并保持覆盖率
2. 📚 **更新文档**以反映任何 API 变更
3. 📋 **遵循 PR 模板**并提供清晰的描述
4. 👀 **请求维护者审查**

### 🐛 问题和错误报告

- 📄 使用提供的问题模板
- 🔍 包含重现步骤和环境详细信息
- 🏷️ 适当标记问题（bug、enhancement、question 等）

## 📄 开源协议

本仓库采用 [XSHA 开源许可证](LICENSE) 授权，基于 Apache 2.0 并附加额外条件。⚖️

---

**由 XSHA 团队用 ❤️ 构建** 👨‍💻👩‍💻
