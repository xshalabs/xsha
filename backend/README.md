# Sleep0 Backend

基于 Golang + Gin 框架的后端项目，采用清洁架构设计，支持 SQLite 和 MySQL 数据库，使用 JWT 进行用户认证。项目实现了完整的项目管理、Git 凭据管理、开发环境管理、任务管理，以及基于定时器的 AI 自动化编程任务执行系统。

## 🚀 主要功能特性

- **用户认证**: JWT token 认证，支持登录日志记录和 token 黑名单
- **项目管理**: Git 仓库项目管理，支持多种协议和认证方式
- **凭据管理**: 支持密码、Token、SSH Key 等多种 Git 认证方式，敏感信息 AES 加密存储
- **开发环境**: Docker 容器化开发环境管理，支持资源限制和环境变量配置
- **任务系统**: 项目任务管理和对话式交互
- **AI 自动化**: 定时器驱动的 AI 任务执行系统，支持代码自动生成和提交
- **国际化**: 多语言支持（中文/英文）
- **操作日志**: 完整的管理员操作审计日志
- **API 文档**: 完整的 Swagger API 文档

## 📁 项目架构

```
backend/
├── main.go                    # 主程序入口
├── config/                    # 配置管理
│   └── config.go             # 应用配置和环境变量
├── database/                  # 数据库层
│   ├── database.go           # 数据库连接管理
│   └── models.go             # 数据模型定义
├── repository/                # 数据访问层（Repository Pattern）
│   ├── interfaces.go         # 仓储接口定义
│   ├── admin_operation_log.go # 管理员操作日志仓储
│   ├── dev_environment.go    # 开发环境仓储
│   ├── git_credential.go     # Git凭据仓储
│   ├── login_log.go          # 登录日志仓储
│   ├── project.go            # 项目仓储
│   ├── task.go               # 任务仓储
│   ├── task_conversation.go  # 任务对话仓储
│   ├── task_execution_log.go # 任务执行日志仓储
│   └── token_blacklist.go    # Token黑名单仓储
├── services/                  # 业务逻辑层（Service Layer）
│   ├── interfaces.go         # 服务接口定义
│   ├── admin_operation_log.go # 操作日志服务
│   ├── ai_task_executor.go   # AI任务执行服务
│   ├── auth.go               # 认证服务
│   ├── dev_environment.go    # 开发环境服务
│   ├── git_credential.go     # Git凭据服务
│   ├── login_log.go          # 登录日志服务
│   ├── project.go            # 项目服务
│   ├── task.go               # 任务服务
│   └── task_conversation.go  # 任务对话服务
├── handlers/                  # HTTP 请求处理层
│   ├── admin_operation_log.go # 操作日志处理器
│   ├── auth.go               # 认证处理器
│   ├── dev_environment.go    # 开发环境处理器
│   ├── git_credential.go     # Git凭据处理器
│   ├── health.go             # 健康检查处理器
│   ├── i18n.go               # 国际化处理器
│   ├── project.go            # 项目处理器
│   ├── task.go               # 任务处理器
│   ├── task_conversation.go  # 任务对话处理器
│   └── task_execution_log.go # 任务执行日志处理器
├── middleware/                # 中间件
│   ├── auth.go               # JWT 认证中间件
│   ├── error.go              # 错误处理中间件
│   ├── i18n.go               # 国际化中间件
│   ├── logger.go             # 日志中间件
│   ├── operation_log.go      # 操作日志中间件
│   └── ratelimit.go          # 速率限制中间件
├── routes/                    # 路由配置
│   └── routes.go             # 路由注册和分组
├── scheduler/                 # 定时器模块 🆕
│   ├── interfaces.go         # 定时器接口定义
│   ├── manager.go            # 定时器管理器
│   └── task_processor.go     # 任务处理器
├── utils/                     # 工具函数
│   ├── crypto.go             # AES 加密工具
│   ├── git.go                # Git 操作工具（支持SSL配置）🆕
│   ├── jwt.go                # JWT 工具
│   └── workspace.go          # 工作目录管理工具 🆕
├── i18n/                      # 国际化模块
│   ├── helper.go             # 国际化助手
│   ├── i18n.go               # 国际化核心
│   └── locales/              # 语言文件
│       ├── en-US.json        # 英文语言包
│       └── zh-CN.json        # 中文语言包
├── cmd/                       # 命令行工具
│   ├── cleanup/              # 清理工具
│   └── encrypt-password/     # 密码加密工具
├── docs/                      # API 文档（自动生成）
├── go.mod                     # Go 模块文件
├── go.sum                     # 依赖版本锁定
└── README.md                  # 项目说明
```

## 🗃️ 数据库模型

