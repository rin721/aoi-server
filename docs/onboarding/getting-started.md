# 新人接手指南

本文面向刚接触 Go 或刚接手本项目的维护者。目标不是一次读完所有代码，而是先跑起来，再沿一条最小业务链路理解项目。

## 先记住这张地图

项目可以先粗略分成四层：

| 层 | 路径 | 先怎么理解 |
| --- | --- | --- |
| 进程入口 | `cmd/main` | 接收命令、解析参数、启动或执行 CLI 任务 |
| 应用装配 | `internal/app` | 把配置、数据库、模块、HTTP/RPC 服务组装起来 |
| 业务模块 | `internal/modules` | 按 `model -> repository -> service -> handler` 写业务 |
| 基础设施 | `pkg` | 数据库、日志、HTTP server、缓存、迁移、Token 等可复用能力 |

第一遍阅读时，不要从 `pkg` 开始。`pkg` 是支撑包，细节多但不代表业务主线。

## 第一天阅读路线

建议按下面顺序打开文件：

1. `docs/overview/project.md`
2. `docs/structure/directory-map.md`
3. `cmd/main/main.go`
4. `cmd/main/run.go`
5. `internal/app/app.go`
6. `internal/app/mainapp/mode.go`
7. `internal/app/initapp/modules.go`
8. `internal/transport/http/router.go`
9. `internal/modules/demo/handler/todo.go`
10. `internal/modules/demo/service/todo.go`
11. `internal/modules/demo/repository/todo.go`
12. `internal/modules/demo/model/todo.go`

这条路线会把你带过一遍：命令入口、应用启动、路由注册、HTTP handler、业务服务、数据库访问和模型定义。

## 本地跑起来

首次使用 IAM 前，先执行迁移：

```powershell
go run ./cmd/main db migrate up --config=configs/config.yaml
```

创建初始管理员：

```powershell
"change-this-local-password" | go run ./cmd/main iam bootstrap-admin --config=configs/config.yaml --org-code=acme --org-name="Acme Corp" --username=admin --email=admin@example.com --password-stdin
```

上面的密码只适合本地练习。生产或共享环境应通过 secrets、CI 变量或受控输入管道传入密码。

启动服务：

```powershell
go run ./cmd/main server --config=configs/config.yaml
```

检查服务：

```powershell
curl http://127.0.0.1:9999/health
curl http://127.0.0.1:9999/ready
```

本地默认使用 SQLite `./data/app.db`。Demo Todo 模块默认启用，且 `demo.apply_schema_on_start=true`，所以 Demo 表结构会在服务启动时自动应用。IAM 表结构默认不会随服务启动自动迁移，需要显式执行 `db migrate up`。

## 一条请求怎么走

以 Demo Todo 为例，请求链路是：

```text
internal/transport/http/router.go
  -> internal/modules/demo/handler/todo.go
  -> internal/modules/demo/service/todo.go
  -> internal/modules/demo/repository/todo.go
  -> internal/modules/demo/model/todo.go
  -> pkg/database
```

读代码时先回答三个问题：

1. 路由在哪里注册？
2. handler 从请求里取了什么参数？
3. service 做了什么业务判断，然后让 repository 读写了哪些数据？

能回答这三个问题，就已经抓住了项目最常见的维护路径。

## 新功能应该改哪里

| 任务 | 优先看哪里 |
| --- | --- |
| 新增或修改 HTTP 接口 | `internal/transport/http/router.go` 和对应模块的 `handler` |
| 修改业务规则 | 对应模块的 `service` |
| 修改数据库读写 | 对应模块的 `repository` |
| 新增表或改表 | 新增 `internal/migrations` 迁移 |
| 修改配置字段 | `internal/config`、`configs/*.example.yaml` 和配置文档 |
| 修改通用能力 | `pkg`，但要确认它不依赖 `internal` |

维护时优先照着 Demo 模块的形态写：`model -> repository -> service -> handler`。不要把业务规则直接塞进 handler。

## Go 新手先补这些

先补够读项目的知识即可：

- `struct`、方法和接口；
- 包路径、导入和 `internal` 目录规则；
- `context.Context`；
- `error` 和 `fmt.Errorf("...: %w", err)`；
- `defer`；
- GORM 的基础 CRUD 和事务；
- HTTP handler、service、repository 分层。

不要等完全学完 Go 再维护项目。更有效的练习是给 Demo Todo 加一个很小的校验或字段，然后补测试并跑通。

## 接手维护检查表

开始改代码前：

- 先确认需求属于哪个模块；
- 先看对应模块已有测试；
- 数据库结构变更优先新增迁移，不改已共享迁移；
- 配置变更优先更新示例配置和文档；
- 只改当前任务需要的文件，不顺手重构无关代码。

改完后按影响范围验证：

```powershell
go test ./... -count=1 -mod=readonly
go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main
```

如果只改文档，可以不跑完整测试，但要检查链接路径、命令和事实是否仍然准确。
