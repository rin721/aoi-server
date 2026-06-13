# Aoi Admin Web

`web/admin` 是当前 Go 后端的 Nuxt 4 后台管理台主线。它复用 Aoi UI tokens、Material Web wrapper 和交互基础，并由 Go 服务在 `/admin` 下静态托管。

## 当前定位

- 后台能力来自 Go HTTP API：IAM 登录、组织、用户、角色权限、API Token、会话、安全、登录日志、审计日志、System 管理、媒体库、版本、服务状态和可选 Demo Todo。
- UI/信息架构参考 Gin-Vue-Admin 的后台壳层、访问标签、设置抽屉、筛选工具条、表格页和仪表盘布局。
- 不实现 Gin-Vue-Admin 的编程辅助、代码生成、插件市场或插件安装/打包系统；已有 Plugins 后端只提供 manifest、健康检查和代理。
- 本地 UI 偏好只写入浏览器 `localStorage` 的 `aoi.admin.ui.v1`；访问标签写入 `sessionStorage` 的 `aoi.admin.visited-tabs.v1`。

## 常用命令

```bash
pnpm install
pnpm dev
pnpm typecheck
pnpm build
pnpm generate
```

如果 Nuxt dev server 和 Go API 分别运行在不同端口，本地开发时需要显式指定后端地址：

```powershell
$env:NUXT_PUBLIC_API_BASE_URL="http://127.0.0.1:9999"
pnpm dev --host 127.0.0.1 --port 3001
```

`pnpm build` 用于 Nuxt 构建检查。Go 后端静态托管需要 `pnpm generate` 生成的产物：

```text
web/admin/.output/public
```

生成后启动 Go 服务，访问：

```text
http://127.0.0.1:9999/admin/login
```

## 当前后台页面

- 工作台：仪表盘、组织、用户、角色权限。
- 安全审计：会话、登录日志、审计日志、错误日志。
- 系统管理：菜单管理、API 管理、字典管理、操作历史、参数管理、系统配置、服务状态。
  API 管理会展示公开、登录后、权限三类访问属性，用于对齐 GVA 的公开路由与 JWT/Casbin 保护路由分层。
- 登录页会在后端 `GET /api/v1/auth/captcha` 返回 `enabled=true` 时显示验证码输入和刷新图片；默认配置关闭验证码。
- 示例：Demo Todo，仅在 `NUXT_PUBLIC_SHOW_DEMO_TODO=true` 或后端 demo 开关开启时显示。

## 运行时配置

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `NUXT_APP_BASE_URL` | `/admin/` | Nuxt 静态资源和路由 baseURL，需要和 Go `webui.mount_path` 对齐。 |
| `NUXT_PUBLIC_API_BASE_URL` | 空字符串 | 管理台 API 基础路径；空值表示同源调用 Go API。 |
| `NUXT_PUBLIC_SHOW_DEMO_TODO` | `false` | 是否显示 Demo Todo 入口。 |

Go 配置侧默认值：

```yaml
webui:
  enabled: true
  mount_path: /admin
  dist_dir: ./web/admin/.output/public
  public_base_url: /admin
```

## 开发约束

- 业务页面统一通过 `useAdminApi()` 调用后端，并保持 `/api/v1` 契约不变；`useAoiApi()` 和 `server/api/mock` 只服务遗留 Aoi/component demo，不作为后台业务入口。
- 后台响应类型维护在 `app/types/admin.ts`。
- 业务页面不要直接使用 `md-*`；需要 Material Web 能力时，先通过 `app/components/aoi/` 暴露。
- 可见前端变更必须用 Browser 做桌面和移动端视觉检查；会影响后台工作流的后端变更也必须检查对应前端路由。最小视口为 `1440x900` 和 `390x844`，交付说明需要记录检查路线、视口和残余风险。
- 后台页面只接入当前 Go 后端已经暴露的模型。菜单管理、API 管理、字典、参数、系统配置、版本和媒体库已有 System API；代码生成、插件市场、插件安装/打包等未暴露能力不要在前端先行造模型。

更多长期约束见 `design/rules.md`。
