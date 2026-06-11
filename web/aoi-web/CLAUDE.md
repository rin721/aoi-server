# Claude Code 规则

请遵循 `AGENTS.md` 中的项目规则。这个文件用于让 Claude Code 发现同一套仓库约定。

## Claude Code 专用说明

- 修改代码前先阅读 `AGENTS.md`。
- 较大的产品、架构、UI、API 或交互变更，同时阅读 `design/rules.md`。
- 只使用 `pnpm` 命令；不要引入 npm、Yarn 或 Bun 的 lockfile。
- Material Web 必须保持在 `app/components/aoi/` 的 Aoi wrapper 后面。
- 普通内部路由或外部链接使用 `AoiLink`，不要在业务代码中直接使用 `NuxtLink` 或裸 `<a>`。
- 按钮式链接使用 `AoiButton` 或 `AoiIconButton` 的 `to`/`href`。
- 优先做小而可验证的改动，并保留无关的工作区改动。
- 代码变更后，在可行时运行 `pnpm typecheck`，并报告跳过的检查。
- 用户使用中文时，默认使用简体中文回复，除非用户要求其他语言。
