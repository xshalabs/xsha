# 🚀 XSha

XSha 是一个由 AI 驱动（目前支持 `Cluade Code` ）的项目任务自动化开发平台。✨

## ✨ 核心特性

### 🤖 **AI 驱动的任务自动化**

- **智能任务执行**：AI 驱动的任务处理和自动化 🧠
- **对话式工作流**：自然语言任务描述和执行 💬
- **任务调度**：具有灵活调度选项的自动化任务执行 ⏰

### 🎯 **项目管理**

- **多项目支持**：管理具有独立配置和设置的多个项目 📚
- **项目模板**：使用预定义模板和配置快速设置项目 📝
- **项目分析**：跟踪项目指标和开发进度 📈

### 🔐 **Git 凭证管理**

- **安全凭证存储**：基于角色访问的 Git 凭证加密存储 🔒
- **多提供商支持**：支持 GitHub、GitLab、Bitbucket 和自定义 Git 服务器 🌐
- **凭证共享**：具有细粒度权限的团队凭证共享 👥

### 🚀 **开发环境编排**

- **环境模板**：预定义的开发环境配置 📦
- **Docker 集成**：容器化开发环境确保一致性 🐳
- **环境配置**：开发环境的自动设置和清理 ⚡

### 📊 **管理与监控**

- **操作日志**：所有系统操作的全面审计跟踪 📋
- **用户管理**：基于角色的访问控制和用户管理 👨‍💼
- **系统配置**：灵活的系统级配置管理 ⚙️

### 🌐 **现代技术栈**

- **后端**：Go + Gin 框架、GORM ORM、JWT 认证 🐹
- **前端**：React 18+、TypeScript、Vite、shadcn/ui、Tailwind CSS ⚛️
- **数据库**：MySQL 和自动化迁移 🗄️
- **部署**：Docker 和 Docker Compose 与健康检查 🐳
- **API 文档**：OpenAPI/Swagger 集成 📚

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
