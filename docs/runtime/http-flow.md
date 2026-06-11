# HTTP 流程

HTTP 路由位于 `internal/transport/http`。路由器在应用启动期间创建，此时模块和基础设施已经就绪。

## 中间件顺序

1. i18n，可用时启用；
2. 请求 trace ID；
3. CORS；
4. 请求日志；
5. panic recovery。

传输层中间件只处理 HTTP 关注点。业务决策应放在模块 service 中。

## 路由组

| 路由 | Handler 来源 |
| --- | --- |
| `GET /health` | transport router |
| `GET /ready` | transport router，包含数据库 ping |
| `/api/v1/demo/todos` | demo handler |
| `/api/v1/auth/*` | IAM auth handler |
| `/api/v1/me`, `/api/v1/me/orgs` | IAM profile handler |
| `/api/v1/orgs`, `/api/v1/users/*`, `/api/v1/roles`, `/api/v1/permissions`, `/api/v1/sessions`, `/api/v1/audit-logs` | IAM admin handler，需认证和权限校验 |

IAM 路由只在 `auth.enabled=true` 且模块装配成功时注册。受保护路由先经过 Bearer access token 校验，再按 `obj/act` 调用 Casbin domain RBAC。

## 请求形态

```text
HTTP request
  -> middleware
  -> handler bind/parse
  -> service validation/business rules
  -> repository/database or infrastructure package
  -> service result
  -> handler result helper
  -> JSON response
```

handler 不应隐藏事务或业务规则。service 负责业务校验和事务边界。
