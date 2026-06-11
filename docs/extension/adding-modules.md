# 新增模块

新增应用能力时，优先放在 `internal/modules/<name>`，再通过 `internal/app/initapp` 装配。

## 推荐形态

```text
model -> repository -> service -> handler
```

只使用实际需要的层。很小的模块起步时不一定需要每个文件都齐全。

## 接线步骤

1. 在 `internal/modules/<name>` 下新增模块代码。
2. 如需配置，在 `internal/config` 中新增配置结构。
3. 在 `internal/app/initapp` 中装配 repository、service、handler。
4. 在 `internal/transport/http` 中注册路由。
5. 为 service 和路由行为补充聚焦测试。
6. 如果属于托管任务，同步更新工程文档和运行时证据。

## 身份和访问控制

当前 IAM 能力位于 `internal/modules/iam`，底层 JWT、Casbin、TOTP 和 goose 分别通过 `pkg/token`、`pkg/authorization`、`pkg/mfa`、`pkg/migrator` 封装。新增业务模块如需认证主体，应从 request context 读取 IAM middleware 写入的 Principal；如需权限控制，应通过 service 暴露的授权边界或 HTTP 权限中间件调用，不要在业务代码中直接依赖底层 JWT/Casbin 类型。
