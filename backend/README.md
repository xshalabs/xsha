# Sleep0 Backend

基于 Golang + Gin 框架的后端项目，支持 SQLite 和 MySQL 数据库。

## 项目结构

```
backend/
├── main.go              # 主程序入口
├── config/              # 配置管理
│   └── config.go
├── database/            # 数据库连接
│   └── database.go
├── handlers/            # 请求处理器
│   ├── auth.go          # 认证相关处理器
│   └── health.go        # 健康检查处理器
├── middleware/          # 中间件
│   └── auth.go          # 认证中间件
├── routes/              # 路由配置
│   └── routes.go
├── go.mod               # Go 模块文件
├── go.sum               # 依赖版本锁定
└── README.md            # 项目说明
```

## 快速开始

### 1. 环境配置

复制环境变量示例文件：
```bash
cp .env.example .env
```

根据需要修改 `.env` 文件中的配置。

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 运行项目

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

## API 接口

### 健康检查
- **GET** `/health` - 服务健康检查

### 认证管理
- **POST** `/api/v1/auth/login` - 用户登录
- **POST** `/api/v1/auth/logout` - 用户登出

### 用户信息（需要认证）
- **GET** `/api/v1/user/current` - 获取当前用户信息

### 请求示例

#### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

#### 获取当前用户信息
```bash
curl http://localhost:8080/api/v1/user/current \
  -H "Cookie: session=your-session-cookie"
```

#### 用户登出
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Cookie: session=your-session-cookie"
```

## 数据库支持

项目同时支持 SQLite 和 MySQL 数据库：

- **SQLite**: 适用于开发环境，数据存储在本地文件
- **MySQL**: 适用于生产环境，需要配置数据库连接

默认使用 SQLite 数据库，如需切换到 MySQL，请：
1. 修改 `.env` 文件中的 MySQL 连接配置
2. 在代码中将 `database.GetSQLiteDB()` 改为 `database.GetMySQLDB()`

## 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| SLEEP0_PORT | 服务器端口 | 8080 |
| SLEEP0_ENVIRONMENT | 运行环境 (development/production) | development |
| SLEEP0_DATABASE_TYPE | 数据库类型 (sqlite/mysql) | sqlite |
| SLEEP0_SQLITE_PATH | SQLite 数据库文件路径 | database/app.db |
| SLEEP0_MYSQL_DSN | MySQL 数据库连接字符串 | - |
| SLEEP0_ADMIN_USER | 管理员用户名 | admin |
| SLEEP0_ADMIN_PASS | 管理员密码 | admin123 |
| SLEEP0_SESSION_SECRET | Session 密钥 | your-secret-key-change-this-in-production |

## 主要依赖

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [gin-contrib/sessions](https://github.com/gin-contrib/sessions) - Session 管理

## 开发说明

### 项目架构遵循 Golang + Gin 最佳实践：

1. **handlers/**: 处理 HTTP 请求的处理器函数
   - `auth.go` - 认证相关的处理器（登录、登出、获取用户信息）
   - `health.go` - 健康检查处理器

2. **middleware/**: 中间件函数
   - `auth.go` - 认证中间件，用于保护需要认证的路由

3. **routes/**: 路由配置
   - `routes.go` - 路由注册和分组

4. **config/**: 配置管理
   - `config.go` - 应用配置加载和管理

5. **database/**: 数据库连接
   - `database.go` - 数据库初始化和连接管理

### 添加新功能

1. **添加新的处理器**：
   - 在 `handlers/` 目录下创建新文件
   - 实现处理器函数
   - 在 `routes/routes.go` 中注册路由

2. **添加新的中间件**：
   - 在 `middleware/` 目录下创建新文件
   - 实现中间件函数
   - 在路由中使用中间件

## 下一步计划

- [ ] 添加数据模型 (models/)
- [ ] 添加业务逻辑层 (services/)
- [ ] 添加 API 文档 (Swagger)
- [ ] 添加单元测试
- [ ] 添加 Docker 支持
- [ ] 添加日志中间件 