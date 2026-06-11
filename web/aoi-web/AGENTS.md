# Agent 规则

## 项目概览

- 当前项目是 `aoi-web`，Go scaffold 的 Nuxt 4 后台管理台主线，使用 Vue 3、TypeScript、Pinia、`@nuxt/icon`，并通过本地 Aoi wrapper 封装 Material Web。
- 包管理器只使用 pnpm。仓库声明的版本是 `pnpm@10.22.0`。
- 应用以 Go 后端 IAM、探针和可选 Demo API 为数据源，不在 Nuxt 内新增生产后端能力。
- `web/admin` 是旧版后台实现，仅作为迁移回滚和对照参考保留。
- 较大的产品、架构、UI、API 或交互变更，需要优先参考 `design/rules.md`。

## 常用命令

- 安装依赖：`pnpm install`
- 启动开发服务：`pnpm dev`
- 类型检查：`pnpm typecheck`
- Nuxt 构建检查：`pnpm build`
- 生成 Go 静态托管产物：`pnpm generate`

Go 后端默认托管 `web/aoi-web/.output/public` 到 `/admin`。如果要验证 Go 静态挂载，必须先运行 `pnpm generate`，确认 `.output/public/index.html` 存在。

当前仓库还没有提交 `lint` 脚本。除非后续新增或用户明确提供 lint 命令，否则不要声称已经完成 lint 验证。

## 仓库边界

- 前端应用代码位于 `app/`。
- 后台 API DTO 放在 `app/types/admin.ts`。
- 旧 mock、shared、content 目录可以作为 Aoi 框架遗留资产保留；新增后台功能优先接 Go API。
- 本地化文案位于 `i18n/locales/`。
- 长期设计、技术、API 和交互约束位于 `design/rules.md`；`design/` 不保留临时研究、原型或阶段计划。
- 不要编辑 `.nuxt/`、`.output/`、`node_modules/` 等生成目录或依赖目录。
- 内部连接和导入应优先使用相对路径或 Nuxt 自动导入，避免引入不必要的全局工具或模块。

## 代码风格

- 使用 TypeScript 和 Vue 3 Composition API。
- 遵循现有格式：2 空格缩进、双引号、LF 换行、Vue/TS 文件不加分号。
- 优先使用 Nuxt 自动导入和本地 composable，不随意新增全局工具。
- 变更保持聚焦，避免无关重构。
- 搜索仓库内容时优先使用 `rg` 或 `rg --files`。

## UI 与组件规则

- 业务页面和功能组件不要直接使用 `md-*` Material Web 元素。
- Material Web 的导入集中在 `app/plugins/material-web.client.ts`。
- 如需暴露新的 Material Web 行为，先在 `app/components/aoi/` 新增或扩展 Aoi wrapper。
- 优先使用已有 Aoi 组件，例如 `AoiLink`、`AoiButton`、`AoiIconButton`、`AoiTextField`、`AoiSelect`、`AoiTabs`、`AoiCheckbox`、`AoiDialog`、`AoiMenu`、`AoiProgress`。
- 普通文本链接、卡片链接、标签链接和导航链接统一使用 `AoiLink`；业务代码不要直接使用 `NuxtLink` 或裸 `<a>`。
- 按钮式导航继续使用 `AoiButton` 或 `AoiIconButton`，它们的 `to`/`href` 会委托给 `AoiLink`。
- 使用 `app/assets/css/tokens.css` 中的 CSS 变量，以及 `app/assets/css/main.css` 中的共享布局规则。
- 保持响应式行为、可访问标签、键盘焦点、触控尺寸和 reduced-motion 支持。
- 图标优先通过 `@nuxt/icon` 使用本地 Lucide 集合，避免远程图标依赖。

## 状态、API 与数据规则

- 后台 API 访问统一走 `useAdminApi()`，并保持 `/api/v1` 后端契约不变。
- 不新增当前 Go 后端没有暴露的菜单管理、API 管理、字典、参数、代码生成等后端模型。
- Gin-Vue-Admin 只作为布局、交互和信息架构参考；不实现它的编程辅助、代码生成、插件市场或插件安装/打包系统。
- 浏览器本地 store 必须只在客户端安全 hydrate，能从损坏的 `localStorage` 恢复，并避免 SSR 崩溃。
- 本地 UI 偏好只保存展示状态，不保存凭据、token 或私有 API 响应。
- Admin UI 偏好使用 `localStorage` key `aoi.admin.ui.v1`。
- 访问标签使用 `sessionStorage` key `aoi.admin.visited-tabs.v1`。
- 认证 token pair 通过 `adminSession.ts` 写入 session storage，不要迁移到持久 `localStorage`。

## i18n 规则

- 默认语言是 `zh-CN`，路由策略是 `no_prefix`。
- 新增共享用户可见文案时，同步维护 `zh-CN.json`、`en.json` 和 `ja.json`。
- 部分现有功能页仍有内联中文文案。大幅触碰这些区域时，优先把可复用文案迁移到 locale 文件。

## 验证规则

- 修改 TypeScript、Vue、路由、composable 或 store 后，运行 `pnpm typecheck`。
- 修改 Nuxt 配置、server route、runtime config 或构建敏感模块后，运行 `pnpm build`。
- 需要由 Go 静态托管时，运行 `pnpm generate` 并确认 `web/aoi-web/.output/public/index.html` 存在。
- 可见 UI 变更必须用 Browser/视觉检查桌面和移动端表现，至少覆盖 `1440x900` 与 `390x844`，并在最终回复中说明检查路线、视口和风险。
- 如果未能运行必要验证，需要在最终回复中说明。

## Git 与协作

- 编辑前先检查工作区状态。
- 不要回滚用户改动或无关脏文件。
- 除非用户明确要求，不要提交、创建分支或推送。
- 只有在通过 pnpm 有意变更依赖时，才保留 lockfile 变化。
