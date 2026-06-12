# HTTP API 文档

本文档面向开发者阅读，概述当前服务暴露的 HTTP API。机器可读的完整契约见
`docs/api/openapi.yaml`。

## 通用约定

默认本地服务地址：

```text
http://127.0.0.1:9999
```

除特别说明外，响应统一包裹在 `Result` 结构中：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | integer | 业务错误码，`0` 表示成功。 |
| `message` | string | 响应消息。 |
| `data` | any | 响应数据，错误时通常为空。 |
| `traceId` | string | 请求追踪 ID，错误响应中常见。 |
| `serverTime` | integer | 服务端 Unix 秒级时间戳。 |

受保护的 IAM 接口需要请求头：

```http
Authorization: Bearer <accessToken>
```

## 探针接口

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/health` | 否 | 存活检查，只说明进程和路由可响应。 |
| GET | `/ready` | 否 | 就绪检查，会检查数据库依赖。 |

## 演示 Todo

演示 Todo 接口不要求认证，主要用于验证模块分层、路由和统一响应格式。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/demo/todos` | 创建 Todo。 |
| GET | `/api/v1/demo/todos` | 查询 Todo 列表。 |
| GET | `/api/v1/demo/todos/{id}` | 查询单个 Todo。 |
| PUT | `/api/v1/demo/todos/{id}` | 更新 Todo。 |
| DELETE | `/api/v1/demo/todos/{id}` | 删除 Todo。 |

### 创建 Todo

```http
POST /api/v1/demo/todos
Content-Type: application/json
```

```json
{
  "title": "Router integration",
  "description": "through handler and service",
  "completed": false
}
```

必填字段：`title`。

### 更新 Todo

```http
PUT /api/v1/demo/todos/1
Content-Type: application/json
```

```json
{
  "title": "Updated through router",
  "completed": true
}
```

`title`、`description`、`completed` 都是可选字段，只更新请求体中出现的字段。

## 插件接口

