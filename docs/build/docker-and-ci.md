# Docker 与 CI

构建和 CI 文件为脚手架提供基础质量门禁。

## 本地构建

```bash
go build -trimpath -ldflags="-s -w" -o bin/go-scaffold-server ./cmd/main
```

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
