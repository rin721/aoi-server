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
- System parity module: `internal/modules/system` follows the same
  `model -> repository -> service -> handler` flow for 上游样板-inspired
  menus, API catalog sync, dictionaries, operation records, parameters,
  read-only config snapshots, server status, and idempotent default data
  seeding during startup.
- API catalog entries expose route access mode (`public`, `authenticated`, or
  `permission`) inferred from HTTP route registration and permission metadata;
  this mirrors 上游样板的 public/private router split without renaming transport
  packages or storing duplicate middleware state.
- Login logs are a 管理后台式 admin page over existing IAM `auth.login` audit
  records; do not create a duplicate login-log table until there is a product
  requirement for different retention or fields.
- Login captcha is a 管理后台式 optional IAM workflow. `GET /api/v1/auth/captcha`
  is public, login accepts optional `captchaId` and `captchaCode`, and challenge
  state stays in the IAM service memory by default; do not add a captcha table
  without a persistence or multi-node requirement.
- Error logs are a 管理后台式 admin page over `system_operation_records` with
  `4xx`, `5xx`, and all-error filters; do not create a duplicate error-log
  table while the operation record schema contains status, response, trace ID,
  and error message fields.
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

For 上游样板 parity slices, keep the existing Go scaffold boundaries:

1. Route registration and API catalog metadata stay in `internal/transport/http`.
2. Request parsing and response writing stay in module handlers.
3. Validation, transactions, sync behavior, masking, and domain rules stay in
   services.
4. Database access stays behind repository interfaces.
5. Shared envelopes and constants stay in `types` and reusable helpers stay in
   `pkg`.
6. If the backend change is visible in the admin UI, perform Browser visual
   inspection of the affected route at desktop and mobile sizes.

## High-Risk Areas

- IAM auth, authorization, sessions, MFA, invitations, and audit logs.
- Database migrations and rollback behavior.
- Config reload and startup lifecycle.
- Shared response/error helpers used by HTTP middleware.
- Security-sensitive helpers in `pkg/token`, `pkg/crypto`, `pkg/mfa`, and
  `pkg/authorization`.
