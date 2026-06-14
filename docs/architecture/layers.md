# 分层架构

服务采用面向装配根的分层架构。`pkg` 提供可复用基础能力，`internal/app` 负责初始化和生命周期编排，业务模块通过 `internal/ports` 中的端口获取能力，不直接依赖 `pkg` 实现。

```text
cmd/main
  -> internal/app
      -> internal/config
      -> internal/app/adapters
          -> pkg 基础设施
      -> internal/modules
      -> internal/middleware
      -> internal/transport/http
      -> internal/transport/rpc
```

## 分层职责

| 层 | 职责 |
| --- | --- |
| `cmd/main` | 声明命令规格、桥接执行函数和进程入口，保持轻量。 |
| `pkg` | 封装数据库、缓存、日志、配置加载、存储、HTTP/RPC server、token、RBAC、TOTP、主机指标等基础能力。 |
| `internal/ports` | 定义业务层、middleware 和 transport 可见的最小接口与内部错误。 |
| `internal/app` | 应用启动、初始化、生命周期、reload、模块装配和基础设施适配。 |
| `internal/app/adapters` | 把 `pkg` 实现转换为 `internal/ports` 端口。 |
| `internal/config` | 配置结构、加载、校验、环境覆盖和持久化转换。 |
| `internal/modules` | 业务模块，按 `model -> repository -> service -> handler` 组织。 |
| `internal/middleware` | HTTP 链路中间件，只依赖端口和业务契约。 |
| `internal/transport` | HTTP/RPC 路由注册、请求绑定和响应转换，不创建基础设施实例。 |

## 依赖方向

允许直接依赖 `pkg` 的 internal 生产代码只有 `internal/app/**` 与 `internal/config/**`：

- `internal/app` 是组合根，可以创建数据库、日志、存储、HTTP engine、RPC registry、token manager、RBAC enforcer、TOTP 和 host metrics collector。
- `internal/config` 负责配置加载和配置类型转换，可以使用配置相关的 `pkg` 类型。
- `internal/modules`、`internal/middleware`、`internal/transport` 不直接导入 `pkg`，需要能力时依赖 `internal/ports` 或模块内接口。
- `pkg` 不导入 `internal/app` 或 `internal/modules`，保持基础设施层可复用。

边界由 `internal/import_boundary_test.go` 守护：除 `internal/app/**`、`internal/config/**` 外，internal 生产代码不得导入 `github.com/rei0721/go-scaffold/pkg/`。

## 装配顺序

`internal/app/initapp` 按以下顺序构建应用：

1. 核心服务：配置、日志、国际化、ID 生成器。
2. 基础设施：数据库、缓存、执行器、存储。
3. 可选迁移：按配置执行 goose 迁移和 demo schema。
4. 模块：Demo、IAM、Plugins、System。
5. 传输层：创建 `pkg/web` engine 和 `pkg/rpcserver` registry，通过 adapters 注入 transport。

重载和关闭由应用层统一编排，不分散到业务模块内部。

## 测试约定

业务测试需要真实基础设施时，通过 `internal/app/testsupport` 获取端口化依赖，例如数据库、HTTP router、token manager 和 RBAC enforcer。测试代码可以验证真实集成路径，但不应在业务包中复制生产初始化逻辑。
