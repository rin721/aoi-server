# DB CLI 工作流

`db` 命令提供基于 sqlgen 的数据库 DDL、Demo Todo 操作和 goose 迁移执行。

该命令由 `cmd/main` 声明为 `cli.CommandSpec`，通过 `pkg/cli` 的 Cobra 封装解析参数。无参数启动应用时，内置 TUI 首页会展示 `db` 命令并允许查看它的帮助，但不会自动执行数据库操作。

## 示例

```bash
go run ./cmd/main db --operation=schema
go run ./cmd/main db --operation=schema --apply
go run ./cmd/main db --operation=todo-create --title="交付文档"
go run ./cmd/main db --operation=todo-list
go run ./cmd/main db migrate status --config=configs/config.yaml
go run ./cmd/main db migrate up --config=configs/config.yaml
go run ./cmd/main db migrate down --config=configs/config.yaml
```

## 范围

`db --operation=*` 聚焦 Demo Todo 表结构和 CRUD 操作。`db migrate *` 通过 `pkg/migrator` 执行 `internal/migrations` 中的 goose 迁移，用于创建 IAM 表结构等版本化数据库变更。

## 维护提示

- 命令声明保持在 `cmd/main`，使用 `cli.CommandSpec`/`cli.FlagSpec` 描述命令和 flag。
- CLI 路由、flag 解析、help 和无参数 TUI 首页由 `pkg/cli` 统一封装。
- SQL 生成和执行行为保持在 `internal/app/dbapp`。
- 迁移执行行为保持在 `pkg/migrator`，通过 `pkg/database.SQLDB()` 获取标准 SQL 连接。
- flag 行为的测试放在命令层附近，SQL 行为的测试放在 `dbapp` 附近。

IAM 初始化命令见 [IAM CLI 工作流](iam-cli.md)。
