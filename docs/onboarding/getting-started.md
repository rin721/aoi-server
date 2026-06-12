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
13. `internal/modules/demo/handler/customer.go`
14. `internal/modules/demo/service/customer.go`
15. `internal/modules/demo/repository/customer.go`
16. `internal/modules/demo/model/customer.go`

这条路线会把你带过一遍：命令入口、应用启动、路由注册、HTTP handler、业务服务、数据库访问和模型定义。

## 本地跑起来

本地默认配置会在服务启动时自动应用 goose 迁移。启动后打开 `http://127.0.0.1:9999/admin`，如果 IAM 还没有任何用户，后台会进入首次初始化页面。

如果需要通过 CLI 初始化管理员，也可以直接运行：

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

本地默认使用 SQLite `./data/app.db`。Demo 模块默认启用，且 `demo.apply_schema_on_start=true`，所以 `demo_todos` 和 `demo_customers` 表结构会在服务启动时自动应用。IAM 表结构由 goose 管理，默认 `migration.auto_apply=true`，会随服务启动自动应用；生产或发布窗口仍建议显式执行 `db migrate status` 和 `db migrate up`。

## 在后台试一次客户资源示例

登录 `http://127.0.0.1:9999/admin` 后，进入左侧菜单的 `客户列表`。点击 `新增`，填写客户名和客户电话，保存后会生成一条带创建者、组织和角色上下文的客户记录。

如果页面提示 403，先确认当前角色拥有 `customer:read`、`customer:create`、`customer:update`、`customer:delete`。新环境中 owner 拥有通配权限，admin 会在内置权限种子里获得这些权限；已有环境可以进入 `API 管理` 同步权限，再到 `角色权限` 给角色勾选 `customer:*`。

## 在后台试一次用户管理

登录后台后进入左侧菜单的 `用户`。页面支持按关键字、用户名、显示名、邮箱、角色和成员状态筛选，并以分页对象读取
`GET /api/v1/orgs/{orgId}/users`。本项目暂不直接创建用户账号，新增成员仍走
`邀请用户`：填写邮箱和角色后生成邀请，用户通过邀请链接接受并设置自己的密码。

如果用户列表为空，先确认当前组织是否有成员；如果筛选后为空，点击 `重置`
回到第一页。API Token 签发页也复用同一个用户分页接口，并会读取前 100 个成员
作为候选用户。

## 在后台试一次组织管理

登录后台后进入左侧菜单的 `组织`。页面会以分页对象读取 `GET /api/v1/orgs`，
支持按关键字、组织 Code、组织名称和状态筛选。点击列表中的 `切换` 会重新签发
当前访问上下文；右侧 `当前组织` 面板只允许修改当前登录令牌绑定的组织名称。

新建组织会自动把当前用户加入新组织并授予 owner 角色。创建后如需切换到新组织，
可在组织列表或顶部组织切换器中选择。

## 在后台试一次 API Token

登录 `http://127.0.0.1:9999/admin` 后，进入左侧菜单的 `API Token`。先确认当前用户已经在组织里拥有一个角色，然后点击 `签发`，选择用户、角色和有效期。创建成功后完整 token 只显示一次，复制后可以这样测试：

```powershell
curl -H "Authorization: Bearer <api-token>" http://127.0.0.1:9999/api/v1/me
```

如果页面提示 `iam_api_tokens` 表不可用，说明数据库迁移还没有应用；本地通常重启服务即可自动应用，生产环境应显式执行 `go run ./cmd/main db migrate up --config=configs/config.yaml`。

## 在后台试一次媒体库

登录 `http://127.0.0.1:9999/admin` 后，进入左侧菜单的 `媒体库`。如果只想体验外链导入，可以点击 `导入URL`，输入 `我的图片|https://example.com/a.png` 这样的多行文本；系统只保存链接，不会下载远程文件。

如果要体验普通上传，请在示例配置或本地配置中启用：

```yaml
storage:
  enabled: true
  fs_type: basepath
  base_path: ./data
```

重启服务并应用迁移后，上传文件会写入 `data/media/...`。如果页面提示 `system_media_assets` 表不可用，先执行数据库迁移；如果提示对象存储不可用，检查 `storage.enabled`。

## 在后台试一次断点上传

断点上传入口在 `媒体库` 附近的 `断点上传`。它适合体验大文件分片流程，本项目会用 SHA-256 校验整文件和每个分片，并把完成后的文件写回媒体库。

试用步骤：

1. 确认已经启用 `storage.enabled=true`，并应用 `20260612000600_create_system_media.sql` 与 `20260612000700_create_system_media_resumable_uploads.sql` 迁移。
2. 登录后台，进入 `断点上传`，点击 `选择文件`。
3. 页面会先计算文件哈希并创建上传会话，再点击 `上传文件`。
4. 完成后点击 `查看媒体库`，确认新资产已经出现在媒体列表。

如果中途关闭页面，再次选择同一个文件时，服务端会返回已存在的分片序号，页面会从缺失分片继续。临时分片位于 `data/media/chunks/...`，完成或中止会话时会自动清理。

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

客户资源示例的链路类似，但会多一步读取登录 principal，并在 service 层限制资源可见范围：

```text
internal/transport/http/router.go
  -> middleware.Auth / RequirePermission(customer:*)
  -> internal/modules/demo/handler/customer.go
  -> internal/modules/demo/service/customer.go
  -> internal/modules/demo/repository/customer.go
  -> internal/modules/demo/model/customer.go
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

不要等完全学完 Go 再维护项目。更有效的练习是给 Demo Todo 加一个很小的校验或字段，或者给客户资源示例加一个筛选字段，然后补测试并跑通。

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
- 版本管理发布包：后台 `版本管理` 页面可以把菜单、API 和字典打成 JSON 发布包。它不是 Go 构建版本，也不是 goose 迁移版本；导入时当前只会幂等补齐字典，菜单和 API 会保留在包记录中并报告跳过。页面提示 `system_versions` 表不可用时，先执行数据库迁移。
- 媒体库：后台 `媒体库` 页面可以管理分类、上传本地文件、导入外链、重命名、下载和删除。普通上传依赖 storage，外链导入不依赖对象存储。