### 核心实体
- **TokenBlacklist**: JWT token 黑名单
- **LoginLog**: 用户登录日志
- **AdminOperationLog**: 管理员操作审计日志
- **GitCredential**: Git 认证凭据（加密存储）
- **Project**: Git 项目配置
- **DevEnvironment**: 开发环境配置
- **Task**: 项目任务
- **TaskConversation**: 任务对话记录
- **TaskExecutionLog**: 任务执行日志 🆕

### 关键关系
```
Project 1:N Task
Task 1:N TaskConversation
TaskConversation 1:1 TaskExecutionLog
Project N:1 GitCredential
Task N:1 DevEnvironment
```

## 🚀 快速开始

### 1. 环境配置

设置环境变量（可选，如不设置将使用默认值）：

```bash
# 基础配置
export SLEEP0_PORT="8080"
export SLEEP0_ENVIRONMENT="development"

# 数据库配置
export SLEEP0_DATABASE_TYPE="sqlite"  # sqlite 或 mysql
export SLEEP0_SQLITE_PATH="app.db"
export SLEEP0_MYSQL_DSN="user:password@tcp(localhost:3306)/sleep0?charset=utf8mb4&parseTime=True&loc=Local"

# 认证配置
export SLEEP0_ADMIN_USER="admin"
export SLEEP0_ADMIN_PASS="admin123"
export SLEEP0_JWT_SECRET="your-strong-jwt-secret-key-here"
export SLEEP0_AES_KEY="your-32-byte-aes-encryption-key-here"

# Git配置 🆕
export SLEEP0_GIT_SSL_VERIFY="false"                    # Git SSL验证开关（默认禁用以解决兼容性问题）

# 定时器配置 🆕
export SLEEP0_SCHEDULER_INTERVAL="30s"              # 定时器扫描间隔
export SLEEP0_WORKSPACE_BASE_DIR="/tmp/sleep0-workspaces"  # AI任务工作目录
export SLEEP0_DOCKER_TIMEOUT="30m"                  # Docker执行超时时间
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 运行项目

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动，并自动启动定时器服务。

### 4. API 文档

启动后访问 Swagger API 文档：
- **Swagger UI**: http://localhost:8080/swagger/index.html

## 📚 API 接口

### 认证管理
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/user/current` - 获取当前用户信息

### Git 凭据管理
- `POST /api/v1/git-credentials` - 创建 Git 凭据
- `GET /api/v1/git-credentials` - 获取凭据列表
- `GET /api/v1/git-credentials/:id` - 获取单个凭据
- `PUT /api/v1/git-credentials/:id` - 更新凭据
- `DELETE /api/v1/git-credentials/:id` - 删除凭据

### 项目管理
- `POST /api/v1/projects` - 创建项目
- `GET /api/v1/projects` - 获取项目列表
- `GET /api/v1/projects/:id` - 获取单个项目
- `PUT /api/v1/projects/:id` - 更新项目
- `DELETE /api/v1/projects/:id` - 删除项目
- `GET /api/v1/projects/:id/branches` - 获取项目分支列表

### 开发环境管理
- `POST /api/v1/dev-environments` - 创建开发环境
- `GET /api/v1/dev-environments` - 获取环境列表
- `GET /api/v1/dev-environments/:id` - 获取单个环境
- `PUT /api/v1/dev-environments/:id` - 更新环境
- `DELETE /api/v1/dev-environments/:id` - 删除环境
- `POST /api/v1/dev-environments/:id/control` - 控制环境（启动/停止/重启）

### 任务管理
- `POST /api/v1/tasks` - 创建任务
- `GET /api/v1/tasks` - 获取任务列表
- `GET /api/v1/tasks/:id` - 获取单个任务
- `PUT /api/v1/tasks/:id` - 更新任务
- `DELETE /api/v1/tasks/:id` - 删除任务

### 任务对话管理
- `POST /api/v1/conversations` - 创建对话
- `GET /api/v1/conversations` - 获取对话列表
- `GET /api/v1/conversations/:id` - 获取单个对话
- `PUT /api/v1/conversations/:id` - 更新对话
- `DELETE /api/v1/conversations/:id` - 删除对话

### AI 任务执行 🆕
- `GET /api/v1/task-conversations/:conversationId/execution-log` - 获取执行日志
- `POST /api/v1/task-conversations/:conversationId/execution/cancel` - 取消任务执行

### SSE 实时日志 🆕
- `GET /api/v1/logs/stream` - SSE实时日志流
- `GET /api/v1/logs/stats` - 获取连接统计信息
- `POST /api/v1/logs/test/:conversationId` - 发送测试消息

