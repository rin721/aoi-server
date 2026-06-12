# AI Workspace

This directory is the isolated AI operating layer for `go-scaffold`. It gives
future Codex sessions a fast path into the repo without mixing agent notes into
application packages.

## Index

- `project-map.md`: repo shape, module boundaries, and extension flow.
- `tooling.md`: installed AI, GitHub, lint, and security tools.
- `prompts.md`: reusable prompts for common development tasks.
- `handoff-template.md`: compact handoff format for long-running work.
- `gin-vue-admin-parity.md`: persistent task book, status board, visual
  reference notes, and handoff point for incremental Gin-Vue-Admin parity work.

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
## GVA 平替任务书

`gin-vue-admin-parity.md` 是当前逐步平替 Gin-Vue-Admin demo 的持久化任务书。每个切片开始前先记录研究计划和视觉证据；版本管理和媒体库切片已记录 GVA demo 截图、上游源码入口、本地实现边界和验证计划。
