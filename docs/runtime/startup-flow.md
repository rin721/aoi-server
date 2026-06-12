# 启动流程

主进程从 `cmd/main` 开始。

```text
main.go
  -> cli.NewApp(cli.Config)
  -> 注册 server/db/iam CommandSpec
  -> 无参数进入 TUI 首页，或按参数分派 Cobra 命令
  -> server command
  -> runApp
  -> app.New
  -> application.Run
```

## CLI 入口

`cmd/main/main.go` 只负责装配 `pkg/cli` 应用、注册顶层命令并处理退出码。`pkg/cli` 内部封装 Cobra 和 Bubble Tea/Lip Gloss v2：

- `--help`、`--version`、子命令和 flag 解析由 Cobra 处理；
- 无参数运行时进入默认 TUI 首页，用于浏览命令和查看帮助；
- 命令层通过 `CommandSpec` 和 `FlagSpec` 声明接口，不直接暴露 Cobra 对象。

## server 命令

`cmd/main/app.go` 定义 `server` 命令，并通过 `Spec()` 转换为 `cli.CommandSpec` 注册到根命令。它支持 `--config` 并读取 `RIN_CONFIG_PATH`。命令层只处理入口参数，实际启动交给 `runApp`。

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

IAM 表结构由 goose 迁移管理。默认 `migration.auto_apply=false`，应通过 `db migrate up` 显式应用；如果显式开启自动迁移，`internal/app/initapp` 会在模块装配前运行迁移。

System 模块在装配时会读取 `system.seed_defaults_on_start`。默认开启时，它会幂等补齐 `system.status`、`http.method`、`operation.result` 三组内置字典，以及 `admin.title`、`admin.home_path`、`system.reference` 三个系统参数。这个过程会跳过未迁移或不可用的 system 表，并且不会覆盖已经存在的参数值。

## HTTP 启动

HTTP 服务由 `pkg/httpserver` 包装。端口绑定错误会同步返回。

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
