## 无新增配置的后台功能

版本管理发布包使用 `system_versions` 数据表和既有 IAM 权限，不新增 YAML 或环境变量配置项。启用该能力只需要执行数据库迁移，并给角色分配 `version:*` 相关权限。

媒体库使用 `system_media_categories`、`system_media_assets`、`system_media_upload_sessions` 和 `system_media_upload_chunks` 数据表，并复用既有 Storage 配置。只浏览记录或导入外链时不需要对象存储；普通上传、断点上传、本地下载和本地对象删除需要 `storage.enabled=true`，推荐本地使用 `storage.fs_type=basepath`、`storage.base_path=./data`，最终文件会写入 `data/media/YYYY/MM/...`，断点上传临时分片会写入 `data/media/chunks/...`。

客户资源示例使用 `demo_customers` 表和 IAM `customer:*` 权限，不新增 YAML 或环境变量配置项。它跟随 `demo.enabled` 启用，跟随 `demo.apply_schema_on_start` 创建示例表。

模板配置、代码生成、表单生成和导出模板能力当前不是运行时后台功能，不新增 YAML、环境变量或 Nuxt runtime config。现有 `pkg/sqlgen` 与 `pkg/yaml2go` 定位为离线/开发期工具：`pkg/sqlgen` 被 `db` CLI 和 Demo schema 使用，`pkg/yaml2go` 只返回生成代码并不写文件。后续如果要把生成器做成后台工作流，必须先从 `docs/ai/generator-product-spec.md` 定义产品规格、安全边界、写入目录、覆盖策略、字段映射、权限码、审计日志、导出格式和候选配置，再按“新增配置字段”流程补齐统一配置、示例配置和文档。

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

## Docker 部署变量

`deploy.sh` 的基础部署变量只影响 Docker 和宿主机部署行为，不会进入 Go 配置结构。
这些变量可作为脚本默认值，命令行参数优先级更高：

| 变量 | 对应参数 | 用途 |
| --- | --- | --- |
| `DEPLOY_PATH` | `--path` | 保存 Compose 文件的运行目录。 |
| `DEPLOY_IMAGE` | `--image` | 构建或运行的镜像名。 |
| `GITHUB_PROXY_HOST` | `--github-proxy-host` | Git 克隆代理主机，例如 `github-com-gh.helloworlds.eu.org`。 |
| `APP_PORT` | `--port` | 宿主机 HTTP 端口。 |
| `APP_CONTAINER_PORT` | `--server-port` | 容器内 HTTP 端口，并与 `RIN_APP_SERVER_PORT` 对齐。 |
| `HOST_CONFIG_DIR` | `--config-dir` | 宿主机配置目录，脚本会管理其中的 `config.yaml`。 |
| `HOST_DATA_DIR` | `--data-dir` | 宿主机数据目录，映射到容器 `/app/data`。 |
| `HOST_LOGS_DIR` | `--logs-dir` | 宿主机日志目录，映射到容器 `/app/logs`。 |

Compose 模板实际接收 `HOST_CONFIG_FILE`、`HOST_DATA_DIR` 和 `HOST_LOGS_DIR`。手动
运行 Compose 时需要显式导出这些变量；通过 `deploy.sh` 运行时脚本会从目录参数
派生并导出。

## 常用变量

