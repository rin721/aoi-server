# 目录地图

项目采用务实的分层目录：`cmd/main` 启动进程，`internal/app` 装配应用，`internal` 存放应用专属代码，`pkg` 存放可复用支撑包，`types` 存放共享响应和错误辅助类型。

## 根目录

| 路径 | 职责 | 使用者 |
| --- | --- | --- |
| `go.mod`, `go.sum` | Go 模块依赖 | 全部 Go 包 |
| `Dockerfile` | 多阶段 Linux 镜像构建 | CI 和发布流程 |
| `.github/workflows/ci.yml` | 分支/PR 质量检查 | 维护者 |
| `deploy.sh` | Bash 部署包装脚本 | 手动部署 |

## 应用代码

| 路径 | 职责 | 备注 |
| --- | --- | --- |
| `cmd/main` | CLI 命令声明、进程入口、信号处理 | 保持轻薄，命令以 `CommandSpec` 注册 |
| `internal/app` | 装配根、生命周期、重载编排 | 连接模块和基础设施 |
| `internal/config` | 配置结构、加载、环境变量覆盖、校验、监听 | `envname` 标签是环境变量名来源 |
| `internal/transport/http` | Gin 路由和 HTTP handler 注册 | 连接模块 handler |
| `internal/middleware` | trace、logger、recovery、i18n、CORS 中间件 | 只处理传输层关注点 |
| `internal/modules/demo` | Demo Todo 业务模块 | 分层模块示例 |
| `internal/modules/iam` | 企业级 IAM 业务模块 | 本地账号、组织、权限、会话、MFA 和审计 |
| `internal/migrations` | goose SQL 迁移 | IAM 表结构基线 |

## 基础设施包

| 路径 | 职责 | 边界 |
| --- | --- | --- |
| `pkg/database` | GORM 数据库管理、ping、重载、事务 | 不感知应用模块 |
| `pkg/cache` | Redis 管理器 | 可选基础设施 |
| `pkg/logger` | Zap 日志包装 | 被应用和包复用 |
| `pkg/httpserver` | HTTP 服务启动/关闭/重载包装 | 不定义路由 |
| `pkg/cli` | Cobra 命令路由、flag 解析和 Bubble Tea TUI 首页封装 | 对外暴露项目级 spec，不暴露 Cobra 对象 |
| `pkg/token` | JWT token 管理和 refresh token 哈希 | 不向业务层暴露 JWT 库细节 |
| `pkg/authorization` | Casbin domain RBAC 封装 | 不向业务层暴露 Casbin enforcer |
| `pkg/mfa` | TOTP 生成和校验封装 | 不向业务层暴露 OTP 库细节 |
| `pkg/migrator` | goose 迁移执行封装 | 使用标准 `*sql.DB` 边界 |
| `pkg/executor` | 协程池管理器 | 可选支撑能力 |
| `pkg/crypto` | 哈希和加密辅助函数 | 通用加密辅助能力 |
| `pkg/storage` | 文件存储和 watcher 辅助能力 | 可选存储基础设施 |
| `pkg/sqlgen` | SQL DDL/代码生成辅助能力 | 被 DB CLI 和表结构应用使用 |
| `pkg/i18n` | 消息包和语言辅助能力 | 可选本地化支撑 |
| `pkg/utils` | 小型工具函数 | 避免隐藏业务逻辑 |

## 共享类型

`types` 保存常量、类型化错误和 HTTP 响应辅助类型。`types/result` 当前依赖 Gin，因此还不是完全与传输层无关。

## 文档和运行时

| 路径 | 职责 |
| --- | --- |
| `docs/README.md` | 面向人的工程文档入口 |
| `docs` | 面向人的结构化工程文档 |
| `docs/onboarding` | 新维护者入门、阅读路径和接手检查表 |
| `docs/workflows` | DB 和 IAM CLI 工作流 |
| `docs/ai` | AI 运行时状态、任务树、决策、证据、知识、技能和交接 |
| `AGENTS.md` | Agent 运行时索引 |
