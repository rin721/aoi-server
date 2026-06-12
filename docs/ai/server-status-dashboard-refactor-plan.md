# Server Status / Dashboard Refactor Plan

## 项目现状识别

- 状态：IN_PROGRESS
- 实际技术栈：Go 1.25.7 后端，HTTP 传输层通过 `pkg/web` 封装 Gin；后台前端位于 `web/admin`，使用 Nuxt 4、Vue 3、TypeScript、Pinia、Material Web，并通过 Aoi wrapper 暴露业务可用组件。
- 前端构建方式：`web/admin` 使用 `pnpm typecheck`、`pnpm build` 做检查，使用 `pnpm generate` 生成 Go 静态托管产物。
- 后端加载静态产物方式：Go WebUI 配置默认 `webui.mount_path=/admin`，`webui.dist_dir=./web/admin/.output/public`；HTTP 路由通过 `MountStaticSPA` 托管 SPA 并保留 API、health、ready 路由边界。
- 当前页面入口：前端页面为 `web/admin/app/pages/server-info.vue`，API 调用为 `useAdminApi().getSystemServerInfo()`，后端路由为 `GET /api/v1/system/server-info`。
- 当前数据边界：接口返回 runtime、CPU、RAM、disk、Go memory、GC、build；当前不返回 GPU、CI/CD、服务进程明细或任务队列状态。
- 主要问题概览：页面信息层级偏平，CPU 与构建信息容易形成细长高卡片；状态判断、字段 label、单位格式化、KPI 顺序仍散落在页面；当前无服务明细却容易被误设计成伪数据区域；后续必须继续治理配置、状态、视觉和文档。

## 总体重构目标

- UI 目标：让用户进入服务器状态页 3 秒内判断整体健康，CPU、内存、磁盘、Go 运行态和构建信息能快速扫视，异常指标优先突出。
- 架构目标：页面只消费配置和派生模型，不直接写阈值、字段映射、排序规则、状态文案、单位换算或刷新策略。
- 配置治理目标：集中管理指标阈值、状态文案、状态权重、KPI 顺序、资源面板顺序、字段 label、格式化规则、空值文案和局部滚动限制。
- 文档治理目标：同步维护开发者、二次开发者、维护者、使用者和新手说明；示例配置不得包含真实账号、密码、Token 或生产地址。
- Git 工作流目标：每个阶段先更新任务书，再小步实现、验证、提交，并合并到 main；禁止强制覆盖和危险清理。

## 阶段路线

| 阶段 | 状态 | 目标 |
| --- | --- | --- |
| 阶段 0：工作区归属与安全收口 | DONE | 记录当前脏工作区，把现有未提交 UI wrapper 改动作为 pre-existing baseline 单独归档。 |
| 阶段 1：现状审计与任务书建立 | DONE | 建立本任务书，确认技术栈、页面入口、静态产物链路、数据边界和风险。 |
| 阶段 2：配置体系梳理与硬编码治理 | DONE | 新增 Server Status Dashboard 配置入口，迁出 KPI、字段、阈值、状态文案和 API endpoint。 |
| 阶段 3：状态体系与格式化工具治理 | DONE | 新增统一派生模型、状态判断、单位换算、空值 fallback 和异常优先规则。 |
| 阶段 4：UI 视觉基础治理 | DONE | 完善 Aoi 数据状态组件和 Server Status CSS tokens，继续复用既有 Aoi wrapper。 |
| 阶段 5：Dashboard 布局重构 | DONE | 重构服务器状态页头部、KPI、资源区、CPU 局部滚动和构建信息展示。 |
| 阶段 6：资源模块优化 | DONE | 优化 CPU、内存、磁盘、Go heap/GC；GPU 仅保留无数据扩展点，不伪造数据。 |
| 阶段 7：服务与任务区域优化 | NEXT | 等后端提供服务明细或任务数据后再表格化；当前只记录空状态策略和接口扩展点。 |
| 阶段 8：文档、示例配置、注释完善 | DONE | 更新开发、维护、使用、新手和示例配置文档。 |
| 阶段 9：最终验证 | DONE | 运行前后端检查，验证 Go 静态托管、Browser 视觉和 Git 状态。 |

## 当前任务

