# AI And Tooling Setup

## Codex Capabilities

Installed personal skills:

- `security-best-practices`
- `security-threat-model`
- `security-ownership-map`

Available plugin-backed workflows in this environment:

- GitHub repo, PR, issue, and comment triage.
- GitHub review-comment follow-up through `gh-address-comments`.
- GitHub Actions failure investigation through `gh-fix-ci`.
- Draft PR publishing through `yeet`.
- OpenAI docs lookup and Browser-based local web checks when needed.

Restart Codex after skill installation so newly installed skills are loaded in
future sessions.

## Local Tools

Expected tools:

```powershell
codex --version
gh --version
golangci-lint --version
govulncheck -version
gosec -version
osv-scanner --version
```

GitHub CLI needs user authentication before PR and CI workflows can inspect
private or account-scoped data:

```powershell
gh auth login
gh auth status
```

## Why These Tools

- `gh`: required for GitHub Actions logs and current-branch PR discovery.
- `golangci-lint`: fast local quality gate with a repo-specific config.
- `govulncheck`: official Go vulnerability analysis with reachability context.
- `gosec`: Go security static analysis for common secure coding issues.
- `osv-scanner`: dependency and source vulnerability scan using OSV data.

## Deferred Tools

Do not install these by default:

- Docker Desktop: only needed for container integration tests or image scans.
- Semgrep: useful for heavier SAST, but it adds another rule surface.
- Testcontainers: useful after real database integration tests are introduced.
- Figma, Notion, Linear, Cloudflare, and frontend Playwright skills: not
  aligned with the current backend scaffold scope.

