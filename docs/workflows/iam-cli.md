# IAM CLI 工作流

`iam` 命令用于执行 IAM 初始化和运维入口。命令声明位于 `cmd/main`，实际组织、账号、角色、权限和审计行为由 `internal/modules/iam` service 处理。

## 首次初始化

首次启用 IAM 前应先应用迁移：

```bash
go run ./cmd/main db migrate up --config=configs/config.yaml
```

然后创建初始组织 owner：

```bash
go run ./cmd/main iam bootstrap-admin \
  --config=configs/config.yaml \
  --org-code=acme \
  --org-name="Acme Corp" \
  --username=admin \
  --email=admin@example.com \
  --password-stdin
```

建议在自动化脚本中通过标准输入或 secrets 管道传入密码，避免把密码写入 shell history。

## bootstrap-admin 行为

`iam bootstrap-admin` 会创建或复用目标组织和管理员用户，并确保：

- 管理员属于目标组织；
- 内置权限已写入 `iam_permissions`；
- `owner`、`admin`、`member` 系统角色已写入目标组织；
- owner/admin/member 的 Casbin policy 已写入 `iam_casbin_rules`；
- 管理员拥有目标组织的 `owner` 角色；
- 审计日志记录初始化动作。

重复执行同一组织和同一管理员时应尽量保持幂等；如果用户名、邮箱或组织 code 与已有数据冲突，数据库唯一约束仍会保护数据一致性。

## 配置要求

IAM CLI 使用和 server 相同的配置加载路径：

- `--config` 优先；
- 未传入时读取 `RIN_CONFIG_PATH`；
- 再回退到 `configs/config.yaml`。

生产环境必须注入：

- `AUTH_SIGNING_KEY` 或 `RIN_APP_AUTH_SIGNING_KEY`；
- `AUTH_REFRESH_TOKEN_PEPPER` 或 `RIN_APP_AUTH_REFRESH_TOKEN_PEPPER`；
- `AUTH_MFA_SECRET_KEY` 或 `RIN_APP_AUTH_MFA_SECRET_KEY`。

## 后续运维

当前 CLI 只提供初始化管理员和数据库迁移入口。用户邀请、角色、权限、会话撤销和审计查看由 HTTP 管理接口提供；后续如需离线运维命令，应继续保持 `cmd/main` 轻薄，把真实业务逻辑放在 IAM service。