- 本次目标：完成阶段 2 到阶段 9 的一个可验证切片；把 Server Status 的展示配置、API endpoint、状态判断、格式化、派生模型、视觉结构、文档和静态托管验证纳入治理。
- 是否需要先分析研究：DONE。已只读确认 Git 状态、项目结构、技术栈、服务器状态页面、后端接口、采集逻辑、静态 SPA 挂载和文档位置。
- 计划修改范围：本任务书、`web/admin/app/config`、`web/admin/app/utils`、`web/admin/app/components/aoi`、`useAdminApi()` 的 endpoint 引用、`server-info.vue`、后台 CSS tokens、系统/维护/使用/环境文档和示例配置。
- 不修改范围：不改 Go API、DTO、数据库迁移、认证逻辑、真实账号信息、生产配置或数据目录。
- 风险：RISK 当前页面已存在 baseline UI 改造，后续配置化必须保持视觉行为兼容；RISK 当前无 GPU、CI/CD、服务明细数据，不能伪造模块。
- 验证方式：运行 `pnpm typecheck`、`pnpm build`、`pnpm generate`、`go test ./... -count=1 -mod=readonly`、`go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main`，并通过 Go 静态托管访问 `/admin/login` 和 `/admin/server-info`；Browser 抽查 `1440x900`、`1280x720`、`390x844`。
- Git 分支计划：在 `codex/server-status-dashboard-governance` 完成本阶段提交，再合并回 main。

## 任务状态表

| 项 | 状态 | 说明 |
| --- | --- | --- |
| 读取 AGENTS 与项目约束 | DONE | 根目录和 `web/admin` 约束已识别。 |
| 检查 Git 状态 | DONE | 当前分支、main、未提交文件和超前提交已确认。 |
| 识别技术栈 | DONE | Go + Nuxt 4 + Vue 3 + TypeScript + Pinia + Aoi wrapper。 |
| 识别 Server Status 调用链 | DONE | 页面、API、DTO、service、hostmetrics、router 已确认。 |
| 识别静态托管链路 | DONE | `webui` 配置和 `MountStaticSPA` 已确认。 |
| 创建本地任务书 | DONE | 本文件为持久化任务书。 |
| 提交任务书 | DONE | commit：`e885016 docs: add server status dashboard refactor plan`。 |
| 归档现有 UI baseline | DONE | commit：`7c1f101 refactor: consolidate admin visual baseline`。 |
| 合并当前分支到 main | DONE | 已 fast-forward 合并 `codex/org-management-pagination` 到 main。 |
| 创建后续治理分支 | DONE | 当前分支：`codex/server-status-dashboard-governance`。 |
| 集中 Dashboard 配置 | DONE | 新增 Server Status 配置入口，集中 KPI、panel、字段、阈值、状态文案和格式化规则。 |
| 统一派生模型 | DONE | 新增状态判断、格式化工具、异常优先、容量换算和空值 fallback。 |
| 重构 Server Status 页面 | DONE | 页面改为健康优先 header、配置驱动 KPI、资源面板、CPU 和构建信息局部滚动。 |
| 更新文档与示例配置 | DONE | 更新开发、维护、使用、新手、环境配置和示例配置文档。 |
| 静态托管与浏览器验证 | DONE | `pnpm generate` 后通过 Go 服务访问 `/admin/login` 与 `/admin/server-info`。 |

## 变更记录

### 2026-06-13

- 状态：DONE
- 修改文件：`docs/ai/server-status-dashboard-refactor-plan.md`
- 修改摘要：创建 Server Status / Dashboard 持续治理任务书，记录现状、路线、本次计划、风险和 Git 策略。
- 验证结果：`pnpm typecheck`、`pnpm build`、`pnpm generate` 通过；`.output/public/index.html` 存在。
- commit hash：`e885016`、`7c1f101`
- 是否已合并 main：是，fast-forward。
- 下一步建议：在 `codex/server-status-dashboard-governance` 完成配置治理和派生模型治理。

### 2026-06-13 阶段 2-9

- 状态：DONE
- 修改文件：`web/admin/app/config/admin-api.ts`、`web/admin/app/config/server-status-dashboard.ts`、`web/admin/app/utils/serverStatusDashboard.ts`、`web/admin/app/components/aoi/AoiDataState.vue`、`web/admin/app/pages/server-info.vue`、`web/admin/app/assets/css/main.css`、`web/admin/app/composables/useAdminApi.ts`、`docs/modules/system.md`、`docs/maintenance/server-status-dashboard.md`、`docs/onboarding/server-status-dashboard.md`、`docs/environment/configuration.md`、`configs/config.example.yaml`、`docs/ai/admin-template-parity.md`。
- 修改摘要：集中 Server Status 配置、API endpoint、状态判断、格式化和派生模型；重构页面为健康优先的配置驱动 Dashboard；补齐数据状态组件、CSS tokens、文档和示例配置说明；移除 AI 文档中的明文测试账号组合表述。
- 验证结果：`pnpm typecheck`、`pnpm build`、`pnpm generate`、`.output/public/index.html` 检查、`go test ./... -count=1 -mod=readonly`、`go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main` 通过；Go 静态托管 `/admin/login` 和 `/admin/server-info` 访问通过；Browser 在 `1440x900`、`1280x720`、`390x844` 下无横向溢出、无 `undefined/null/NaN` 文本。
- commit hash：`c14c45a`
- 是否已合并 main：待本记录提交后合并。
- 下一步建议：NEXT 等后端真实提供 GPU、CI/CD、服务进程明细后，再继续阶段 7 的服务与任务区域结构化治理。
