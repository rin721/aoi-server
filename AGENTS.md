# AGENTS.md

This file is the repo-level operating guide for coding agents working in
`go-scaffold`. Keep it focused on repository facts and workflows. Human-facing
engineering documentation starts in `docs/README.md`; AI-specific notes live in
`docs/ai`.

## Project Snapshot

- Runtime: Go 1.25.7.
- Module: `github.com/rei0721/go-scaffold`.
- Shape: Go backend scaffold with Gin HTTP routing, Cobra/Bubble Tea CLI,
  GORM database support, goose migrations, IAM, Redis cache, storage helpers,
  SQL generation, Docker examples, and GitHub Actions.
- Default local service: `go run ./cmd/main server`, listening on
  `127.0.0.1:9999` with SQLite under `data/`.

## Source Map

- `cmd/main`: process entrypoint and command specs. Keep it thin.
- `internal/app`: application composition, lifecycle, reload, and startup.
- `internal/config`: config structs, env overrides, validation, and watching.
- `internal/modules`: application modules. Existing modules use
  `model -> repository -> service -> handler`.
- `internal/transport/http`: router assembly and HTTP route registration.
- `internal/middleware`: transport middleware for trace, auth, i18n, CORS,
  recovery, and logging.
- `pkg`: reusable infrastructure packages that must not depend on app modules.
- `types`: shared constants, errors, and result helpers.
- `docs/ai` and `tools/ai`: isolated AI workspace and AI tooling configuration.

## Change Boundaries

- Do not mix AI artifacts into `cmd`, `internal`, `pkg`, or `types`.
- Do not edit `configs/config.yaml`, `data/`, `tmp/`, local env files, or
  generated runtime output unless the task explicitly requires it.
- Prefer updating `configs/*.example.yaml`, `.env.example`, or docs when
  documenting configuration behavior.
- Keep `pkg` reusable and free of imports from `internal/modules` or
  `internal/app`.
- Keep business logic out of handlers; put validation, transactions, and
  domain rules in services.
- Treat migrations in `internal/migrations` as append-only once shared.

## Standard Commands

Use these from the repository root:

```powershell
go test ./... -count=1 -mod=readonly
go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main
golangci-lint run --config tools/ai/golangci.yml ./...
govulncheck ./...
gosec ./...
osv-scanner scan source .
```

For focused checks, run the nearest package test first, then the full suite when
the change crosses package, config, HTTP, database, or shared type boundaries.

## GitHub Workflow

- The remote is expected to be `git@github.com:rin721/go-scaffold.git`.
- Use the GitHub plugin/connector for repository, PR, issue, label, comment,
  and PR creation workflows when available.
- Use GitHub CLI for local branch PR discovery and Actions logs:
  `gh auth status`, `gh pr view`, `gh pr checks`, and `gh run view`.
- If `gh auth status` reports no login, ask the user to run `gh auth login`
  before attempting CI-log or PR-thread workflows.

## AI Workspace

- `docs/ai/README.md`: index for AI operating notes.
- `docs/ai/project-map.md`: compact architecture map.
- `docs/ai/tooling.md`: installed tools and setup commands.
- `docs/ai/prompts.md`: reusable prompts for common repo work.
- `docs/ai/handoff-template.md`: short handoff template for long tasks.
- `tools/ai/golangci.yml`: AI-assisted lint configuration.
- `tools/ai/security-checks.md`: local security scan runbook.
- Put short-lived reports under `tmp/ai`; `tmp/` is ignored by git.

## Documentation Notes

- Existing docs are intended to describe current behavior, not future wishes.
  Put future or missing capabilities in `docs/backlog/known-gaps.md` or
  `docs/ai` if they are only for agent operation.
- Write documentation comments in Chinese.
- Preserve concrete commands, file paths, and verified facts.
- If terminal output displays mojibake, inspect the file in an editor or raw
  bytes before rewriting large doc sections.
