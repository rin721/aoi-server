# Server Status Dashboard 维护指南

本文面向维护者，记录服务器状态 Dashboard 的治理入口、排查方式和发布检查。持续任务书位于 `docs/ai/server-status-dashboard-refactor-plan.md`。

## 阶段状态

| 阶段 | 状态 | 说明 |
| --- | --- | --- |
| 任务书与现状审计 | DONE | 已识别 Go 静态托管链路、Nuxt 前端入口和当前 API 数据边界。 |
| 配置治理 | DONE | KPI、字段 label、阈值、状态文案、刷新策略和 endpoint 已集中到前端配置。 |
| 状态与格式化治理 | DONE | 状态判断、异常排序、容量换算、空值 fallback 已收敛到派生工具。 |
| 视觉基础治理 | DONE | Server Status 使用 Aoi wrapper、后台 tokens 和统一数据状态组件。 |
| 服务明细 / GPU / CI | NEXT | 当前接口没有真实数据，不在前端伪造。 |

## 配置治理范围

- API endpoint：`web/admin/app/config/admin-api.ts`。
- Dashboard 展示配置：`web/admin/app/config/server-status-dashboard.ts`。
- UI tokens：`web/admin/app/assets/css/main.css` 中的 `--aoi-admin-*` 变量。
- 派生模型：`web/admin/app/utils/serverStatusDashboard.ts`。

不要在 `web/admin/app/pages/server-info.vue` 中新增阈值、接口路径、字段 label、单位换算、状态文案、状态颜色或刷新间隔。

## 常见问题

| 问题 | 排查入口 |
| --- | --- |
| 页面显示 `-` | 检查后端 DTO 是否返回字段；再检查 `valueForField` 是否已映射字段。 |
| 状态颜色不符合预期 | 检查 `thresholds`、`statusLabels`、`statusWeights`，不要改页面 CSS。 |
| CPU 核心列表过高 | 检查 `--aoi-admin-server-cpu-scroll-max-height` 和 `AoiMasonryGrid` 布局偏好。 |
| 构建信息挤成细列 | 检查 `AoiKeyValueList` rows 布局和 `--aoi-admin-server-build-scroll-max-height`。 |
| 静态资源 404 | 确认先在 `web/admin` 执行 `pnpm generate`，再检查 `webui.mount_path` 与 `webui.dist_dir`。 |
| API 不通 | `NUXT_PUBLIC_API_BASE_URL` 为空时表示同源请求 Go API；非空时检查部署网关配置。 |

## 验证清单

1. 在 `web/admin` 执行 `pnpm typecheck`。
2. 执行 `pnpm build`，记录非阻断警告。
3. 执行 `pnpm generate`，确认 `.output/public/index.html` 存在。
4. 在仓库根目录执行 `go test ./... -count=1 -mod=readonly`。
5. 启动 Go 服务后访问 `/admin/login` 和 `/admin/server-info`，确认 SPA fallback 与静态资源正常。
6. 在 1440x900、1280x720、390x844 检查无横向溢出、无 `undefined`、`null`、`NaN` 文本。

## Git 收口

每个可验证阶段都应提交并合并到 `main`。合并优先使用 fast-forward；如 `main` 前进，先只读检查差异，再用普通 merge 解决冲突，禁止强制覆盖或清理他人改动。
