# 新增模块

新增应用能力时，优先放在 `internal/modules/<name>`，再通过 `internal/app/initapp` 装配。模块代码不得直接导入 `pkg`，需要数据库、日志、ID、HTTP、存储、token、RBAC、TOTP、主机指标等能力时，通过 `internal/ports`、模块内接口或构造函数注入。

## 推荐形态

```text
model -> repository -> service -> handler
```

只使用实际需要的层。很小的模块起步时不一定需要每个文件都齐全，但依赖方向应保持稳定：handler 调 service，service 调 repository 或端口接口，repository 使用端口化数据库执行器。

## 接线步骤

1. 在 `internal/modules/<name>` 下新增模块代码。
2. 如需配置，在 `internal/config` 中新增配置结构和校验。
3. 如需新的基础能力，在 `internal/ports` 定义最小接口，并在 `internal/app/adapters` 适配 `pkg` 实现。
4. 在 `internal/app/initapp` 创建基础设施实例、适配端口，并装配 repository、service、handler。
5. 在 `internal/transport/http` 或 `internal/transport/rpc` 注册路由或方法，注册函数只接收传入的 router/registry。
6. 为 service、handler 和路由行为补充聚焦测试；需要真实基础设施时使用 `internal/app/testsupport`。

## 边界规则

- 不要在 `internal/modules` 中导入 `pkg/database`、`pkg/logger`、`pkg/token`、`pkg/authorization`、`pkg/mfa`、`pkg/web` 等具体实现。
- 不要在模块或 transport 中自行创建数据库连接、HTTP engine、RPC registry、token manager、RBAC enforcer、TOTP provider、host metrics collector。
- 不要让 `pkg` 反向依赖业务模块。
- 插件 manifest、静态 WebUI、RPC registry、host metrics 等基础设施相关读取或初始化应放在 `internal/app`。
- 模块测试如果需要真实 SQLite、Web router 或 IAM 基础组件，应从 `internal/app/testsupport` 获取适配后的端口。

## 身份和访问控制

IAM 能力位于 `internal/modules/iam`。新增业务模块如需认证主体，应从 request context 读取 IAM middleware 写入的 Principal；如需权限控制，应通过 service 暴露的授权边界或 HTTP 权限中间件调用，不要直接依赖底层 JWT/Casbin 类型。

新增受保护 HTTP 接口时，需要为路由分配明确的权限码，例如 `report:read` 或 `report:export`，在 `internal/transport/http/router.go` 中通过权限中间件保护，并把权限纳入 System API/权限同步。

## 参考模块

Demo 模块提供两个起点：公开 Todo 适合学习最小 CRUD；客户资源示例适合学习后台菜单、IAM 权限、当前 principal、资源归属和分页列表。新增真实业务模块时，优先参考客户资源示例，把可见范围和业务校验放在 service 层，把请求绑定留在 handler 层。

System 模块提供菜单、API、字典、参数、媒体库和版本包等共享能力。其他模块如需挂接媒体资源，优先保存 `mediaAssetId` 或外链 URL，不要直接拼接 `storage_key` 路径；读取文件应通过 System service 或受保护下载 API 走 IAM 权限和 storage 安全边界。
