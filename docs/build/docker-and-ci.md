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

## CI Workflow

`.github/workflows/ci.yml` 执行：

- 根据 `go.mod` 设置 Go；
- 报告 gofmt drift；
- 根模块 `go test ./... -count=1 -mod=readonly`；
- server 构建；
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
