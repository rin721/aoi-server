# IAM 模块

`internal/modules/iam` 提供企业级本地账号与组织权限管理。模块沿用 `model -> repository -> service -> handler` 分层，底层 JWT、Casbin、goose、bcrypt 和数据库细节通过 `pkg` 或 repository 边界隔离。

## 能力范围

| 能力 | 说明 |
| --- | --- |
| 本地账号 | 邮箱/用户名全局唯一，密码使用 `pkg/crypto` bcrypt 哈希 |
| 组织租户 | access token 固定绑定单个 `orgId`，切换组织会重新签发 token |
| JWT 会话 | `pkg/token` 签发 access/refresh token，refresh token 只存 HMAC/SHA-256 hash |
| 权限 | `pkg/authorization` 封装 Casbin domain RBAC，模型为 `sub, org, obj, act` |
| 邀请 | 管理员发出邀请，用户通过 token 接受并创建账号或加入组织 |
| 找回密码 | 生成一次性 reset token，真实通知通道由 `Notifier` 适配 |
| MFA | `pkg/mfa` 封装 TOTP，密钥加密存储，登录时校验一次性 code |
| 会话撤销 | 登出、refresh 轮换和管理员撤销都会更新 `iam_sessions.revoked_at` |
| 审计 | 登录、邀请、重置密码、MFA、角色、会话等关键动作写入审计日志 |

## 表结构

IAM 迁移位于 `internal/migrations`，包含：

`iam_organizations`、`iam_users`、`iam_memberships`、`iam_roles`、`iam_permissions`、`iam_sessions`、`iam_invitations`、`iam_password_resets`、`iam_mfa_factors`、`iam_audit_logs`、`iam_casbin_rules`。

迁移默认不会随服务启动自动执行。首次使用前应显式运行：

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
| `POST /api/v1/auth/login` | 登录；MFA 开启后需要 `mfaCode` |
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
| `/api/v1/orgs`、`/api/v1/users/*`、`/api/v1/roles`、`/api/v1/permissions`、`/api/v1/sessions`、`/api/v1/audit-logs` | 管理接口，需认证和 Casbin 权限 |

## 配置

核心配置位于 `auth` 和 `migration`：

- `auth.enabled` 控制 IAM 模块是否装配；
- `auth.signing_key`、`auth.refresh_token_pepper`、`auth.mfa_secret_key` 是敏感值，生产必须从 secrets 注入；
- `auth.access_token_ttl_seconds` 和 `auth.refresh_token_ttl_seconds` 控制 token 生命周期；
- `auth.login_max_failures` 和 `auth.login_lock_minutes` 控制账号锁定；
- `migration.auto_apply` 默认 `false`，生产建议通过 CLI 显式迁移。

## 后续扩展

当前 v1 不包含 SSO/OIDC/SAML、短信/邮件 MFA、真实 SMTP 或企业消息网关。通知层已经通过 `Notifier` 接口预留适配点，后续可以替换 no-op/log 实现。
