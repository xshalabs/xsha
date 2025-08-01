# XSHA

XSHA 是一个项目管理、任务分配、AI 自驱动开发、Git 代码管理融合的现代化全栈应用开发平台。

## 核心特性

### 🤖 **AI 驱动的任务自动化**

- **智能任务执行**：AI 驱动的任务处理和自动化
- **对话式工作流**：自然语言任务描述和执行
- **任务调度**：具有灵活调度选项的自动化任务执行

### 🎯 **项目管理**

- **多项目支持**：管理具有独立配置和设置的多个项目
- **项目模板**：使用预定义模板和配置快速设置项目
- **项目分析**：跟踪项目指标和开发进度

### 🔐 **Git 凭证管理**

- **安全凭证存储**：基于角色访问的 Git 凭证加密存储
- **多提供商支持**：支持 GitHub、GitLab、Bitbucket 和自定义 Git 服务器
- **凭证共享**：具有细粒度权限的团队凭证共享

### 🚀 **开发环境编排**

- **环境模板**：预定义的开发环境配置
- **Docker 集成**：容器化开发环境确保一致性
- **环境配置**：开发环境的自动设置和清理

### 📊 **管理与监控**

- **操作日志**：所有系统操作的全面审计跟踪
- **用户管理**：基于角色的访问控制和用户管理
- **系统配置**：灵活的系统级配置管理

### 🌐 **现代技术栈**

- **后端**：Go + Gin 框架、GORM ORM、JWT 认证
- **前端**：React 18+、TypeScript、Vite、shadcn/ui、Tailwind CSS
- **数据库**：MySQL 和自动化迁移
- **部署**：Docker 和 Docker Compose 与健康检查
- **API 文档**：OpenAPI/Swagger 集成

## 快速开始

### 前置要求

- **Docker & Docker Compose**：用于容器化部署
- **Go 1.21+**：用于本地开发（可选）
- **Node.js 18+**：用于前端开发（可选）
- **Make**：用于构建自动化（可选）

### 使用 Docker Compose（推荐）

1. **克隆仓库**

```bash
git clone https://github.com/XShaLabs/xsha.git
cd xsha
```

2. **启动应用程序**

```bash
docker-compose up -d
```

3. **访问应用程序**

- **前端**：http://localhost:8080
- **数据库管理**：http://localhost:8081 (phpMyAdmin，可选)

4. **默认凭证**

- **用户名**：admin
- **密码**：admin123

### 本地开发

1. **后端设置**

```bash
cd backend
make deps          # 下载依赖
make dev           # 启动开发服务器
```

2. **前端设置**

```bash
cd frontend
npm install        # 安装依赖
npm run dev        # 启动开发服务器
```

3. **数据库设置**

```bash
docker-compose up mysql -d  # 仅启动 MySQL
make db-reset               # 重置数据库（如果需要）
```

### 环境配置

复制并自定义环境变量：

```bash
# 后端配置
export XSHA_PORT=8080
export XSHA_ENVIRONMENT=development
export XSHA_DATABASE_TYPE=mysql
export XSHA_MYSQL_DSN="root:password@tcp(localhost:3306)/xsha?charset=utf8mb4&parseTime=True&loc=Local"
export XSHA_JWT_SECRET="your-secret-key"
export XSHA_WORKSPACE_BASE_DIR="/tmp/workspaces"
```

### 可用命令

```bash
# 开发
make dev           # 启动开发服务器
make build         # 构建生产二进制文件
make test          # 运行测试
make check         # 运行所有检查（格式化、检查、lint、测试）

# 数据库
make db-reset      # 重置数据库

# 部署
docker-compose up -d              # 启动所有服务
docker-compose up -d --profile tools  # 包含 phpMyAdmin
```

## 参与贡献

我们欢迎社区的贡献！以下是参与方式：

### 开发设置

1. **Fork 仓库**并克隆您的 fork
2. **创建功能分支**：`git checkout -b feature/amazing-feature`
3. **遵循编码标准**：
   - Go：遵循标准 Go 约定并运行 `make check`
   - TypeScript：使用 ESLint 和 Prettier 配置
   - 提交消息：使用常规提交格式

### 代码组织

- **后端**：Repository → Service → Handler 层的清洁架构
- **前端**：基于组件的架构，使用 hooks 和 contexts
- **数据库**：GORM 模型和自动化迁移
- **API**：RESTful API 设计和 OpenAPI 文档

### 测试

```bash
# 后端测试
cd backend
make test              # 运行所有测试
make test-coverage     # 生成覆盖率报告

# 前端测试
cd frontend
npm run test          # 运行前端测试
```

### Pull Request 流程

1. **确保测试通过**并保持覆盖率
2. **更新文档**以反映任何 API 变更
3. **遵循 PR 模板**并提供清晰的描述
4. **请求维护者审查**

### 问题和错误报告

- 使用提供的问题模板
- 包含重现步骤和环境详细信息
- 适当标记问题（bug、enhancement、question 等）

## 开源协议

本仓库采用 [XSHA 开源许可证](LICENSE) 授权，基于 Apache 2.0 并附加额外条件。

---

**由 XSHA 团队用 ❤️ 构建**
