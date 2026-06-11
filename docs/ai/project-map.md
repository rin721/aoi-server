# Project Map For Agents

## Architecture

`go-scaffold` is a layered Go backend scaffold:

```text
cmd/main
  -> internal/app
      -> internal/config
      -> pkg infrastructure
      -> internal/modules
      -> internal/transport/http
```

The dependency direction matters. `pkg` packages are reusable infrastructure and
should not import `internal/app` or business modules. Application modules should
use infrastructure interfaces and keep HTTP details in handlers.

## Main Areas

- CLI: `cmd/main` declares commands and delegates execution.
- Composition: `internal/app` wires config, logger, i18n, database, cache,
  storage, modules, router, and lifecycle.
- Config: `internal/config` owns YAML, dotenv, env overrides, validation, and
  runtime reload support.
- HTTP: `internal/transport/http` registers health, readiness, demo, and IAM
  routes through `pkg/web` and Gin.
- Modules: `internal/modules/demo` and `internal/modules/iam` follow
  `model -> repository -> service -> handler`.
- Infrastructure: `pkg/database`, `pkg/cache`, `pkg/logger`, `pkg/httpserver`,
  `pkg/storage`, `pkg/sqlgen`, `pkg/token`, `pkg/authorization`, `pkg/mfa`,
  `pkg/migrator`, and related helpers.
- Shared types: `types/constants`, `types/errors`, and `types/result`.

## Extension Flow

For a new application module:

1. Add code under `internal/modules/<name>`.
2. Add config under `internal/config` only if runtime behavior needs config.
3. Wire repository, service, and handler in `internal/app/initapp`.
4. Register HTTP routes in `internal/transport/http`.
5. Add focused service and route tests.
6. Update human docs and, when useful, add AI notes in `docs/ai`.

## High-Risk Areas

- IAM auth, authorization, sessions, MFA, invitations, and audit logs.
- Database migrations and rollback behavior.
- Config reload and startup lifecycle.
- Shared response/error helpers used by HTTP middleware.
- Security-sensitive helpers in `pkg/token`, `pkg/crypto`, `pkg/mfa`, and
  `pkg/authorization`.

