# 维护指南

维护工作应保持代码、测试、文档和 AI 运行态一致。

## 常规变更流程

1. 阅读 `AGENTS.md` 和相关 `docs/ai` 运行态。
2. 识别本次变更真实影响的代码边界。
3. 更新代码和测试。
4. 更新 `docs` 下的结构化人类文档。
5. 如果运行态变化，更新或修复 `docs/ai` artifact。
6. 根据影响范围运行定向测试和更广测试。
7. 将剩余风险记录到 `docs/backlog/known-gaps.md` 或运行态证据中。

## 文档卫生

人类文档应解释当前代码，而不是把未来想法写成既成事实。未来工作或缺失能力应放入 backlog/known gaps，除非明确标记为计划变更。

优先在结构化目录中增加文档，避免继续增加顶层零散文件。如果旧链接可能被外部引用，可以保留短兼容入口。

## 运行态卫生

`docs/ai` 是运行态系统。当某个 task 或 slice 成为当前工作，它应能通过 current status、task tree、requirement ledger、evidence index
和 handoff 被发现。如果某个 artifact 缺失或太薄，应修复物理 artifact，而不是依赖聊天历史。

## API Token 维护

- 发布包含 API Token 的版本前，确认 `internal/migrations/20260612000400_create_iam_api_tokens.sql` 已在目标环境执行。
- `auth.refresh_token_pepper` 同时保护 refresh token 和 API Token hash；轮换该值会让既有 refresh token 与 API Token 全部失效，发布说明里必须提前告知调用方。
- API Token 明文只在签发成功时出现一次。排障时不要要求用户把完整 token 贴到 issue、日志或聊天记录中，优先使用 `tokenPrefix`、`tokenId` 和审计日志定位。
- 调用方泄漏 token、用户被禁用、角色权限收缩或自动化任务下线时，应在 `/admin/api-tokens` 或 `DELETE /api/v1/orgs/{orgId}/api-tokens/{tokenId}` 立即撤销。

## Review 清单

- 变更是否保持目录边界？
- 配置示例和 env 文档是否同步？
- 启动、reload、shutdown 影响是否记录？
- 测试是否靠近它保护的行为？
- 生产风险是否清楚标记？
- AI 运行态 artifact 是否与当前工作状态一致？
## 版本发布包维护

- 发布包含版本管理的代码前，确认 `internal/migrations/20260612000500_create_system_versions.sql` 已在目标环境执行。
- `system_versions.version_data` 是完整 JSON 包，可能包含菜单路径、API 路径和字典内容；生产排障可以下载比对，但不要把敏感业务字典直接贴到公开 issue。
- 当前导入只会幂等创建缺失字典和字典项。菜单和 API 来自代码/路由目录，导入结果中的 `menusSkipped` 与 `apisSkipped` 属于预期行为。
- 如果后续把菜单或 API 改成数据库可编辑，必须同步更新版本导入冲突策略、回滚说明、OpenAPI 和 `docs/modules/system.md`。

## 媒体库维护

- 发布包含媒体库的代码前，确认 `internal/migrations/20260612000600_create_system_media.sql` 已在目标环境执行，并同步 `media:*` IAM 权限。
- 普通上传依赖 `storage.enabled=true`。本地建议使用 `storage.fs_type=basepath` 和 `storage.base_path=./data`，避免把对象写到进程工作目录的非预期位置。
- URL 导入只保存外链，不下载远程文件。排障时优先检查 `system_media_assets.external`、`url`、`storage_key` 和 storage 配置。
- 删除本地媒体资源会先尝试删除对象，再软删除数据库记录；如果 storage 暂不可用，先恢复 storage，再执行删除，避免形成孤儿对象。