| 范围 | 变量 |
| --- | --- |
| Server | `RIN_APP_SERVER_HOST`, `RIN_APP_SERVER_PORT`, `RIN_APP_SERVER_MODE`, `RIN_APP_SERVER_READ_TIMEOUT`, `RIN_APP_SERVER_WRITE_TIMEOUT`, `RIN_APP_SERVER_IDLE_TIMEOUT` |
| RPC | `RIN_APP_RPC_ENABLED`, `RIN_APP_RPC_HOST`, `RIN_APP_RPC_PORT`, `RIN_APP_RPC_READ_TIMEOUT`, `RIN_APP_RPC_WRITE_TIMEOUT`, `RIN_APP_RPC_IDLE_TIMEOUT` |
| Database | `RIN_APP_DB_DRIVER`, `RIN_APP_DB_HOST`, `RIN_APP_DB_PORT`, `RIN_APP_DB_USER`, `RIN_APP_DB_PASSWORD`, `RIN_APP_DB_NAME`, `RIN_APP_DB_MAX_OPEN_CONNS`, `RIN_APP_DB_MAX_IDLE_CONNS` |
| Redis | `RIN_APP_REDIS_ENABLED`, `RIN_APP_REDIS_HOST`, `RIN_APP_REDIS_PORT`, `RIN_APP_REDIS_PASSWORD`, `RIN_APP_REDIS_DB`, `RIN_APP_REDIS_POOL_SIZE`, `RIN_APP_REDIS_MIN_IDLE_CONNS`, `RIN_APP_REDIS_MAX_RETRIES`, `RIN_APP_REDIS_DIAL_TIMEOUT`, `RIN_APP_REDIS_READ_TIMEOUT`, `RIN_APP_REDIS_WRITE_TIMEOUT` |
| Logger | `RIN_APP_LOG_LEVEL`, `RIN_APP_LOG_FORMAT`, `RIN_APP_LOG_CONSOLE_FORMAT`, `RIN_APP_LOG_FILE_FORMAT`, `RIN_APP_LOG_OUTPUT`, `RIN_APP_LOG_FILE_PATH`, `RIN_APP_LOG_MAX_SIZE`, `RIN_APP_LOG_MAX_BACKUPS`, `RIN_APP_LOG_MAX_AGE` |
| I18n | `RIN_APP_I18N_DEFAULT`, `RIN_APP_I18N_SUPPORTED`, `RIN_APP_I18N_MESSAGES_DIR` |
| Executor | `RIN_APP_EXECUTOR_ENABLED` |
| Storage | `RIN_APP_STORAGE_ENABLED`, `RIN_APP_STORAGE_FS_TYPE`, `RIN_APP_STORAGE_BASE_PATH`, `RIN_APP_STORAGE_ENABLE_WATCH`, `RIN_APP_STORAGE_WATCH_BUFFER_SIZE` |
| Demo | `RIN_APP_DEMO_ENABLED`, `RIN_APP_DEMO_APPLY_SCHEMA_ON_START` |
| System | `RIN_APP_SYSTEM_SEED_DEFAULTS_ON_START` |
| WebUI | `RIN_APP_WEBUI_ENABLED`, `RIN_APP_WEBUI_MOUNT_PATH`, `RIN_APP_WEBUI_DIST_DIR`, `RIN_APP_WEBUI_PUBLIC_BASE_URL`, `NUXT_APP_BASE_URL`, `NUXT_PUBLIC_API_BASE_URL`, `NUXT_PUBLIC_SHOW_DEMO_TODO` |
| Plugins | `RIN_APP_PLUGINS_ENABLED`, `RIN_APP_PLUGINS_MANIFESTS`, `RIN_APP_PLUGINS_HEALTH_TIMEOUT_SECONDS`, `RIN_APP_PLUGINS_PROXY_TIMEOUT_SECONDS` |
| Auth | `RIN_APP_AUTH_ENABLED`, `RIN_APP_AUTH_SELF_SIGNUP_ENABLED`, `RIN_APP_AUTH_ISSUER`, `RIN_APP_AUTH_AUDIENCE`, `RIN_APP_AUTH_SIGNING_KEY`, `RIN_APP_AUTH_ACCESS_TOKEN_TTL_SECONDS`, `RIN_APP_AUTH_REFRESH_TOKEN_TTL_SECONDS`, `RIN_APP_AUTH_REFRESH_TOKEN_PEPPER`, `RIN_APP_AUTH_MFA_ISSUER`, `RIN_APP_AUTH_MFA_SECRET_KEY`, `RIN_APP_AUTH_LOGIN_MAX_FAILURES`, `RIN_APP_AUTH_LOGIN_LOCK_MINUTES`, `RIN_APP_AUTH_LOGIN_CAPTCHA_ENABLED`, `RIN_APP_AUTH_CAPTCHA_TTL_SECONDS`, `RIN_APP_AUTH_INVITATION_TTL_SECONDS`, `RIN_APP_AUTH_PASSWORD_RESET_TTL_SECONDS`, `RIN_APP_AUTH_NOTIFICATION_DRIVER`, `RIN_APP_AUTH_SMTP_HOST`, `RIN_APP_AUTH_SMTP_PORT`, `RIN_APP_AUTH_SMTP_USERNAME`, `RIN_APP_AUTH_SMTP_PASSWORD`, `RIN_APP_AUTH_SMTP_FROM`, `RIN_APP_AUTH_SMTP_FROM_NAME`, `RIN_APP_AUTH_SMTP_STARTTLS`, `RIN_APP_AUTH_PASSWORD_MIN_LENGTH`, `RIN_APP_AUTH_PASSWORD_REQUIRE_LOWER`, `RIN_APP_AUTH_PASSWORD_REQUIRE_UPPER`, `RIN_APP_AUTH_PASSWORD_REQUIRE_NUMBER`, `RIN_APP_AUTH_PASSWORD_REQUIRE_SYMBOL`, `RIN_APP_AUTH_CASBIN_RELOAD_INTERVAL_SECONDS` |
| Migration | `RIN_APP_MIGRATION_AUTO_APPLY`, `RIN_APP_MIGRATION_DIR` |
| CORS | `RIN_APP_CORS_ENABLED`, `RIN_APP_CORS_ALLOW_ORIGINS`, `RIN_APP_CORS_ALLOW_METHODS`, `RIN_APP_CORS_ALLOW_HEADERS`, `RIN_APP_CORS_EXPOSE_HEADERS`, `RIN_APP_CORS_ALLOW_CREDENTIALS`, `RIN_APP_CORS_MAX_AGE` |

