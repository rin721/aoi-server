# 配置说明

配置由 `internal/config` 加载。默认配置文件是 `configs/config.yaml`，示例文件位于 `configs/config.example.yaml` 和 `deploy/config.production.example.yaml`。

## 配置路径解析

`server`、`db` 和 `iam` 命令都支持 `--config`：

```bash
go run ./cmd/main server --config=configs/config.yaml
go run ./cmd/main db --config=configs/config.yaml --operation=schema
go run ./cmd/main iam bootstrap-admin --config=configs/config.yaml --org-code=acme --username=admin --email=admin@example.com --password-stdin
```

如果未传入 `--config`，进程会先读取 `RIN_CONFIG_PATH`，再回退到默认路径。

## 加载顺序

1. 从命令参数、环境变量或默认值选择配置路径。
2. 当前工作目录存在 `.env` 时加载它。
3. 读取 YAML 配置。
4. 替换 `${VAR}` 和 `${VAR:default}` 占位符。
5. 反序列化到 `internal/config.Config`。
6. 按 `envname` 标签应用环境变量覆盖。
7. 校验最终配置。
8. 存入配置管理器。

真实系统环境变量优先级高于 `.env`。

## 环境变量命名

当前应用前缀是 `Rin`，所以应用配置环境变量使用 `RIN_APP_*`。配置路径变量是 `RIN_CONFIG_PATH`。

示例：

```text
RIN_APP_DB_HOST
DB_HOST
```

优先使用带前缀名称；不带前缀名称只作为兼容 fallback。

## 常用变量

| 范围 | 变量 |
| --- | --- |
| Server | `RIN_APP_SERVER_HOST`, `RIN_APP_SERVER_PORT`, `RIN_APP_SERVER_MODE`, `RIN_APP_SERVER_READ_TIMEOUT`, `RIN_APP_SERVER_WRITE_TIMEOUT`, `RIN_APP_SERVER_IDLE_TIMEOUT` |
| RPC | `RIN_APP_RPC_ENABLED`, `RIN_APP_RPC_HOST`, `RIN_APP_RPC_PORT`, `RIN_APP_RPC_READ_TIMEOUT`, `RIN_APP_RPC_WRITE_TIMEOUT`, `RIN_APP_RPC_IDLE_TIMEOUT` |
| Database | `RIN_APP_DB_DRIVER`, `RIN_APP_DB_HOST`, `RIN_APP_DB_PORT`, `RIN_APP_DB_USER`, `RIN_APP_DB_PASSWORD`, `RIN_APP_DB_NAME`, `RIN_APP_DB_MAX_OPEN_CONNS`, `RIN_APP_DB_MAX_IDLE_CONNS` |
| Redis | `RIN_APP_REDIS_ENABLED`, `RIN_APP_REDIS_HOST`, `RIN_APP_REDIS_PORT`, `RIN_APP_REDIS_PASSWORD`, `RIN_APP_REDIS_DB`, `RIN_APP_REDIS_POOL_SIZE` |
| Logger | `RIN_APP_LOG_LEVEL`, `RIN_APP_LOG_FORMAT`, `RIN_APP_LOG_OUTPUT`, `RIN_APP_LOG_FILE_PATH`, `RIN_APP_LOG_MAX_SIZE`, `RIN_APP_LOG_MAX_BACKUPS`, `RIN_APP_LOG_MAX_AGE` |
| I18n | `RIN_APP_I18N_DEFAULT`, `RIN_APP_I18N_SUPPORTED`, `RIN_APP_I18N_MESSAGES_DIR` |
| Executor | `RIN_APP_EXECUTOR_ENABLED` |
| Storage | `RIN_APP_STORAGE_ENABLED`, `RIN_APP_STORAGE_FS_TYPE`, `RIN_APP_STORAGE_BASE_PATH`, `RIN_APP_STORAGE_ENABLE_WATCH`, `RIN_APP_STORAGE_WATCH_BUFFER_SIZE` |
| Demo | `RIN_APP_DEMO_ENABLED`, `RIN_APP_DEMO_APPLY_SCHEMA_ON_START` |
| Auth | `RIN_APP_AUTH_ENABLED`, `RIN_APP_AUTH_ISSUER`, `RIN_APP_AUTH_AUDIENCE`, `RIN_APP_AUTH_SIGNING_KEY`, `RIN_APP_AUTH_ACCESS_TOKEN_TTL_SECONDS`, `RIN_APP_AUTH_REFRESH_TOKEN_TTL_SECONDS`, `RIN_APP_AUTH_REFRESH_TOKEN_PEPPER`, `RIN_APP_AUTH_MFA_SECRET_KEY` |
| Migration | `RIN_APP_MIGRATION_AUTO_APPLY`, `RIN_APP_MIGRATION_DIR` |
| CORS | `RIN_APP_CORS_ENABLED`, `RIN_APP_CORS_ALLOW_ORIGINS`, `RIN_APP_CORS_ALLOW_METHODS`, `RIN_APP_CORS_ALLOW_HEADERS`, `RIN_APP_CORS_EXPOSE_HEADERS`, `RIN_APP_CORS_ALLOW_CREDENTIALS`, `RIN_APP_CORS_MAX_AGE` |

完整字段列表以 `internal/config/*` 和 `.env.example` 为准。

## 默认值

本地配置：

- SQLite 路径为 `./data/app.db`；
- Redis 默认关闭；
- JSON-RPC 独立入口默认关闭，启用后默认监听 `127.0.0.1:10099`；
- Demo 模块默认开启；
- IAM 模块默认开启；
- 迁移默认不自动执行，应通过 `db migrate up` 显式应用。

生产示例：

- 监听 `0.0.0.0`；
- 默认关闭 Demo；
- 敏感值应通过环境变量、CI/CD secrets 或容器编排系统注入。

## 新增配置字段

1. 在对应的 `internal/config/*Config` 结构体中新增字段。
2. 添加 `mapstructure` 和 `envname` 标签。
3. 必要时补充校验。
4. 更新 `configs/config.example.yaml`、`.env.example` 和生产示例。
5. 在 `internal/config` 中新增或调整测试。
6. 字段面向用户或运维时，同步更新本文档。
