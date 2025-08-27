# 任务对话创建Docker命令的映射逻辑

## 概述

XSha 系统使用 Docker 容器来隔离执行AI驱动的任务对话。本文档详细说明了如何从 `TaskConversation` 实体映射生成 Docker 执行命令的完整逻辑。

## 核心组件

### 1. 数据结构映射

#### TaskConversation 结构体
```go
type TaskConversation struct {
    ID        uint           `json:"id"`
    TaskID    uint           `json:"task_id"`
    Task      *Task          `json:"task"`
    Content   string         `json:"content"`        // 任务对话内容
    Status    ConversationStatus `json:"status"`    // 对话状态
    EnvParams string         `json:"env_params"`    // 环境参数JSON
    CreatedBy string         `json:"created_by"`
}
```

#### Task 和 DevEnvironment 相关字段
```go
type Task struct {
    ProjectID        uint            `json:"project_id"`
    DevEnvironmentID *uint           `json:"dev_environment_id"`
    DevEnvironment   *DevEnvironment `json:"dev_environment"`
    SessionID        string          `json:"session_id"`
    WorkspacePath    string          `json:"workspace_path"`
}

type DevEnvironment struct {
    Type         string  `json:"type"`          // 环境类型：claude-code, opencode, gemini-cli
    DockerImage  string  `json:"docker_image"`  // Docker镜像名
    CPULimit     float64 `json:"cpu_limit"`     // CPU限制
    MemoryLimit  int64   `json:"memory_limit"`  // 内存限制(MB)
    EnvVars      string  `json:"env_vars"`      // 环境变量JSON
    SessionDir   string  `json:"session_dir"`   // 会话目录
    SystemPrompt string  `json:"system_prompt"` // 系统提示
}
```

## Docker 命令构建逻辑

### 2. 基础命令构建 (buildDockerCommandCore)

位于 `backend/services/executor/docker_executor.go:148`

```go
func (d *dockerExecutor) buildDockerCommandCore(conv *database.TaskConversation, workspacePath string, opts buildDockerCommandOptions) string
```

#### 命令结构映射：

1. **基础命令**
   ```bash
   docker run --rm
   ```

2. **交互式标志** (根据 includeStdinFlag 选项)
   ```bash
   -i  # 保持 STDIN 开启
   ```

3. **容器命名** (根据 containerName 选项)
   ```bash
   --name=xsha-task-{taskID}-conv-{conversationID}
   ```

4. **卷挂载策略**

   **容器内执行时** (`utils.IsRunningInContainer()` 返回 true):
   ```bash
   -v xsha_workspaces:/app
   -v xsha_dev_sessions:/xsha_dev_sessions
   -w /app/{workspacePath}
   ```

   **主机执行时**:
   ```bash
   -v {absoluteWorkspacePath}:/app
   -v {absoluteSessionDir}:/home/xsha  # 如果 SessionDir 存在
   -w /app
   ```

5. **资源限制映射**
   ```bash
   --cpus={DevEnvironment.CPULimit}      # 如果 CPULimit > 0
   --memory={DevEnvironment.MemoryLimit}m # 如果 MemoryLimit > 0
   ```

6. **环境变量映射**
   - 从 `DevEnvironment.EnvVars` JSON 解析
   - 每个变量添加 `-e KEY=VALUE`
   - 敏感值在日志中会被掩码处理

### 3. AI 命令构建 (buildAICommand)

位于 `backend/services/executor/docker_executor.go:208`

```go
func (d *dockerExecutor) buildAICommand(envType, content string, isInContainer bool, task *database.Task, devEnv *database.DevEnvironment, conv *database.TaskConversation) []string
```

#### 根据环境类型映射不同命令：

**claude-code 类型** (最复杂的映射):
```bash
claude -p --output-format=stream-json --verbose
```

参数映射：
- `SessionID` → `-r {task.SessionID}` (如果存在)
- `EnvParams.model` → `--model {modelName}` (如果不是 "default")
- `EnvParams.is_plan_mode` → `--permission-mode plan` (如果为 true) 或 `--dangerously-skip-permissions` (如果为 false 或未设置)
- `Project.SystemPrompt` → `--append-system-prompt "{prompt}"`
- `DevEnvironment.SystemPrompt` → `--append-system-prompt "{prompt}"`
- `Content` → 最终的用户输入参数
- `SessionDir` → `-d /xsha_dev_sessions/{sessionDir}` (容器内执行时)