`RIN_APP_AUTH_REFRESH_TOKEN_PEPPER` 同时用于 refresh token 和 IAM API Token 的 HMAC hash。生产环境轮换该值前，需要把既有 refresh token 与 API Token 会统一失效这件事纳入发布通知和回滚预案。

完整字段列表以 `internal/config/*` 和 `.env.example` 为准。生产 Compose 示例暴露
部署模板常用的 `RIN_APP_*` 运行期变量，`deploy.sh` 也为这些变量提供显式参数。
后续新增面向运维的配置字段时，需要同步 `internal/config`、示例配置、`.env.example`、
Compose 模板、部署脚本参数和本文档。

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

AI 渐进式审计流程记录在 `docs/ai/progressive-project-audit.md`，它本身不新增运行时
配置。后续切片一旦涉及阈值、状态、路径、接口、刷新策略、布局参数、字段映射
或构建产物路径，必须先在该任务书记录配置落点，再按上面的步骤补齐统一配置、
示例配置、测试和文档。

HTTP API 公共前缀 `/api/v1` 是对外接口契约，不是 YAML、env 或 Nuxt runtime
config。后端路由、API catalog、操作记录判定和服务端生成的媒体下载 URL 通过
`types/constants` 中的 `APIBasePath`、`APIPath()` 和 `MediaAssetDownloadPath()`
维护；前端调用路径仍由 `web/admin/app/config/admin-api.ts` 维护。若未来要让 API
前缀可配置，需要先重新设计 OpenAPI、前端 endpoint、权限同步、操作记录筛选和
部署兼容策略，不能只新增一个环境变量。

## Admin WebUI 配置

`web/admin` 是当前后台前端主线。Go 后端默认在 `webui.mount_path=/admin` 挂载静态产物，默认产物目录为 `webui.dist_dir=./web/admin/.output/public`，公开后台地址由 `webui.public_base_url=/admin` 描述。`webui.mount_path` 必须是非根绝对路径，不能设置为 `/`，避免 SPA fallback 覆盖 API、健康检查和就绪检查路由。

Nuxt 侧运行时配置保持最小化：

| 变量 | 默认值 | 用途 |
| --- | --- | --- |
| `NUXT_APP_BASE_URL` | `/admin/` | Nuxt 静态资源和路由 baseURL，需要与 Go `webui.mount_path` 对齐。 |
| `NUXT_PUBLIC_API_BASE_URL` | 空字符串 | API 前缀；空值表示同源调用 Go API。 |
| `NUXT_PUBLIC_SHOW_DEMO_TODO` | `false` | 是否在后台导航兜底菜单中显示 Demo Todo。客户列表由服务端菜单返回。 |

用于 Go 静态托管时，先在 `web/admin` 执行 `pnpm generate`，确保 `.output/public/index.html` 存在。`pnpm build` 保留为构建检查，不替代静态产物生成。

Docker 镜像构建时通过 build args 注入 Nuxt 构建期配置：

