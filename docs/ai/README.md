# AI Workspace

This directory is the isolated AI operating layer for `go-scaffold`. It gives
future Codex sessions a fast path into the repo without mixing agent notes into
application packages.

## Index

- `project-map.md`: repo shape, module boundaries, and extension flow.
- `tooling.md`: installed AI, GitHub, lint, and security tools.
- `prompts.md`: reusable prompts for common development tasks.
- `handoff-template.md`: compact handoff format for long-running work.
- `gin-vue-admin-parity.md`: reference notes for incremental Gin-Vue-Admin
  parity work.

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