### 管理功能
- `GET /api/v1/admin/operation-logs` - 获取操作日志
- `GET /api/v1/admin/login-logs` - 获取登录日志

### 国际化
- `GET /api/v1/languages` - 获取支持的语言列表
- `POST /api/v1/language` - 设置语言

### 健康检查
- `GET /health` - 服务健康检查

## 🔧 Git 配置与故障排除 🆕

### Git SSL 验证配置

项目支持灵活的 Git SSL 验证配置，以解决不同 Git 服务器的兼容性问题。

#### 环境变量配置
```bash
# 禁用 SSL 验证（推荐用于解决 TLS 连接问题）
export SLEEP0_GIT_SSL_VERIFY=false

# 启用 SSL 验证（推荐用于生产环境）
export SLEEP0_GIT_SSL_VERIFY=true
```

#### 配置说明
- **默认值**: `false` - 禁用 SSL 验证，解决常见的 TLS 连接问题
- **推荐设置**: 
  - 开发环境：`false` - 避免证书问题
  - 生产环境：`true` - 保证安全性

#### 常见问题解决

**问题 1: TLS 连接错误**
```
错误: fatal: unable to access 'https://gitee.com/xxx/xxx.git/': 
TLS connect error: error:0A000126:SSL routines::unexpected eof while reading
```

**解决方案:**
```bash
export SLEEP0_GIT_SSL_VERIFY=false
# 重启应用
```

**问题 2: SSL 证书验证失败**
```
错误: SSL certificate problem: unable to get local issuer certificate
```

**解决方案:**
```bash
# 临时禁用 SSL 验证
export SLEEP0_GIT_SSL_VERIFY=false

# 或者更新系统证书（推荐）
# macOS:
brew install ca-certificates
# Ubuntu/Debian:
sudo apt-get update && sudo apt-get install ca-certificates
```

#### 支持的 Git 服务器
- ✅ **GitHub**: 支持 SSL 验证
- ✅ **GitLab**: 支持 SSL 验证  
- ⚠️ **Gitee**: 建议禁用 SSL 验证（`SLEEP0_GIT_SSL_VERIFY=false`）
- ✅ **Bitbucket**: 支持 SSL 验证
- ⚠️ **自建 Git 服务器**: 根据证书配置情况选择

#### 错误处理增强
当启用 SSL 验证且遇到相关错误时，系统会自动提供解决建议：
```
仓库访问验证失败: [具体错误信息]
建议: 可尝试设置环境变量 SLEEP0_GIT_SSL_VERIFY=false 禁用SSL验证
```

#### 安全性考虑
- **开发环境**: 可以安全地禁用 SSL 验证以提高兼容性
- **生产环境**: 建议启用 SSL 验证，确保通信安全
- **内网环境**: 对于内网 Git 服务器，可根据实际情况选择
- **代理环境**: 使用代理时可能需要禁用 SSL 验证

## 🤖 AI 自动化功能

### 定时器系统
- **自动扫描**: 每 30 秒扫描待处理的任务对话
- **并发执行**: 支持多个任务同时执行，默认最大并发数为5
- **智能调度**: 自动管理并发数量，避免资源过载
- **实时日志**: 执行过程实时记录到数据库并通过SSE推送
- **优雅关闭**: 支持优雅停止，确保任务完成

### 并发执行管理 🆕
- **执行管理器**: 统一管理所有并发任务的生命周期
- **资源控制**: 通过配置限制最大并发数（`SLEEP0_MAX_CONCURRENT_TASKS`）
- **状态跟踪**: 实时跟踪每个任务的执行状态和进度
- **强制取消**: 支持强制取消正在运行的任务

### 实时日志系统 🆕
- **SSE推送**: 使用Server-Sent Events技术实时推送执行日志
- **日志广播**: 支持多客户端同时接收实时日志
- **自动重连**: 客户端断线后自动重连
- **消息过滤**: 支持按对话ID过滤特定任务的日志
- **连接管理**: 自动清理非活跃连接，优化资源使用

### 任务执行流程
1. **扫描**: 定时器扫描 `pending` 状态的 TaskConversation
2. **准备**: 获取或创建任务级工作空间，克隆/拉取代码仓库
3. **执行**: 根据开发环境配置构建并执行 Docker 命令
4. **记录**: 实时记录执行日志和状态变化
5. **提交**: 成功执行后自动提交代码更改
6. **持久化**: 保留任务工作空间供后续对话使用，任务完成时自动清理

### 支持的开发环境
- **Claude Code**: Claude AI 编程环境
- **Gemini CLI**: Google Gemini 命令行工具
- **OpenCode**: 开源代码生成工具

## 🔧 环境变量完整列表

