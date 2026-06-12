# IAM 模块

`internal/modules/iam` 提供企业级本地账号与组织权限管理。模块沿用 `model -> repository -> service -> handler` 分层，底层 JWT、Casbin、goose、bcrypt 和数据库细节通过 `pkg` 或 repository 边界隔离。

## 能力范围

| 能力 | 说明 |
| --- | --- |
| 本地账号 | 邮箱/用户名全局唯一，密码使用 `pkg/crypto` bcrypt 哈希 |
| 自助开通 | 可选公开注册入口一次性创建组织、首个 owner 用户和登录会话 |
| 组织租户 | access token 固定绑定单个 `orgId`，切换组织会重新签发 token |
| JWT 会话 | `pkg/token` 签发 access/refresh token，refresh token 只存 HMAC/SHA-256 hash |
| API Token | 组织内按用户和角色签发长期或定期 Bearer token，服务端只保存 hash 和显示前缀 |
| 权限 | `pkg/authorization` 封装 Casbin domain RBAC，模型为 `sub, org, obj, act` |
| 邀请 | 管理员发出邀请，用户通过 token 接受并创建账号或加入组织 |
| 用户管理 | `/admin/users` 提供筛选分页，按用户名、显示名、邮箱、角色和成员状态查询当前组织成员 |
| 找回密码 | 生成一次性 reset token，真实通知通道由 `Notifier` 适配 |
| MFA | `pkg/mfa` 封装 TOTP，密钥加密存储，登录时校验一次性 code |
| 会话撤销 | 登出、refresh 轮换和管理员撤销都会更新 `iam_sessions.revoked_at` |
| 审计 | 登录、邀请、重置密码、MFA、角色、会话等关键动作写入审计日志 |

## 表结构

IAM 迁移位于 `internal/migrations`，包含：

`iam_organizations`、`iam_users`、`iam_memberships`、`iam_roles`、`iam_permissions`、`iam_sessions`、`iam_api_tokens`、`iam_invitations`、`iam_password_resets`、`iam_mfa_factors`、`iam_audit_logs`、`iam_casbin_rules`。

本地默认配置会在服务启动时自动应用这些 goose 迁移，首次打开 `/admin` 时可以直接进入浏览器初始化。生产配置仍建议关闭自动迁移，并在发布窗口显式运行：

```bash
go run ./cmd/main db migrate up --config=configs/config.yaml
```

## 初始管理员

创建初始组织和管理员：

```bash
go run ./cmd/main iam bootstrap-admin \
  --config=configs/config.yaml \
  --org-code=acme \
  --org-name="Acme Corp" \
  --username=admin \
  --email=admin@example.com \
  --password-stdin
```

该命令会初始化内置权限、`owner/admin/member` 角色、owner 组织成员关系和 Casbin policy。重复执行同一管理员会尽量保持幂等。

## HTTP 路由

| 路由 | 说明 |
| --- | --- |
| `GET /api/v1/auth/captcha` | 获取登录验证码开关、图片和 `captchaId`；关闭时返回 `enabled=false` |
| `POST /api/v1/auth/login` | 登录；MFA 开启后需要 `mfaCode` |
| `POST /api/v1/auth/signup` | 自助注册并创建首个组织 owner |
| `GET /api/v1/auth/setup/status` | 查询是否需要首次初始化管理员 |
| `POST /api/v1/auth/setup/initial-admin` | 空 IAM 用户表时创建首个组织 owner |
| `POST /api/v1/auth/refresh` | refresh token 轮换 |
| `POST /api/v1/auth/logout` | 撤销当前会话 |
| `POST /api/v1/auth/switch-org` | 切换组织并重新签发 token |
| `POST /api/v1/auth/password/forgot` | 发起找回密码 |
| `POST /api/v1/auth/password/reset` | 重置密码并撤销旧会话 |
| `POST /api/v1/auth/mfa/setup` | 创建 TOTP secret 和 otpauth URL |
| `POST /api/v1/auth/mfa/verify` | 校验 TOTP 并启用 MFA |
| `POST /api/v1/invitations/:token/accept` | 接受邀请 |
| `GET /api/v1/me` | 当前用户资料 |
| `GET /api/v1/me/orgs` | 当前用户组织列表 |
| `/api/v1/orgs`、`/api/v1/orgs/:orgId/users/*`、`/api/v1/orgs/:orgId/invitations/*`、`/api/v1/orgs/:orgId/roles/*`、`/api/v1/orgs/:orgId/permissions`、`/api/v1/orgs/:orgId/api-tokens`、`/api/v1/orgs/:orgId/sessions`、`/api/v1/orgs/:orgId/audit-logs` | 管理接口，需认证和 Casbin 权限 |

