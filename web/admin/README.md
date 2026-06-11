# Aoi Admin WebUI

`web/admin` 是 Go 后端的 Nuxt 4 静态管理台。它消费现有 IAM、探针和 Demo Todo HTTP API，不再启动 Nuxt server/BFF 代理。

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `NUXT_PUBLIC_API_BASE_URL` | 空字符串 | 浏览器请求 Go API 的前缀。空值表示同源根路径，例如 `/api/v1`、`/health`。 |
| `NUXT_APP_BASE_URL` | `/admin/` | Nuxt 静态资源和路由 baseURL，需要和 Go 的 `webui.mount_path` 保持一致。 |

登录后 access token 和 refresh token 保存在 Pinia 内存与 `sessionStorage` 中，请求时通过 `Authorization: Bearer <accessToken>` 访问 Go API。401 会自动用 refresh token 刷新一次，失败后清理本地会话并回到登录页。

## 构建并由 Go 托管

先构建静态产物：

```powershell
cd web/admin
pnpm install
pnpm generate
```

默认产物目录是 `web/admin/.output/public`。Go 服务启动后会读取该目录，并在 `/admin` 挂载管理台：

```powershell
go run ./cmd/main server
```

打开 `http://127.0.0.1:9999/admin/login` 即可访问。若未执行 `pnpm generate` 或目录缺少 `index.html`，Go 服务仍会正常启动，只是 `/admin` 返回 404 并记录 warning。

## 本地初始化示例

```powershell
go run ./cmd/main db migrate up --config=configs/config.yaml
Get-Content .\admin-password.txt | go run ./cmd/main iam bootstrap-admin --config=configs/config.yaml --org-code=acme --org-name="Acme Corp" --username=admin --email=admin@example.com --password-stdin
go run ./cmd/main server
```

## 常用命令

```powershell
pnpm typecheck
pnpm generate
pnpm build
```

开发期如果仍使用 `pnpm dev` 单独预览前端，需要显式指定 Go API 地址：

```powershell
$env:NUXT_PUBLIC_API_BASE_URL="http://127.0.0.1:9999"
pnpm dev
```
