# 分层架构

服务采用务实的分层架构。`internal/app` 是装配根，只在这里把配置、基础设施、模块和传输层连接起来。

```text
cmd/main
  -> internal/app
      -> internal/config
      -> pkg 基础设施
      -> internal/modules
      -> internal/transport/http
      -> internal/transport/rpc
```

## 分层职责

| 层 | 职责 |
| --- | --- |
| CLI | 声明命令规格、解析命令参数、提供无参数 TUI 首页、选择配置路径、处理进程退出 |
| 应用装配 | 创建核心服务、基础设施、模块、传输层和生命周期 |
| 配置 | 加载、校验、监听并暴露当前配置 |
| 基础设施 | 数据库、缓存、日志、执行器、存储、HTTP 服务、RPC 服务、主机/进程探针、SQL 生成 |
| 模块 | 业务行为和模块内校验 |
| 传输层 | HTTP/RPC 路由、中间件、请求绑定和响应转换 |

## 依赖方向

应用模块可以使用基础设施接口和包。可复用的 `pkg` 包不应反向导入应用模块。`cmd/main` 应保持轻薄，只声明 `CommandSpec`、桥接执行函数和处理进程边界，不承载业务逻辑。

`types/result` 当前依赖 Gin，因此 `types` 还不是完全与传输层无关的领域层。

## 模块形态

当前 Demo、IAM 和 System 模块主要使用以下结构：

```text
model -> repository -> service -> handler
```

- model：持久化结构和表元数据；
- repository：只负责数据库访问；
- service：负责校验、事务和业务规则；
- handler：负责 HTTP 绑定、状态码和响应转换。

## 装配顺序

`internal/app/initapp` 按以下顺序构建应用：

1. 核心服务：配置、日志、国际化、ID 生成器；
2. 基础设施：数据库、缓存、执行器、存储；
3. 可选迁移：当 `migration.auto_apply=true` 时执行 goose 迁移；
4. 模块：Demo、IAM、Plugins、System；
5. 传输层：HTTP 路由、WebUI 静态挂载、RPC 方法注册、HTTP 服务和可选 RPC 服务。

重载和关闭由应用层包统一编排，不分散到各业务模块内部。

## pkg 封装边界

IAM 相关底层库通过 `pkg` 包隔离：

- `pkg/token` 封装 JWT 签发、校验和 refresh token 哈希；
- `pkg/authorization` 封装 Casbin domain RBAC；
- `pkg/mfa` 封装 TOTP 生成和校验；
- `pkg/migrator` 封装 goose 迁移执行；
- `pkg/rpcserver` 封装 JSON-RPC 2.0 单请求入口和方法注册；
- `pkg/hostmetrics`、`pkg/processx` 封装主机采样和进程状态探针；
- `pkg/database` 只向迁移层暴露标准 `*sql.DB`，业务模块仍通过 repository 使用数据库执行器。
