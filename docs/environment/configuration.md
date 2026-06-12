## 无新增配置的后台功能

版本管理发布包使用 `system_versions` 数据表和既有 IAM 权限，不新增 YAML 或环境变量配置项。启用该能力只需要执行数据库迁移，并给角色分配 `version:*` 相关权限。

媒体库使用 `system_media_categories`、`system_media_assets`、`system_media_upload_sessions` 和 `system_media_upload_chunks` 数据表，并复用既有 Storage 配置。只浏览记录或导入外链时不需要对象存储；普通上传、断点上传、本地下载和本地对象删除需要 `storage.enabled=true`，推荐本地使用 `storage.fs_type=basepath`、`storage.base_path=./data`，最终文件会写入 `data/media/YYYY/MM/...`，断点上传临时分片会写入 `data/media/chunks/...`。

客户资源示例使用 `demo_customers` 表和 IAM `customer:*` 权限，不新增 YAML 或环境变量配置项。它跟随 `demo.enabled` 启用，跟随 `demo.apply_schema_on_start` 创建示例表。

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
| System | `RIN_APP_SYSTEM_SEED_DEFAULTS_ON_START` |
| WebUI | `RIN_APP_WEBUI_ENABLED`, `RIN_APP_WEBUI_MOUNT_PATH`, `RIN_APP_WEBUI_DIST_DIR`, `RIN_APP_WEBUI_PUBLIC_BASE_URL`, `NUXT_PUBLIC_SHOW_DEMO_TODO` |
| Plugins | `RIN_APP_PLUGINS_ENABLED`, `RIN_APP_PLUGINS_MANIFESTS`, `RIN_APP_PLUGINS_HEALTH_TIMEOUT_SECONDS`, `RIN_APP_PLUGINS_PROXY_TIMEOUT_SECONDS` |
| Auth | `RIN_APP_AUTH_ENABLED`, `RIN_APP_AUTH_SELF_SIGNUP_ENABLED`, `RIN_APP_AUTH_ISSUER`, `RIN_APP_AUTH_AUDIENCE`, `RIN_APP_AUTH_SIGNING_KEY`, `RIN_APP_AUTH_ACCESS_TOKEN_TTL_SECONDS`, `RIN_APP_AUTH_REFRESH_TOKEN_TTL_SECONDS`, `RIN_APP_AUTH_REFRESH_TOKEN_PEPPER`, `RIN_APP_AUTH_MFA_SECRET_KEY`, `RIN_APP_AUTH_LOGIN_CAPTCHA_ENABLED`, `RIN_APP_AUTH_CAPTCHA_TTL_SECONDS`, `RIN_APP_AUTH_NOTIFICATION_DRIVER`, `RIN_APP_AUTH_SMTP_HOST`, `RIN_APP_AUTH_SMTP_PORT`, `RIN_APP_AUTH_SMTP_USERNAME`, `RIN_APP_AUTH_SMTP_PASSWORD`, `RIN_APP_AUTH_SMTP_FROM`, `RIN_APP_AUTH_PASSWORD_MIN_LENGTH`, `RIN_APP_AUTH_PASSWORD_REQUIRE_LOWER`, `RIN_APP_AUTH_PASSWORD_REQUIRE_UPPER`, `RIN_APP_AUTH_PASSWORD_REQUIRE_NUMBER`, `RIN_APP_AUTH_PASSWORD_REQUIRE_SYMBOL` |
| Migration | `RIN_APP_MIGRATION_AUTO_APPLY`, `RIN_APP_MIGRATION_DIR` |
| CORS | `RIN_APP_CORS_ENABLED`, `RIN_APP_CORS_ALLOW_ORIGINS`, `RIN_APP_CORS_ALLOW_METHODS`, `RIN_APP_CORS_ALLOW_HEADERS`, `RIN_APP_CORS_EXPOSE_HEADERS`, `RIN_APP_CORS_ALLOW_CREDENTIALS`, `RIN_APP_CORS_MAX_AGE` |

`RIN_APP_AUTH_REFRESH_TOKEN_PEPPER` 同时用于 refresh token 和 IAM API Token 的 HMAC hash。生产环境轮换该值前，需要把既有 refresh token 与 API Token 会统一失效这件事纳入发布通知和回滚预案。

完整字段列表以 `internal/config/*` 和 `.env.example` 为准。

## 默认值

本地配置：

- SQLite 路径为 `./data/app.db`；
- Redis 默认关闭；
- JSON-RPC 独立入口默认关闭，启用后默认监听 `127.0.0.1:10099`；
- Demo 模块默认开启，包含公开 Todo 与受保护客户资源示例；
- System 模块默认在启动时幂等补齐内置字典和系统参数；表未迁移时自动跳过，已有数据不覆盖；
- IAM 模块默认开启；
- IAM 登录验证码默认关闭；开启 `RIN_APP_AUTH_LOGIN_CAPTCHA_ENABLED=true` 后，`RIN_APP_AUTH_CAPTCHA_TTL_SECONDS` 控制验证码有效期，默认 120 秒；
- Admin WebUI 默认不在主导航显示 Demo Todo 的静态兜底入口，可在构建前设置 `NUXT_PUBLIC_SHOW_DEMO_TODO=true` 打开；登录后的客户列表入口来自服务端菜单和 `customer:read` 权限；
- 示例配置开启自助注册，并使用 `debug` 通知驱动返回邀请和重置密码调试链接；
- 插件默认关闭；启用后会读取 `plugins.manifests` 中的 JSON/YAML manifest；
- 迁移默认随本地服务启动自动执行，用于首次启动后直接进入浏览器初始化；需要手动检查或生产发布时仍可通过 `db migrate up` 显式应用。

生产示例：

- 监听 `0.0.0.0`；
- 默认关闭 Demo；
- 默认关闭自助注册，通知驱动配置为 `smtp`，不在 API 响应中返回邀请或重置 token；
- 插件默认关闭，sidecar 进程由 Compose、systemd 或 Kubernetes 等外部编排系统管理；
- 迁移默认不自动执行，发布时应先通过 `db migrate status` 和 `db migrate up` 显式应用；
- 敏感值应通过环境变量、CI/CD secrets 或容器编排系统注入。

## 新增配置字段

1. 在对应的 `internal/config/*Config` 结构体中新增字段。
2. 添加 `mapstructure` 和 `envname` 标签。
3. 必要时补充校验。
4. 更新 `configs/config.example.yaml`、`.env.example` 和生产示例。
5. 在 `internal/config` 中新增或调整测试。
6. 字段面向用户或运维时，同步更新本文档。

## Admin WebUI 配置

`web/admin` 是当前后台前端主线。Go 后端默认在 `webui.mount_path=/admin` 挂载静态产物，默认产物目录为 `webui.dist_dir=./web/admin/.output/public`，公开后台地址由 `webui.public_base_url=/admin` 描述。

Nuxt 侧运行时配置保持最小化：

| 变量 | 默认值 | 用途 |
| --- | --- | --- |
| `NUXT_APP_BASE_URL` | `/admin/` | Nuxt 静态资源和路由 baseURL，需要与 Go `webui.mount_path` 对齐。 |
| `NUXT_PUBLIC_API_BASE_URL` | 空字符串 | API 前缀；空值表示同源调用 Go API。 |
| `NUXT_PUBLIC_SHOW_DEMO_TODO` | `false` | 是否在后台导航兜底菜单中显示 Demo Todo。客户列表由服务端菜单返回。 |

用于 Go 静态托管时，先在 `web/admin` 执行 `pnpm generate`，确保 `.output/public/index.html` 存在。`pnpm build` 保留为构建检查，不替代静态产物生成。
