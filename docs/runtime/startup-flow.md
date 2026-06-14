# 启动流程

主进程从 `cmd/main` 开始。

```text
main.go
  -> cli.NewApp(cli.Config)
  -> 注册 server/db/iam/build/run/service/init CommandSpec
  -> 无参数进入 TUI 首页，或按参数分派 Cobra 命令
  -> server command
  -> runApp
  -> app.New
  -> application.Run
```

## CLI 入口

`cmd/main/main.go` 只负责装配 `pkg/cli` 应用、注册顶层命令并处理退出码。当前顶层命令包含 `server`、`db`、`iam`、`build`、`run`、`service` 和 `init`。`pkg/cli` 内部封装 Cobra 和 Bubble Tea/Lip Gloss v2：

- `--help`、`--version`、子命令和 flag 解析由 Cobra 处理；
- 无参数运行时进入默认 TUI 首页，用于浏览命令和查看帮助；
- 命令层通过 `CommandSpec` 和 `FlagSpec` 声明接口，不直接暴露 Cobra 对象。

## server 命令

`cmd/main/app.go` 定义 `server` 命令，并通过 `Spec()` 转换为 `cli.CommandSpec` 注册到根命令。它支持 `--config` 并读取 `RIN_CONFIG_PATH`。命令层只处理入口参数，实际启动交给 `runApp`。

## System Center 命令

`init`、`run` 和 `service` 是本地 System Center 入口。`init` 会执行迁移、可选 Demo schema、System 默认数据同步、API/权限同步、可选管理员和服务 API Token 创建；`run server` 会以受管服务方式启动后端；`service status/info/logs/terminal/restart/stop server` 读取默认 `data/cli-runtime` 下的运行态记录。可以通过 `RIN_CLI_RUNTIME_DIR` 改变运行态目录；受管服务进程会设置 `RIN_CLI_MANAGED` 和 `RIN_CLI_SERVICE`。

这些命令也支持通用链式 prompt 参数。`--chain.*` 会在 Cobra 解析前剥离，作为后续菜单、确认、输入和密码 prompt 的自动答案，不会进入普通 flag 集合；缺失的 key 仍回到 TUI 或 stdin prompt。示例：

```powershell
.\tmp\go-scaffold-server.exe run --chain.service=server --chain.config=configs/config.yaml --chain.privacy=false
.\tmp\go-scaffold-server.exe run --chain.service=db --chain.config=configs/config.yaml
.\tmp\go-scaffold-server.exe service --chain.action=logs --chain.logs.follow=true
.\tmp\go-scaffold-server.exe build --chain.build.target=current --chain.build.output=build/releases --chain.build.generate-web=false --chain.build.cgo=false --chain.build.proceed=true
```

`run` 的动态隐私配置项使用 `privacy.<path>.action` 和 `privacy.<path>.value`，例如 `--chain.privacy.auth.signing_key.action=force-file --chain.privacy.auth.signing_key.value=generate`。链式 key 支持 `*` 通配，动态路径较多时可使用 `--chain.privacy.*.action=skip` 或 `--chain.privacy.auth.*.value=generate`。旧的 `run --service=server --yes` 仍保留为兼容入口，内部会映射到 `service`、`config` 和 `privacy=false` 这些 prompt 答案。

受管服务启动时会记录实际派生的可执行文件路径。普通构建产物会直接复用当前可执行文件；如果入口来自 `go run` 的 `go-build.../exe/main(.exe)` 临时路径，CLI 会复制到 `data/cli-runtime/bin/go-scaffold-managed(.exe)` 再启动后台服务，避免 Windows 清理 Go 临时 exe 时出现 `unlinkat ... Access is denied`。长期后台运行仍建议先 `go build` 出固定二进制，再执行 `run server`。

## 应用构建

`internal/app.New` 通过应用子包构建应用：

| 包 | 作用 |
| --- | --- |
| `initapp` | 创建核心服务、基础设施、模块和传输层 |
| `mainapp` | 构建主应用运行模式；当前真实模式是 `server` |
| `lifecycleapp` | 启动并关闭 HTTP、RPC、存储、执行器、缓存、数据库和日志 |
| `reloadapp` | 将配置变化应用到可重载子系统 |

## 表结构应用

启动期间，只有在 Demo 模块启用且 `demo.apply_schema_on_start` 为 true 时才会应用 Demo 表结构。

IAM、System 和插件相关表结构由 goose 迁移管理。本地默认 `migration.auto_apply=true`，`internal/app/initapp` 会在模块装配前运行迁移，便于首次启动后直接进入 `/admin` 浏览器初始化。生产示例保持 `migration.auto_apply=false`，应通过 `db migrate up` 显式应用。

System 模块在装配时会读取 `system.seed_defaults_on_start`。默认开启时，它会幂等补齐 `system.status`、`http.method`、`operation.result` 三组内置字典，以及 `admin.title`、`admin.home_path`、`system.reference` 三个系统参数。这个过程会跳过未迁移或不可用的 system 表，并且不会覆盖已经存在的参数值。

## HTTP 启动

HTTP 服务由 `pkg/httpserver` 包装。端口绑定错误会同步返回。

WebUI 静态托管也在 HTTP 路由装配阶段注册。Go 服务默认读取 `web/admin/.output/public` 并挂载到 `/admin`；API、健康检查和就绪检查优先注册，SPA fallback 只覆盖 WebUI 挂载路径。

## RPC 启动

JSON-RPC 服务由 `pkg/rpcserver` 包装，默认关闭。开启 `rpc.enabled=true` 后，`server` 进程会在 `rpc.host:rpc.port` 上额外监听 `/rpc` 和 `/health`；RPC 端口绑定失败会让启动返回错误，并回滚已启动的 HTTP 服务。

## 关闭流程

`cmd/main/run.go` 监听 `SIGINT` 和 `SIGTERM`。关闭过程使用配置的超时时间，并按以下顺序释放资源：

1. 主 HTTP 服务；
2. RPC 服务；
3. 存储；
4. 执行器；
5. 缓存；
6. 数据库；
7. 日志 sync。
