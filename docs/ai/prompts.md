# Reusable Prompts

Use these as starting points. Add concrete files, failing commands, and scope
limits before sending them to an agent.

## Understand A Flow

```text
Trace the <feature> flow in this repo. Start from the CLI or HTTP route, follow
composition through internal/app, then identify service/repository/pkg
boundaries. Do not edit files. Return the main files and a concise data flow.
```

## Add A Module

```text
Add a <module-name> module following the existing model -> repository -> service
-> handler pattern. Keep infrastructure in pkg untouched unless required.
Add focused service and HTTP route tests. Update docs only for implemented
behavior.
```

## Debug CI

```text
Use GitHub tooling to inspect failing checks for the current PR. Fetch the
failing job logs, identify the first actionable failure, then propose a focused
fix before editing. If gh is not authenticated, stop and say so.
```

## Security Review

```text
Review the IAM/auth/security-sensitive changes in this branch. Prioritize bugs,
auth bypasses, missing validation, secret exposure, unsafe file or SQL handling,
and missing tests. Findings first, with file and line references.
```

## Documentation Refresh

```text
Compare docs for <area> against the current code. Update docs only for verified
current behavior. Put future work or missing capabilities in known gaps or
docs/ai notes, not in user-facing docs as if already implemented.
```