| 变量名 | 描述 | 默认值 | 类型 |
|--------|------|--------|------|
| `SLEEP0_PORT` | 服务器端口 | 8080 | string |
| `SLEEP0_ENVIRONMENT` | 运行环境 | development | string |
| `SLEEP0_DATABASE_TYPE` | 数据库类型 | sqlite | string |
| `SLEEP0_SQLITE_PATH` | SQLite 数据库文件路径 | app.db | string |
| `SLEEP0_MYSQL_DSN` | MySQL 数据库连接字符串 | - | string |
| `SLEEP0_ADMIN_USER` | 管理员用户名 | admin | string |
| `SLEEP0_ADMIN_PASS` | 管理员密码 | admin123 | string |
| `SLEEP0_JWT_SECRET` | JWT 密钥 | your-jwt-secret-key-change-this-in-production | string |
| `SLEEP0_AES_KEY` | AES 加密密钥 | default-aes-key-change-in-production | string |
| `SLEEP0_SCHEDULER_INTERVAL` | 定时器间隔 🆕 | 30s | duration |
| `SLEEP0_WORKSPACE_BASE_DIR` | 工作目录基础路径 🆕 | /tmp/sleep0-workspaces | string |
| `SLEEP0_DOCKER_TIMEOUT` | Docker 执行超时时间 🆕 | 30m | duration |
| `SLEEP0_MAX_CONCURRENT_TASKS` | 最大并发任务数 🆕 | 5 | int |
| `SLEEP0_GIT_SSL_VERIFY` | Git SSL验证开关 🆕 | false | boolean |

## 🏗️ 架构设计

### 清洁架构分层
```
┌─────────────────────────────────────────┐
│             Handlers Layer              │  HTTP 请求处理
├─────────────────────────────────────────┤
│             Services Layer              │  业务逻辑处理
├─────────────────────────────────────────┤
│            Repository Layer             │  数据访问抽象
├─────────────────────────────────────────┤
│             Database Layer              │  数据持久化
└─────────────────────────────────────────┘

           ┌─────────────────┐
           │  Scheduler      │  定时器模块
           │  - Manager      │
           │  - Processor    │
           └─────────────────┘
```

### 设计原则
- **依赖注入**: 通过接口解耦各层依赖
- **单一职责**: 每个模块职责明确
- **开闭原则**: 对扩展开放，对修改关闭
- **接口隔离**: 最小化接口依赖
- **配置外部化**: 所有配置通过环境变量管理

## 🔒 安全特性

- **JWT 认证**: 无状态 token 认证
- **Token 黑名单**: 支持 token 撤销
- **AES 加密**: 敏感信息加密存储
- **速率限制**: 登录接口防暴力破解
- **操作审计**: 完整的操作日志记录
- **输入验证**: 所有输入参数验证
- **错误隐藏**: 生产环境隐藏敏感错误信息

## 📦 主要依赖

### 核心框架
- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - JWT 认证

### 数据库驱动
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite 驱动
- [mysql](https://github.com/go-sql-driver/mysql) - MySQL 驱动

### 工具库
- [gin-swagger](https://github.com/swaggo/gin-swagger) - API 文档生成
- [validator](https://github.com/go-playground/validator) - 数据验证
- [testify](https://github.com/stretchr/testify) - 测试框架

## 🚀 部署指南

### Docker 部署
```bash
# 构建镜像
docker build -t sleep0-backend .

# 运行容器
docker run -d \
  --name sleep0-backend \
  -p 8080:8080 \
  -e SLEEP0_ENVIRONMENT=production \
  -e SLEEP0_JWT_SECRET=your-production-secret \
  -e SLEEP0_AES_KEY=your-production-aes-key \
  -e SLEEP0_GIT_SSL_VERIFY=true \
  -v /data/sleep0:/data \
  sleep0-backend
```

### 生产环境建议
1. **使用 MySQL**: 生产环境建议使用 MySQL 数据库
2. **强密钥**: 使用强随机密钥作为 JWT 和 AES 密钥
3. **HTTPS**: 配置 HTTPS 传输加密
4. **反向代理**: 使用 Nginx 作为反向代理
5. **Git SSL 验证**: 生产环境启用 SSL 验证 (`SLEEP0_GIT_SSL_VERIFY=true`) 🆕
6. **监控**: 配置应用监控和日志收集
7. **备份**: 定期备份数据库

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./services/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

该项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 🔗 相关链接

- [API 文档](http://localhost:8080/swagger/index.html)
- [Gin 框架文档](https://gin-gonic.com/)
- [GORM 文档](https://gorm.io/docs/)
- [Docker 部署指南](./docs/DOCKER.md)

---

**Sleep0 Backend** - 构建智能化的开发工作流 🚀 