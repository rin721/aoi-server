# Docker 与 CI

构建和 CI 文件为脚手架提供基础质量门禁。

## 本地构建

```bash
go build -trimpath -ldflags="-s -w" -o bin/go-scaffold-server ./cmd/main
```

## CLI 发布包

`build` 指令会生成可解压运行的发布包，并会出现在交互首页。无参数执行时会进入确认流；
使用 `--yes` 或显式传入任一 build 参数时会直接构建。默认输出到 `build/releases/`：

```bash
go run ./cmd/main build
go run ./cmd/main build --yes
go run ./cmd/main build --target linux/amd64 --target windows/amd64
go run ./cmd/main build --output build/releases --skip-web-generate
go run ./cmd/main build --cgo
```

默认目标为 `linux/amd64`、`windows/amd64` 和 `darwin/amd64`。Windows 目标输出 `.zip`，
Linux 和 macOS 目标输出 `.tar.gz`。每个压缩包内包含：

- `go-scaffold-server` 或 `go-scaffold-server.exe`
- `configs/config.yaml`，来源于 `deploy/config.production.example.yaml`
- `configs/config.example.yaml` 和 `configs/locales/`
- `internal/migrations/`
- `plugins/demo1/plugin.yaml`
- `web/admin/.output/public/`
- 空的 `data/`、`logs/` 目录和 `README.txt`

默认会先在 `web/admin` 执行 `pnpm generate`，并把 Admin WebUI 静态产物打进包内。
可通过下面的参数控制构建期 WebUI 配置：

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `--yes` | `false` | 跳过交互确认，直接使用默认值和显式参数构建 |
| `--webui-build-base-url` | `/admin/` | 传给 `NUXT_APP_BASE_URL` |
| `--webui-api-base-url` | 空字符串 | 传给 `NUXT_PUBLIC_API_BASE_URL` |
| `--webui-show-demo-todo` | `false` | 传给 `NUXT_PUBLIC_SHOW_DEMO_TODO` |
| `--skip-web-generate` | `false` | 跳过 `pnpm generate`，但仍要求 `.output/public/index.html` 已存在 |

默认发布包使用 `CGO_ENABLED=0`，优先保证在当前机器上完成跨平台 Go 构建。该模式下
SQLite 驱动运行时不可用；使用默认 SQLite 配置启动会失败。部署默认包时应切换到
MySQL/Postgres，或在目标平台、交叉 C 工具链可用时使用 `--cgo` 重新构建。

## Docker 构建

```bash
docker build -t go-scaffold:local .
```

Dockerfile 使用 Go build stage 和 slim runtime stage。它会把 production 配置
示例复制到 `/app/configs/config.yaml`，设置 `RIN_CONFIG_PATH`，使用非 root 用户
运行，并启动：

```text
/app/go-scaffold-server server --config=/app/configs/config.yaml
```

生产 Compose 示例通过显式宿主机路径覆盖运行期可变目录：

| 宿主机变量 | 容器路径 | 说明 |
| --- | --- | --- |
| `HOST_CONFIG_FILE` | `/app/configs/config.yaml` | 只读配置文件。 |
| `HOST_DATA_DIR` | `/app/data` | SQLite、媒体库和本地对象存储数据。 |
| `HOST_LOGS_DIR` | `/app/logs` | 文件日志输出目录。 |

`deploy.sh` 会从 `--config-dir`、`--data-dir` 和 `--logs-dir` 派生这些变量；手动
运行 `docker compose` 时需要自行导出。

Dockerfile 还包含 `web-build` 阶段，用 `web/admin` 的 `pnpm generate` 生成 Go 可静态托管的后台产物，并复制到镜像内：

```text
/app/web/admin/.output/public
```

`pnpm build` 可作为 Nuxt 构建检查；实际由 Go `/admin` 托管时必须保证 `.output/public/index.html` 存在。Dockerfile 会在 `pnpm generate` 后检查该文件，缺失时直接让镜像构建失败。

Admin WebUI 相关 Docker build args：

| Build arg | 默认值 | 用途 |
| --- | --- | --- |
| `NUXT_APP_BASE_URL` | `/admin/` | Nuxt 静态资源和路由 baseURL，需要与 Go `webui.mount_path` 对齐。 |
| `NUXT_PUBLIC_API_BASE_URL` | 空字符串 | API 前缀；空值表示同源调用 Go API。 |
| `NUXT_PUBLIC_SHOW_DEMO_TODO` | `false` | 是否在后台导航兜底菜单中显示 Demo Todo。 |

示例：

```bash
docker build \
  --build-arg NUXT_APP_BASE_URL=/admin/ \
  --build-arg NUXT_PUBLIC_API_BASE_URL= \
  --build-arg NUXT_PUBLIC_SHOW_DEMO_TODO=false \
  -t go-scaffold:local .
```

注意：`NUXT_APP_BASE_URL` 是构建期配置。改变后台挂载路径后必须重新构建镜像，不能只通过 Compose 运行期环境变量修复静态资源路径。

## CI Workflow

`.github/workflows/ci.yml` 执行：

- 根据 `go.mod` 设置 Go；
- 根据 `web/admin/pnpm-lock.yaml` 设置 pnpm 和 Node；
- 报告 gofmt drift；
- 根模块 `go test ./... -count=1 -mod=readonly`；
- server 构建；
- Admin WebUI `pnpm typecheck`；
- Admin WebUI `pnpm generate`；
- Docker 构建；
- 空白检查。

当前 CI 会报告 gofmt drift。如果项目希望格式化问题成为硬门禁，需要单独确认后
调整。

## 构建输入

| 输入 | 来源 |
| --- | --- |
| Go 版本 | `go.mod` |
| 服务入口 | `cmd/main` |
| 运行配置 | 镜像内复制的 `deploy/config.production.example.yaml` |
| 运行用户 | Dockerfile 中的非 root UID/GID |

不要把构建期 secret 写入 Docker 镜像。运行期 secret 必须通过环境变量或密钥管理
注入。
