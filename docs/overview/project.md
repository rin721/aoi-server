# 项目概述

`go-scaffold` 是一个可运行的 Go 后端脚手架，包含服务代码、基础设施包、测试、Docker 构建文件、部署示例和 AI 运行时制品。

## 当前能力

| 能力 | 状态 |
| --- | --- |
| HTTP 服务 | 由 `internal/transport/http` 和 `pkg/httpserver` 实现 |
| 配置 | 支持 YAML、`.env`、环境变量覆盖、校验和监听重载 |
| 数据库 | `pkg/database` 支持 SQLite、MySQL、PostgreSQL |
| Demo 模块 | Todo CRUD，采用 handler/service/repository/model 分层 |
| IAM 模块 | 本地账号、JWT、组织租户、Casbin 权限、邀请、找回密码、TOTP MFA、会话撤销和审计 |
| 数据库迁移 | `pkg/migrator` 封装 goose，CLI 提供显式迁移命令 |
| 存储 | 本地文件系统抽象和可选 watcher 辅助能力 |
| CLI/TUI | Cobra 命令路由和 Bubble Tea/Lip Gloss v2 交互式首页 |
| SQL 生成 | Go 结构体到 SQL 的辅助工具，用于 DB CLI 和表结构应用 |
| CI/构建 | Go 测试、服务构建、Docker 构建和空白字符检查 |
| 部署 | 生产配置示例、Docker Compose 示例、本地部署脚本和远程工作流 |

## 当前非目标

- SSO/OIDC/SAML 外部身份提供商；
- 短信、邮件验证码等非 TOTP MFA；
- 生产发布窗口、回滚演练和迁移审计治理；
- 外部扩展运行时；
- v1 发布保证。

## 运行时默认值

本地默认配置是 `configs/config.yaml`。服务监听 `127.0.0.1:9999`，使用 SQLite `./data/app.db`，关闭 Redis，并启用 Demo Todo 与 IAM 模块。迁移默认不自动执行，应通过 `db migrate up` 显式应用。

生产示例位于 `deploy` 目录，默认关闭 Demo 模块。