**权限模式说明**：
- 当 `is_plan_mode=true` 时，使用 `--permission-mode plan` 参数
- 当 `is_plan_mode=false` 或未设置时，使用 `--dangerously-skip-permissions` 参数
- 这两个参数是互斥的，不能同时使用

最终构造示例：
```bash
# 普通模式
--command "claude -p --output-format=stream-json --dangerously-skip-permissions --verbose [参数...] \"用户内容\""

# 计划模式
--command "claude -p --output-format=stream-json --permission-mode plan --verbose [参数...] \"用户内容\""
```

**opencode 和 gemini-cli 类型**:
```bash
"{content}"  # 直接使用对话内容
```

### 4. 容器命名策略

```go
func (d *dockerExecutor) generateContainerName(conv *database.TaskConversation) string {
    return fmt.Sprintf("xsha-task-%d-conv-%d", conv.TaskID, conv.ID)
}
```

容器名格式：`xsha-task-{任务ID}-conv-{对话ID}`

## 完整执行流程

### 5. 任务执行映射流程

位于 `backend/services/executor/service.go:325` (executeTask方法)

1. **工作空间准备**
   - 创建或获取任务工作空间
   - 克隆Git仓库 (如果不存在)
   - 切换到工作分支

2. **附件处理**
   - 复制对话附件到工作空间
   - 替换内容中的附件标签为实际路径
   - 创建临时对话对象用于执行

3. **Docker命令构建**
   ```go
   dockerCmdForLog := s.dockerExecutor.BuildCommandForLog(&tempConv, workspacePath)
   containerID, err := s.dockerExecutor.ExecuteWithContainerTracking(ctx, &tempConv, workspacePath, execLog.ID)
   ```

4. **执行监控**
   - 实时日志流式处理
   - 错误捕获和处理
   - 容器生命周期管理

5. **后处理**
   - 清理工作空间附件
   - 提交Git更改
   - 更新对话状态

### 6. 特殊配置处理

#### 环境参数解析
```go
if conv != nil && conv.EnvParams != "" && conv.EnvParams != "{}" {
    var envParams map[string]interface{}
    if err := json.Unmarshal([]byte(conv.EnvParams), &envParams); err == nil {
        if model, exists := envParams["model"]; exists {
            if modelStr, ok := model.(string); ok && modelStr != "default" {
                claudeCommand = append(claudeCommand, "--model", modelStr)
            }
        }
    }
}
```

#### 敏感值处理
- 使用 `maskEnvVars` 选项决定是否在日志中掩码环境变量
- 使用 `utils.MaskSensitiveValue()` 处理敏感数据

## 示例

### 典型的 Claude Code 任务命令示例：

**输入数据**:
```json
{
  "task": {
    "id": 123,
    "session_id": "session_abc",
    "dev_environment": {
      "type": "claude-code",
      "docker_image": "anthropic/claude-code:latest",
      "cpu_limit": 2.0,
      "memory_limit": 2048,
      "env_vars": "{\"API_KEY\":\"secret\"}",
      "system_prompt": "You are a helpful coding assistant"
    }
  },
  "id": 456,
  "content": "Fix the bug in main.py",
  "env_params": "{\"model\":\"claude-3-opus\"}"
}
```

**生成的Docker命令**:
```bash
docker run --rm -i --name=xsha-task-123-conv-456 \
  -v /path/to/workspace:/app \
  -w /app \
  --cpus=2.00 \
  --memory=2048m \
  -e API_KEY=secret \
  anthropic/claude-code:latest \
  --command "claude -p --output-format=stream-json --dangerously-skip-permissions --verbose -r session_abc --model claude-3-opus --append-system-prompt \"You are a helpful coding assistant\" \"Fix the bug in main.py\""
```

## 总结

这个映射系统提供了灵活且强大的方式来将高级任务对话转换为底层Docker执行命令，支持：

- 多种AI环境类型
- 灵活的资源配置
- 安全的环境变量处理
- 完整的生命周期管理
- 实时日志监控

通过这种设计，XSha 能够安全、隔离地执行各种AI驱动的开发任务。