插件接口需要 IAM access token，并通过 Casbin 的 `plugin:read` 或 `plugin:proxy` 权限控制。

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/v1/plugins` | `plugin:read` | 列出已安装插件 manifest，菜单会按当前用户权限过滤。 |
| GET | `/api/v1/plugins/{id}` | `plugin:read` | 读取单个插件 manifest。 |
| GET | `/api/v1/plugins/{id}/health` | `plugin:read` | 检查 sidecar 插件健康状态。 |
| ANY | `/api/v1/plugins/{id}/proxy/*path` | `plugin:proxy` | 代理访问插件 sidecar API。 |

插件代理会向 sidecar 注入 `X-Aoi-Plugin-ID`、`X-Aoi-User-ID`、`X-Aoi-Org-ID`、`X-Aoi-Trace-ID`、`X-Aoi-Signature-Timestamp` 和 `X-Aoi-Signature` 请求头。签名密钥由 manifest 的 `secretRef` 指向运行时环境变量。

## 系统接口

系统接口需要 IAM access token。菜单接口按当前用户权限过滤，API 目录接口需要 `permission:read`。

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/v1/system/menus` | 登录可见 | 返回当前用户可见的后台菜单分组。 |
| GET | `/api/v1/system/config` | `config:read` | 返回当前进程脱敏后的运行配置快照，敏感字段只显示是否已配置。 |
| GET | `/api/v1/system/server-info` | `server:read` | 返回主机 CPU/RAM/磁盘以及当前后端进程的运行时、内存、GC 和构建信息快照。 |
| GET | `/api/v1/system/apis` | `permission:read` | 返回当前进程真实注册的 HTTP API 目录。 |
| POST | `/api/v1/system/apis/sync` | `permission:read` | 同步当前进程 HTTP API 目录到 `system_apis` 表；表未迁移时只刷新目录并返回未持久化状态。 |
| POST | `/api/v1/system/apis/permissions/sync` | `permission:sync` | 从当前 API 目录提取权限码并补齐 `iam_permissions` 字典，便于角色授权页直接勾选。 |
| GET | `/api/v1/system/operation-records` | `operation:read` | 分页查询后台受保护接口的操作记录，支持请求方法、路径和状态码筛选。 |
| DELETE | `/api/v1/system/operation-records` | `operation:delete` | 按 ID 批量删除操作记录。 |
| GET | `/api/v1/system/parameters` | `parameter:read` | 分页查询系统参数，支持创建日期、参数名称和参数键筛选。 |
| POST | `/api/v1/system/parameters` | `parameter:create` | 创建系统参数。 |
| DELETE | `/api/v1/system/parameters` | `parameter:delete` | 按 ID 批量软删除系统参数。 |
| GET | `/api/v1/system/parameters/value` | `parameter:read` | 按参数键读取系统参数。 |
| GET | `/api/v1/system/parameters/{parameterId}` | `parameter:read` | 按 ID 读取系统参数。 |
| PATCH | `/api/v1/system/parameters/{parameterId}` | `parameter:update` | 更新系统参数名称、键、值或说明。 |
| DELETE | `/api/v1/system/parameters/{parameterId}` | `parameter:delete` | 软删除系统参数。 |
| GET | `/api/v1/system/dictionaries` | `dictionary:read` | 返回系统字典目录和字典项，表未就绪时返回不可用状态。 |
| POST | `/api/v1/system/dictionaries` | `dictionary:create` | 创建系统字典。 |
| PATCH | `/api/v1/system/dictionaries/{dictionaryId}` | `dictionary:update` | 更新系统字典编码、名称、说明或状态。 |
| DELETE | `/api/v1/system/dictionaries/{dictionaryId}` | `dictionary:delete` | 软删除系统字典及其字典项。 |
| POST | `/api/v1/system/dictionaries/{dictionaryId}/items` | `dictionary:update` | 创建字典项。 |
| PATCH | `/api/v1/system/dictionary-items/{itemId}` | `dictionary:update` | 更新字典项标签、值、扩展信息、排序或状态。 |
| DELETE | `/api/v1/system/dictionary-items/{itemId}` | `dictionary:delete` | 软删除字典项。 |

## IAM 公开接口

这些接口不要求 access token。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/auth/setup/status` | 查询是否需要首次初始化管理员。 |
| POST | `/api/v1/auth/setup/initial-admin` | 空 IAM 用户表时创建首个组织 owner，并返回登录令牌。 |
| POST | `/api/v1/auth/signup` | 自助注册，创建组织、首个 owner 和登录令牌。 |
| GET | `/api/v1/auth/captcha` | 获取登录验证码开关、图片和 `captchaId`；验证码关闭时返回 `enabled=false`。 |
| POST | `/api/v1/auth/login` | 登录并签发 access token 与 refresh token。 |
| POST | `/api/v1/auth/refresh` | 使用 refresh token 刷新令牌。 |
| POST | `/api/v1/auth/password/forgot` | 创建密码重置通知；debug/noop/local 通知驱动会返回调试 token/link，smtp 不返回 token。 |
| POST | `/api/v1/auth/password/reset` | 使用重置令牌重置密码。 |
| POST | `/api/v1/invitations/{token}/accept` | 接受组织邀请。 |

### 登录

```json
{
  "identifier": "admin@example.com",
  "password": "secret",
  "orgCode": "acme",
  "captchaId": "captcha-id",
  "captchaCode": "7",
  "mfaCode": "123456"
}
```

必填字段：`identifier`、`password`。

`orgCode` 可用于指定登录组织；开启登录验证码后，先调用 `GET /api/v1/auth/captcha` 并在登录时提交 `captchaId` 与 `captchaCode`；开启 MFA 后需要 `mfaCode`。

成功响应的 `data` 为：

| 字段 | 说明 |
| --- | --- |
| `accessToken` | access token。 |
| `accessExpiresAt` | access token 过期时间。 |
| `refreshToken` | refresh token。 |
| `refreshExpiresAt` | refresh token 过期时间。 |

### 刷新令牌

```json
{
  "refreshToken": "<refreshToken>"
}
```

### 找回密码

```json
{
  "email": "admin@example.com"
}
```

`auth.notification_driver=debug/noop/local` 时响应会包含调试 `token` 和 `url`；`smtp` 或外部通知模式不会在响应中暴露一次性 token。

### 重置密码

```json
{
  "token": "<resetToken>",
  "newPassword": "new-secret"
}
```

### 接受邀请

```http
POST /api/v1/invitations/<token>/accept
Content-Type: application/json
```

```json
{
  "username": "member",
  "displayName": "Member",
  "password": "secret"
}
```

必填字段：`username`、`password`。

## IAM 账号接口

以下接口都需要 `Authorization: Bearer <accessToken>`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/auth/logout` | 撤销当前会话。 |
| POST | `/api/v1/auth/switch-org` | 切换当前组织并签发新令牌。 |
| POST | `/api/v1/auth/mfa/setup` | 创建或轮换 TOTP MFA 密钥。 |
| POST | `/api/v1/auth/mfa/verify` | 验证并启用 TOTP MFA。 |
| GET | `/api/v1/me` | 查询当前用户资料。 |
| GET | `/api/v1/me/orgs` | 查询当前用户所属组织。 |

### 切换组织

```json
{
  "orgId": 10001
}
```

### 验证 MFA

```json
{
  "code": "123456"
}
```

## IAM 组织管理接口

以下接口都需要认证，并根据路由要求检查 Casbin 权限。

| 方法 | 路径 | 权限对象/动作 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/v1/orgs` | `org:read` | 查询组织列表。 |
| POST | `/api/v1/orgs` | `org:create` | 创建组织，并把当前用户设为新组织 owner。 |
| PATCH | `/api/v1/orgs/{orgId}` | `org:update` | 更新当前组织信息。 |
| GET | `/api/v1/orgs/{orgId}/users` | `user:read` | 查询当前组织用户。 |
| PATCH | `/api/v1/orgs/{orgId}/users/{userId}` | `user:update` | 更新成员状态或角色。 |
| POST | `/api/v1/orgs/{orgId}/users/invitations` | `user:invite` | 邀请用户加入当前组织。 |
| GET | `/api/v1/orgs/{orgId}/invitations` | `user:invite` | 查询当前组织邀请。 |
| DELETE | `/api/v1/orgs/{orgId}/invitations/{invitationId}` | `user:invite` | 撤销待处理邀请。 |
| GET | `/api/v1/orgs/{orgId}/roles` | `role:read` | 查询当前组织角色。 |
| POST | `/api/v1/orgs/{orgId}/roles` | `role:create` | 在当前组织创建角色。 |
| PATCH | `/api/v1/orgs/{orgId}/roles/{roleId}` | `role:update` | 更新自定义角色。 |
| GET | `/api/v1/orgs/{orgId}/permissions` | `permission:read` | 查询可用权限。 |

路径中的 `{orgId}` 必须与 access token 中的 `orgId` 一致。

### 创建组织

```json
{
  "code": "acme",
  "name": "Acme Corp"
}
```

### 邀请用户

```json
{
  "email": "member@example.com",
  "roleCode": "member"
}
```

当前 no-op 通知器会在响应中直接返回邀请 `token`。

### 创建角色

```json
{
  "code": "operator",
  "name": "Operator",
  "description": "Daily operator",
  "permissions": ["user:read", "session:read"]
}
```

## IAM 会话接口

| 方法 | 路径 | 权限对象/动作 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/v1/orgs/{orgId}/sessions` | `session:read` | 查询当前用户或指定用户的会话。 |
| DELETE | `/api/v1/orgs/{orgId}/sessions/{sessionId}` | `session:revoke` | 撤销当前组织中的会话。 |

查询会话时可以传入可选查询参数：

```http
GET /api/v1/orgs/10001/sessions?userId=10002
```

未传 `userId` 时查询当前用户的会话。

## IAM 审计接口

| 方法 | 路径 | 权限对象/动作 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/v1/orgs/{orgId}/audit-logs` | `audit:read` | 查询当前组织审计日志。 |

可选查询参数：

```http
GET /api/v1/orgs/10001/audit-logs?action=auth.login&userId=10002&limit=100&cursor=90001
```

`limit` 默认值为 `100`。

## 常见错误

| HTTP 状态码 | 错误码 | 说明 |
| --- | --- | --- |
| 400 | `1000` | 请求参数无效。 |
| 401 | `3000` | 未认证、登录失败或令牌无效。 |
| 403 | `3003` | 权限不足或组织不匹配。 |
| 404 | `4000` | 资源不存在。 |
| 500 | `5000` | 服务端内部错误。 |
| 503 | `5001` | 服务未就绪，常见于数据库不可用。 |
