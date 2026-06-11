# pkg/authorization - 权限封装

`pkg/authorization` 是项目的权限防腐层。它把 Casbin 收敛在 `pkg` 内部，向业务层暴露项目自有的 `Enforcer` 和 `Rule`。

## API 分类

- 定位：[CONFIRMED] 公共基础设施 API。
- 稳定边界：`Enforcer` 接口、`Rule`、`New`、domain RBAC 行为。
- 当前风险：[RISK] policy 持久化由调用方负责，本包只维护内存 enforcer。
- 非目标：[CONFIRMED] 本包不直接访问数据库、不定义业务角色、不决定 HTTP 状态码。

## 权限模型

当前模型是组织域 RBAC：

```text
request: sub, org, obj, act
policy:  sub, org, obj, act
role:    user, role, org
```

匹配规则：

- 用户必须在同一 `org` 中拥有目标角色；
- `obj` 支持 `*` 和 `keyMatch2`；
- `act` 支持 `*` 和正则匹配；
- `owner`、`admin`、`member` 等角色语义由 IAM service 写入 policy 决定。

## 基本用法

```go
enforcer, err := authorization.New()
if err != nil {
    return err
}

_, _ = enforcer.AddPolicy(ctx, "role:admin", "2001", "user", "read")
_, _ = enforcer.AddRoleForUser(ctx, "user:1001", "role:admin", "2001")

allowed, err := enforcer.Enforce(ctx, "user:1001", "2001", "user", "read")
```

## 持久化规则

IAM 通过 `iam_casbin_rules` 表持久化规则，并在启动或策略变化后调用 `LoadRules` 重载内存 enforcer。其他模块如需权限控制，应走 IAM service 或 HTTP 权限中间件，不要直接持有 Casbin 对象。