| Build arg | 默认值 | 对应运行时配置 |
| --- | --- | --- |
| `NUXT_APP_BASE_URL` | `/admin/` | `webui.mount_path` |
| `NUXT_PUBLIC_API_BASE_URL` | 空字符串 | 网关或跨域部署时的 API 入口 |
| `NUXT_PUBLIC_SHOW_DEMO_TODO` | `false` | 仅控制前端兜底菜单，不改变后端 Demo 配置 |

`deploy.sh` 提供 `--config-dir`、`--data-dir`、`--logs-dir`、`--webui-mount-path`、`--webui-build-base-url`、`--webui-api-base-url`、`--webui-show-demo-todo`、`--webui-check` 和 `--webui-check-path`。`--webui-mount-path` 与 Go 配置一样必须是非根绝对路径；未显式传 `--webui-build-base-url` 时，脚本会从 `--webui-mount-path` 派生带尾斜杠的 Nuxt baseURL，避免 Go 挂载路径和 Nuxt 静态资源路径不一致。

## System 配置 API

`PATCH /api/v1/system/config` 面向后台系统配置页，读取当前运行时快照并支持受控持久化。响应中的敏感字段会脱敏；带 `${VAR}` 或 `${VAR:default}` 的环境变量占位值不会被接口回写到配置文件。传入 `persist=true` 时，只会持久化后端明确支持的标量字段和字符串列表字段，避免把运行时派生值、密钥或未知结构写回 `configs/config.yaml`。

该接口不会改变 HTTP API 公共前缀，也不会改变 Nuxt 构建期 baseURL。需要改后台挂载路径时，仍要同时更新 Go `webui.*` 配置和 Nuxt 构建产物。

## System Center CLI 运行态

`init`、`run` 和 `service` 命令使用独立运行态目录保存受管服务元数据：

| 变量 | 默认值 | 用途 |
| --- | --- | --- |
| `RIN_CLI_RUNTIME_DIR` | `data/cli-runtime` | System Center 保存 PID、创建时间、日志路径和服务状态的目录。 |
| `RIN_CLI_MANAGED` | 由 `run server` 设置 | 标记当前服务进程由 CLI 托管。 |
| `RIN_CLI_SERVICE` | 由 `run server` 设置 | 标记受管服务名称，目前为 `server`。 |

这些变量只服务本地 CLI 运行态，不属于 YAML 配置结构，也不会被 System 配置 API 持久化。

## Server Status Dashboard 配置

服务器状态 Dashboard 的后端静态托管仍由 `webui` 配置控制：

- `webui.mount_path`：Go 挂载后台 SPA 的路径，默认 `/admin`。
- `webui.dist_dir`：Go 读取前端静态产物的目录，默认 `./web/admin/.output/public`。
- `webui.public_base_url`：公开访问后台时使用的基础路径，通常与 `mount_path` 保持一致。

前端运行期配置仍走 Nuxt public runtime config：

- `NUXT_APP_BASE_URL` 应与 `webui.mount_path` 对齐。
- `NUXT_PUBLIC_API_BASE_URL` 为空时使用同源 Go API；跨域或网关部署时由环境变量注入。

Server Status 页面自己的展示治理不新增后端配置项，前端入口如下：

- `web/admin/app/config/admin-api.ts`：集中维护后台 API endpoint；页面和 composable 不直接散落新增接口路径。
- `web/admin/app/config/admin-auto-refresh.ts`：集中维护后台通用自动刷新默认值、最小间隔、倒计时 tick、默认手动刷新冷却、时间单位、时间 locale 和共享控件/状态文案。
- `web/admin/app/config/server-status-dashboard.ts`：集中维护指标阈值、状态文案、状态权重、KPI 顺序、字段 label、页面级刷新策略、手动刷新冷却、格式化规则和空状态文案。
- `web/admin/app/assets/css/main.css`：通过 `--aoi-admin-*` token 控制卡片间距、自动刷新控件间距、局部滚动高度、CPU 行尺寸和数据状态尺寸；自动刷新控件依赖 `flex-wrap` 和共享 gap token 做窄屏换行，不再在组件内单独维护断点。

`useAdminAutoRefresh()` 只把真实浏览器点击事件识别为手动刷新并应用 `manualCooldownMs`；自动静默刷新、筛选分页后的程序化刷新和数据变更后的 reload 不受冷却影响，避免页面状态因为按钮防抖而跳过必要的数据更新。

示例配置和文档不得包含真实账号、密码、Token 或生产地址。需要本地登录验证时，使用安全渠道获取本地测试账号信息。