`GET /api/v1/orgs/:orgId/users` 返回分页对象，支持 `keyword`、`username`、
`displayName`/`nickName`、`email`、`roleCode`、`status`、`page`、`pageSize`。
本地账号模型暂不包含手机号和头像字段；新增组织成员仍通过邀请流程完成，
这样用户可以自己设置密码并接受组织角色。

## API Token 管理

API Token 用于脚本、外部系统或自动化任务调用受保护接口。管理员在后台 `/admin/api-tokens` 为组织成员选择一个已拥有的角色并签发 token；返回的完整 token 只在创建响应和成功弹窗中显示一次，之后列表只保留 `tokenPrefix`，数据库只保存 `TokenHash`。

API Token 认证仍走 `Authorization: Bearer <token>`，但不会创建 refresh 会话。通过 API Token 进入的请求会固定到 token 所属组织，并按签发时绑定的角色权限检查 `obj:act`，例如 `api_token:read`、`user:read`、`media:upload` 或 `plugin:proxy`。如果用户被禁用、成员关系失效、角色被移除、token 过期或撤销，请求都会被拒绝。

维护者需要注意：API Token 的 HMAC hash 复用 `auth.refresh_token_pepper`，轮换该 pepper 会让既有 API Token 和 refresh token 一起失效。生产环境应把 token 明文视为一次性密钥，只在安全渠道交给调用方，泄漏时通过后台或 `DELETE /api/v1/orgs/{orgId}/api-tokens/{tokenId}` 立即撤销。

## 配置

核心配置位于 `auth` 和 `migration`：

- `auth.enabled` 控制 IAM 模块是否装配；
- `auth.self_signup_enabled` 控制公开自助注册入口；
- `auth.signing_key`、`auth.refresh_token_pepper`、`auth.mfa_secret_key` 是敏感值，生产必须从 secrets 注入；
- `auth.refresh_token_pepper` 同时参与 refresh token 和 API Token 的 hash；轮换后既有 refresh token 与 API Token 都会失效；
- `auth.access_token_ttl_seconds` 和 `auth.refresh_token_ttl_seconds` 控制 token 生命周期；
- `auth.login_captcha_enabled` 和 `auth.captcha_ttl_seconds` 控制登录验证码；默认关闭，验证码短期存放在 IAM 服务内存中，不新增表结构；
- `auth.login_max_failures` 和 `auth.login_lock_minutes` 控制账号锁定；
- `auth.notification_driver` 为 `debug`、`noop` 或 `local` 时，邀请和重置密码接口会返回调试 token/link；生产应使用 `smtp` 或外部系统接管通知，避免在 API 响应中暴露 token；
- `auth.smtp` 配置 SMTP host、port、账号、发件人和 STARTTLS，用于内置邮件邀请和密码重置通知；
- `auth.password_policy` 控制账号创建、接受邀请和重置密码的最小密码要求；
- `migration.auto_apply` 本地默认 `true`，便于首次启动自动建表并进入浏览器初始化；生产建议保持 `false` 并通过 CLI 显式迁移。

## 后续扩展

当前 v1 不包含 SSO/OIDC/SAML、短信/邮件 MFA 或企业消息网关。通知层已经通过 `Notifier` 接口预留适配点，当前内置 debug/no-op 和 SMTP 实现。
