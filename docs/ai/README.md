# AI Workspace

This directory is the isolated AI operating layer for `go-scaffold`. It gives
future Codex sessions a fast path into the repo without mixing agent notes into
application packages.

## Index

- `project-map.md`: repo shape, module boundaries, and extension flow.
- `tooling.md`: installed AI, GitHub, lint, and security tools.
- `prompts.md`: reusable prompts for common development tasks.
- `handoff-template.md`: compact handoff format for long-running work.
- `progressive-project-audit.md`: active task book for the whole-project
  progressive audit and implementation discipline.
- `admin-template-parity.md`: persistent task book, status board, visual
  reference notes, and handoff point for incremental 外部后台 parity work.
- `generator-product-spec.md`: gated product and security specification for
  template/code/form/export generator work before any runtime implementation.

## Default Workflow

1. Read `AGENTS.md` and this directory before planning broad changes.
2. Verify the current branch and worktree with `git status --short --branch`.
3. Use focused searches before editing: `rg`, `rg --files`, and targeted file
   reads.
4. Keep AI-generated reports in `tmp/ai` unless they are intentionally promoted
   into docs.
5. Validate with the nearest relevant checks, then expand to the full suite
   when touching shared behavior.

## Core Commands

```powershell
go test ./... -count=1 -mod=readonly
go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main
golangci-lint run --config tools/ai/golangci.yml ./...
govulncheck ./...
gosec ./...
osv-scanner scan source .
```
## 外部后台 平替任务书

`admin-template-parity.md` 是当前逐步完善后台功能的持久化任务书。每个切片开始前先记录研究计划和视觉证据；版本管理、媒体库、断点上传和客户资源示例切片已记录外部研究入口、上游源码入口、本地实现边界和验证计划。模板配置、代码生成、表单生成和导出模板切片当前只完成本地审计，确认没有运行时后台/API/配置入口；后续实现前必须先从 `generator-product-spec.md` 补齐产品问题、安全门禁和统一配置方案。

## 渐进式全项目审计

`progressive-project-audit.md` 是长期总任务书。任何实现切片开始前，先按其中的
项目结构、配置体系、前后端调用链、构建流程、静态托管、样式体系和相关文档
清单取证，再记录本轮改什么、不改什么、配置落点、风险和验证计划。
