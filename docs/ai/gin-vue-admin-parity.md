# Gin-Vue-Admin Parity Notes

Verified on 2026-06-11 from the public demo and upstream documentation.
Sources:

- https://demo.gin-vue-admin.com
- https://www.gin-vue-admin.com/guide/server/
- https://github.com/flipped-aurora/gin-vue-admin/tree/main/server

## Persistent Task Book

This file is the durable handoff point for incremental Gin-Vue-Admin parity.
Before implementing a parity slice, update or append a short entry here with:

1. target slice and current status;
2. whether the public demo must be researched again;
3. visual evidence needed before/after the change;
4. backend/source reference needed before coding;
5. docs, example config, and validation commands to update or run.

Captcha rule: if the public demo login asks for a captcha, stop and ask the
user for the visible captcha value. Do not store local project credentials in
this repository document.

Status legend:

- `[done]`: implemented and documented in this repository.
- `[doing]`: actively being implemented in the current parity slice.
- `[audit]`: present locally but needs GVA comparison or polish before calling
  it equivalent.
- `[next]`: preferred next implementation slice.
- `[todo]`: not implemented or intentionally deferred.
- `[defer]`: possible later work that needs a stronger product decision.

### Current Snapshot

Last checked: 2026-06-12.

- Demo access: `https://demo.gin-vue-admin.com` opened directly into an
  authenticated admin dashboard during this session; no captcha blocker was
  observed.
- Visual evidence saved for this session:
  `tmp/ai/gva-dashboard-2026-06-12.png`,
  `tmp/ai/gva-api-token-2026-06-12.png`,
  `tmp/ai/gva-api-token-dialog-2026-06-12.png`,
  `tmp/ai/gva-version-2026-06-12.png`,
  `tmp/ai/gva-version-export-drawer-2026-06-12.png`,
  `tmp/ai/gva-version-import-drawer-2026-06-12.png`,
  `tmp/ai/gva-upload-2026-06-12.png`,
  `tmp/ai/gva-upload-import-url-2026-06-12.png`,
  `tmp/ai/version-desktop-1440x900-final.png`,
  `tmp/ai/version-export-dialog-1440x900-final.png`,
  `tmp/ai/version-mobile-390x844-final.png`,
  `tmp/ai/api-token-desktop-1440x900-final.png`,
  `tmp/ai/api-token-issue-dialog-1440x900-final.png`,
  `tmp/ai/api-token-mobile-390x844-final.png`, and
  `tmp/ai/api-token-mobile-dialog-390x844-final.png`,
  `tmp/ai/media-desktop-1440x900-final.png`,
  `tmp/ai/media-import-url-1440x900-final.png`,
  `tmp/ai/media-mobile-390x844-final.png`, and
  `tmp/ai/media-mobile-390x844-viewport.png`.
- Demo visual readout: fixed white shell, left menu, thin dividers, compact
  operational cards, dense tables, small action buttons, minimal decorative
  effects, and visible diagonal demo watermark.
- Local parity stance: continue mapping GVA responsibilities into this repo's
  existing `internal/modules/*/{model,repository,service,handler}`,
  `internal/transport/http`, `internal/app/initapp`, `internal/config`, and
  `pkg` boundaries instead of renaming the backend to GVA's folders.

### Route And Feature Board

| Status | GVA area | Local target | Research gate | Notes |
| --- | --- | --- | --- | --- |
| [done] | Admin shell, left menu, visited tabs, dense table styling | `web/admin/app` layout and shared CSS | Visual before/after required for every future UI slice | Keep low-noise GVA-like management style. |
| [done] | Dashboard baseline | `/admin` | Re-check demo dashboard before major dashboard redesign | Local dashboard is service/IAM-focused, not plugin-market focused. |
| [done] | Menu management | `/admin/menus`, `/api/v1/system/menus` | Demo page optional unless changing interactions | Server-driven menu catalog is the local source of truth. |
| [done] | API management and sync | `/admin/apis`, `/api/v1/system/apis` | Demo page optional unless changing access/filter UI | Includes access-mode summary and permission sync. |
| [done] | Role authorization matrix | `/admin/roles` | Re-check GVA role page before editing permission UX | Local implementation maps to Casbin domain RBAC. |
| [audit] | User management | `/admin/users` and IAM APIs | Must inspect GVA user page before next user-management change | Verify filters, invite/create flow, role binding, status controls. |
| [audit] | Organization/tenant management | `/admin/organizations` | Research GVA tenant/org equivalent before changing | Local IAM has organizations; GVA demo may not match one-to-one. |
| [audit] | Session/security/MFA pages | `/admin/sessions`, `/admin/security` | Visual and workflow check required before UI changes | Keep token and MFA behavior local; do not copy insecure demo shortcuts. |
| [done] | Dictionary management | `/admin/dictionaries`, system dictionary APIs | Demo page optional unless changing item editing UX | Persisted dictionaries and items are implemented. |
| [done] | Operation history | `/admin/operation-records` | Re-check demo before adding advanced filters/export | Persisted protected request records are implemented. |
| [done] | Parameter management | `/admin/parameters` | Demo page optional unless adding batch/import/export | Persisted parameter CRUD is implemented. |
| [done] | System config read | `/admin/system`, `/api/v1/system/config` | Research required before any write/reload capability | Current slice is read-only and masks secrets. |
| [done] | Server status | `/admin/server-info`, `/api/v1/system/server-info` | Visual check required if charts are added | Uses local runtime and host metrics. |
| [done] | Login log | `/admin/login-logs` | Demo page already observed as menu/tab with limited content | Local page uses IAM `auth.login` audit records. |
| [done] | Error log | `/admin/error-logs` | Demo page observed as unassigned route for admin account | Local page uses `system_operation_records` status filters. |
| [done] | Login captcha | `/api/v1/auth/captcha`, `/admin/login` | Demo login captcha must be checked if login screen changes | Optional config: `auth.login_captcha_enabled`. |
| [done] | API Token | `/admin/api-tokens`, `/api/v1/orgs/:orgId/api-tokens` | Demo page and upstream source inspected on 2026-06-12 | Local opaque token implementation with one-time plaintext display and hash-only storage. |
| [done] | Version management | `/admin/versions`, `/api/v1/system/versions` | Demo page and upstream source inspected on 2026-06-12 | GVA feature is a configuration release package for selected menus, APIs, and dictionaries; local import safely persists dictionaries and stores menu/API package records. |
| [done] | Media library upload/download | `/admin/media`, `/api/v1/system/media/*` | Demo page and upstream source inspected on 2026-06-12 | GVA feature is left-category media management with upload, URL import, keyword filter, preview, rename, download, and delete. |
| [todo] | Breakpoint upload | Storage resumable upload module | Must inspect demo workflow and source first | Higher risk: requires protocol, cleanup, limits, and docs. |
| [todo] | Customer/resource example | Demo module extension | Demo visual/source check required | Implement only as reusable CRUD example if still valuable. |
| [defer] | Template config/code generator/form generator/export template | `pkg/sqlgen` plus explicit product spec | Research required, but do not copy wholesale | Needs product decision and security review before implementation. |
| [defer] | AI workflow, MCP Tools, Skills, AI page drawing | Separate AI tooling boundary | Research required before any local work | Keep AI artifacts under `docs/ai` or `tools/ai`; do not mix into app packages. |
| [defer] | Plugin market/install/package/mail plugin/announcements | Existing `plugins` module plus product spec | Research required | Avoid remote marketplace/install behavior without an explicit requirement. |

### Next Slice Protocol

Preferred next slice: Breakpoint upload.

Before implementation:

- Research the GVA breakpoint upload workflow visually and capture screenshots
  for upload start, progress, pause/resume or retry, and completion states.
- Inspect upstream GVA server and web source for chunk request shape, merge
  behavior, cleanup rules, hash handling, and failure responses.
- Inspect local `pkg/storage`, storage config, media model, and migration
  status before deciding whether resumable upload belongs in System media or a
  separate storage module.
- Write the slice plan here before editing code.

Implementation guardrails:

- Treat this as a higher-risk storage protocol, not a small UI button.
- Use append-only migrations if chunk sessions or temporary object state need
  persistence.
- Keep file-system writes below the configured storage root, sanitize filenames,
  generate server-side object keys, and never trust uploaded or imported paths.
- Add explicit size, chunk count, TTL, cleanup, and idempotency rules before
  accepting any chunk.
- Add IAM permissions, server-driven menu entry, API catalog entries, and admin
  page together when the feature becomes user-visible.
- Update developer, maintainer, user, and beginner-facing docs when behavior is
  exposed.
- Update `configs/*.example.yaml` and `.env.example` when media storage knobs or
  operational notes are introduced.

Validation floor:

- Run focused Go tests for changed backend packages.
- Run `pnpm typecheck` for changed `web/admin` TypeScript/Vue code.
- For visible UI work, visually inspect desktop `1440x900` and mobile
  `390x844` routes with Browser and record results here or in the final note.

### Active Slice: Media Library Upload/Download

Status: `[done]` started and completed on 2026-06-12.

Research completed before implementation:

- Demo route: `https://demo.gin-vue-admin.com/#/layout/example/upload`.
- Visual evidence:
  `tmp/ai/gva-upload-2026-06-12.png` and
  `tmp/ai/gva-upload-import-url-2026-06-12.png`.
- Demo page shape: left category tree with `全部分类`; top warning bar; action
  buttons for normal upload, crop upload, QR upload, compressed upload, and
  URL import; keyword filter; table columns for preview, date, file name/remark,
  link, tag, and row operations; pagination at the lower right.
- Interaction readout: file name is editable from the table row; URL import
  accepts newline-separated `文件名|链接` or bare URL entries in a prompt.
- Upstream primary source checked:
  `web/src/view/example/upload/upload.vue`,
  `web/src/api/fileUploadAndDownload.js`,
  `web/src/api/attachmentCategory.js`,
  `server/model/example/exa_file_upload_download.go`,
  `server/model/example/exa_attachment_category.go`,
  `server/model/example/request/exa_file_upload_and_downloads.go`,
  `server/model/example/response/exa_file_upload_download.go`,
  `server/api/v1/example/exa_file_upload_download.go`,
  `server/api/v1/example/exa_attachment_category.go`,
  `server/service/example/exa_file_upload_download.go`,
  `server/service/example/exa_attachment_category.go`,
  `server/router/example/exa_file_upload_and_download.go`, and
  `server/router/example/exa_attachment_category.go` from
  `github.com/flipped-aurora/gin-vue-admin`.

Local implementation plan:

- Put the user-visible management surface in the System module because local
  media records are operational admin assets backed by `pkg/storage`, not a
  throwaway demo table.
- Add append-only tables `system_media_categories` and `system_media_assets`.
  Categories store ID, parent ID, name, sort, timestamps, and soft delete.
  Assets store ID, category ID, display name, original filename, storage key,
  URL/path, MIME type, extension/tag, byte size, source (`upload` or `url`),
  external flag, uploader identity, timestamps, and soft delete.
- Inject optional `pkg/storage.Storage` into the System service. When storage is
  disabled, list/imported URL records can still be visible from DB, but binary
  upload/download/delete of local objects must return a clear storage
  unavailable error.
- Expose protected APIs under `/api/v1/system/media`: category tree, create or
  update category, delete category, asset list, upload, URL import, rename,
  download, and delete.
- Add IAM permissions `media:read`, `media:upload`, `media:import`,
  `media:update`, `media:download`, and `media:delete`, then wire them through
  API catalog, role permission matrix, and server-driven menus.
- Build `/admin/media` with the GVA-like layout: left category panel, compact
  action/filter row, preview table/list, URL import dialog, rename dialog, and
  desktop/mobile responsive behavior. Start with normal upload and URL import;
  crop/QR/compress buttons may be represented as deferred actions only if they
  are not functionally implemented in this slice.
- Keep storage safety explicit: server-generated keys under a `media/` prefix,
  sanitized display names, size limit, MIME sniffing, no path traversal, no URL
  fetching during import, and best-effort object deletion when DB rows are
  deleted.
- Update developer, maintainer, user, beginner, API, OpenAPI, system module,
  storage/config, and AI handoff docs. Example config should document storage
  enablement and the media prefix/limit if new knobs are added.

Validation plan:

- Run focused Go tests for `internal/modules/system` and
  `internal/transport/http`, then `go test ./... -count=1 -mod=readonly`.
- Run `go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main`.
- Run `pnpm typecheck` from `web/admin`.
- Use Browser visual checks on `/admin/media` at `1440x900` and `390x844`,
  including URL import and normal empty/non-empty states where possible.

Implementation completed:

- Added append-only migration
  `internal/migrations/20260612000600_create_system_media.sql`, the media
  category and asset models, repository methods, service workflow, handlers,
  and protected HTTP routes for category tree, asset list, normal upload, URL
  import, rename, download, and delete.
- Reused `pkg/storage` through dependency injection. With storage disabled,
  media list and external URL imports still work, while local binary
  upload/download/delete report a clear unavailable state.
- Added IAM permissions `media:read`, `media:upload`, `media:import`,
  `media:update`, `media:download`, and `media:delete`, then wired them into
  the route catalog, server-driven menu, and role permission matrix.
- Added `/admin/media` with the GVA-like left category panel, action/filter
  row, warning bars, table/card resource list, URL import dialog, rename
  dialog, and responsive desktop/mobile behavior.
- Kept storage safety explicit: server-generated object keys under `media/`,
  sanitized display names, upload size limits, MIME sniffing, no trusted client
  paths, no remote URL fetching during import, and best-effort object deletion.
- Fixed a visual/runtime issue found during QA: nullable media API item lists
  are normalized before Vue renders `flatMap`, preventing a blank update cycle
  when an unavailable or empty storage-backed catalog returns `items: null`.
- Updated API, OpenAPI, system module, README, onboarding, maintenance,
  extension, environment, overview, IAM, AI handoff, and example config notes.

Validation completed:

- `go test ./internal/modules/system/service ./internal/transport/http ./internal/modules/system/handler ./pkg/web -count=1 -mod=readonly`
- `go test ./... -count=1 -mod=readonly`
- `go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main`
- `pnpm typecheck` from `web/admin`
- Playwright/Edge visual checks after Browser tooling was unavailable:
  `tmp/ai/media-desktop-1440x900-final.png`,
  `tmp/ai/media-import-url-1440x900-final.png`,
  `tmp/ai/media-mobile-390x844-final.png`, and
  `tmp/ai/media-mobile-390x844-viewport.png`.

### Active Slice: Version Management

Status: `[done]` started and completed on 2026-06-12.

Research completed before implementation:

- Demo route: `https://demo.gin-vue-admin.com/#/layout/admin/sysVersion`.
- Visual evidence:
  `tmp/ai/gva-version-2026-06-12.png`,
  `tmp/ai/gva-version-export-drawer-2026-06-12.png`, and
  `tmp/ai/gva-version-import-drawer-2026-06-12.png`.
- Demo page shape: list page filters by created date, version name, and version
  code; primary actions are create release package and import version package;
  row actions are view, download package, and delete. The create drawer collects
  version name/code/description and lets the user select menus, APIs, and
  dictionaries. The import drawer accepts a JSON package, previews its menus,
  APIs, and dictionaries, then imports it.
- Upstream primary source checked:
  `web/src/view/systemTools/version/version.vue`,
  `web/src/api/version.js`,
  `server/model/system/sys_version.go`,
  `server/model/system/request/sys_version.go`,
  `server/model/system/response/sys_version.go`,
  `server/api/v1/system/sys_version.go`,
  `server/service/system/sys_version.go`, and
  `server/router/system/sys_version.go` from
  `github.com/flipped-aurora/gin-vue-admin`.

Local implementation plan:

- Keep the GVA responsibility split, but name the local concept explicitly as a
  system release package. It snapshots menus, API routes, and dictionaries into
  a versioned JSON payload instead of representing the running binary version or
  migration version.
- Add an append-only `system_versions` migration and model with ID, version
  name, version code, description, JSON payload, package counts, source
  (`export` or `import`), creator/importer, created/updated/deleted timestamps.
- Expose protected APIs under `/api/v1/system/versions` for list, detail,
  source catalog, export, import, download, single delete, and batch delete.
- Add IAM permissions `version:read`, `version:create`, `version:import`,
  `version:download`, and `version:delete`, then wire them through API catalog,
  server-driven menus, and the role permission matrix.
- Build `/admin/versions` with the same low-noise table/filter/drawer workflow
  observed in the GVA demo: filters, selection, create package, import package,
  detail preview, JSON download, and batch delete.
- Preserve local architecture on import: dictionaries and dictionary items can
  be created idempotently when missing; menus and API routes are code/router
  owned in this scaffold, so imported menu/API entries are recorded in the
  package and reported as skipped until those catalogs become safely mutable.
- Update developer, maintainer, user, beginner, API, OpenAPI, and AI handoff
  docs. No example config change is expected unless a new configuration knob is
  introduced.

Validation plan:

- Run focused Go tests for `internal/modules/system` and
  `internal/transport/http`, then `go test ./... -count=1 -mod=readonly`.
- Run `go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main`.
- Run `pnpm typecheck` from `web/admin`.
- Use Browser visual checks on `/admin/versions` at `1440x900` and `390x844`,
  including create/import/detail workflows where possible.

Implementation completed:

- Added append-only migration
  `internal/migrations/20260612000500_create_system_versions.sql`, the
  `system_versions` model, repository methods, service workflow, handlers, and
  protected HTTP routes for source catalog, list, detail, export, import,
  download, single delete, and batch delete.
- Added IAM permissions `version:read`, `version:create`, `version:import`,
  `version:download`, and `version:delete`, then wired them into the route
  catalog, role permission matrix, and server-driven system menu.
- Added `/admin/versions` with GVA-like filters, table selection, create
  release package workflow, import JSON workflow, detail/JSON download support,
  and mobile responsive layout.
- Preserved local architecture on import: dictionaries and items are created
  idempotently when missing; menus and API routes remain code/router-owned and
  are reported as skipped while still stored in the package JSON.
- Updated API, OpenAPI, system module, README, onboarding, maintenance,
  extension, environment, overview, AI handoff, and example config notes.
- Fixed visual pollution found during QA: the version page's page-size filter,
  export dialog width/scroll behavior, and mobile filter actions now render
  without clipping or navigation overlap.

Validation completed:

- `go test ./... -count=1 -mod=readonly`
- `go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main`
- `pnpm typecheck` from `web/admin`
- Browser/CDP visual checks:
  `tmp/ai/version-desktop-1440x900-final.png`,
  `tmp/ai/version-export-dialog-1440x900-final.png`, and
  `tmp/ai/version-mobile-390x844-final.png`.

### Active Slice: API Token Management

Status: `[done]` started and completed on 2026-06-12.

Research completed before implementation:

- Demo route: API Token page under Super Admin. Screenshots:
  `tmp/ai/gva-api-token-2026-06-12.png` and
  `tmp/ai/gva-api-token-dialog-2026-06-12.png`.
- Demo page shape: filters by user ID and status; primary action is issue;
  table columns are ID, user, role ID, status, expires at, remark, operation;
  issue drawer asks for user, role, validity period, and remark; success dialog
  shows the token once; operation area includes curl example and invalidate.
- Upstream primary source checked:
  `web/src/view/systemTools/apiToken/index.vue`,
  `web/src/api/sysApiToken.js`,
  `server/model/system/sys_api_token.go`,
  `server/model/system/request/sys_api_token.go`,
  `server/api/v1/system/sys_api_token.go`,
  `server/service/system/sys_api_token.go`, and
  `server/router/system/sys_api_token.go` from
  `github.com/flipped-aurora/gin-vue-admin`.

Local implementation plan:

- Keep the GVA management workflow shape, but implement local API tokens as
  opaque secrets with a display prefix and SHA-256 hash. Do not store the
  plaintext token or copy GVA's raw JWT persistence.
- Put the backend in IAM because the token authenticates callers and belongs to
  users, organizations, and role/permission scope. Keep route registration in
  `internal/transport/http`.
- Add an append-only migration for `iam_api_tokens` with organization, user,
  optional role, name, remark, prefix, token hash, status, expiration, last-used
  metadata, and created/revoked audit columns.
- Expose protected organization-scoped APIs under
  `/api/v1/orgs/:orgId/api-tokens` for list, create, and revoke. List supports
  user, status, and pagination. Create returns the plaintext token only in the
  create response.
- Add permissions `api_token:read`, `api_token:create`, and
  `api_token:revoke`, then surface them in menu, API catalog, and role matrix.
- Add `/admin/api-tokens` using existing Aoi admin components and the same
  low-noise table/filter/drawer pattern as other system pages.
- Update docs for developer, maintainer, user, and beginner readers. Update
  example config only if implementation introduces a new config knob.

Implementation completed:

- Added `iam_api_tokens` through append-only migration
  `internal/migrations/20260612000400_create_iam_api_tokens.sql`.
- Added IAM repository, service, handler, auth fallback, and permission checks
  for opaque API tokens. Plaintext tokens are returned only by create responses;
  persisted records store prefix plus hash.
- Added organization-scoped protected APIs for list, create, and revoke under
  `/api/v1/orgs/:orgId/api-tokens`.
- Added route catalog entries, IAM permissions, server-driven menu entry, and
  `/admin/api-tokens`.
- Fixed the shared `AoiSelect` Material select value/display synchronization
  issue found during visual inspection.
- Updated API, IAM, onboarding, maintenance, extension, environment, OpenAPI,
  and AI handoff documentation. Example config comments now document that the
  refresh-token pepper also protects API token hashes.

Validation completed:

- `go test ./internal/modules/iam/... ./internal/transport/http/... ./internal/modules/system/... -count=1 -mod=readonly`
- `go test ./... -count=1 -mod=readonly`
- `go build -mod=readonly -o ./tmp/go-scaffold-server ./cmd/main`
- `pnpm typecheck` from `web/admin`
- Browser visual checks:
  `tmp/ai/api-token-desktop-1440x900-final.png`,
  `tmp/ai/api-token-issue-dialog-1440x900-final.png`,
  `tmp/ai/api-token-mobile-390x844-final.png`, and
  `tmp/ai/api-token-mobile-dialog-390x844-final.png`.

Follow-up note: no background cleanup job was added for expired tokens. List
and authentication paths compute expired state at read/auth time, which is
sufficient for the current management workflow.

## Visual Reference

- Shell: fixed left menu, top toolbar, visited tabs, dense white work surface.
- Navigation: menu groups expand in place; active item uses a strong blue block.
- Data pages: filters stay at the top, actions sit above the table, and tables
  favor compact row height with operation controls on the right.
- Dashboard: summary cards, chart/table regions, quick-entry panels, and notice
  lists are arranged as operational widgets rather than marketing cards.
- Styling: the demo keeps backgrounds mostly solid white, uses thin borders, and
  avoids blurred or translucent surfaces inside core management workflows.
- Visual pollution to avoid while replacing GVA incrementally: front-site
  background images, colorful navigation gradients, translucent glass panels,
  large marketing gradients, decorative blur, high-opacity watermark patterns,
  and low-contrast table/header text.

## Visual Review Rule

For future parity work, use screenshot or browser-based visual inspection before
and after implementation whenever a frontend change affects the UI, or a backend
change affects an admin workflow that users can see. Record the route, viewport,
and remaining risk in the handoff or final note.

Required minimum viewports for Aoi Admin visual work:

- Desktop: `1440x900`.
- Mobile: `390x844`.

When the local admin account requires MFA, record the blocked route and continue
with visual checks that do not require authenticated state until the code is
available.

## Backend Reference

Gin-Vue-Admin's server is organized around `api/v1`, `config`, `core`,
`global`, `initialize`, `middleware`, `model`, `model/request`,
`model/response`, `router`, `service`, `source`, and `utils`.

This repository should map that pattern into its existing boundaries instead of
renaming the backend wholesale:

- `api/v1` maps to `internal/modules/*/handler` plus
  `internal/transport/http`.
- `router` maps to `internal/transport/http` route registration.
- `service` maps to `internal/modules/*/service`.
- `model`, `request`, and `response` map to module-local model/DTO packages or
  `types/result` for shared envelopes.
- `initialize` maps to `internal/app/initapp`.
- `core` maps to `internal/app` plus reusable `pkg` infrastructure.
- `config` maps to `internal/config`.
- `middleware` maps to `internal/middleware`.
- `utils` maps to reusable packages under `pkg`.

Do not rename this repository to match GVA's folder names wholesale. The parity
target is the responsibility split: router catalog, API handler, service domain
rules, repository persistence, typed request/response shapes, initialization,
middleware, and reusable utilities.

GVA's router initialization separates public routes from private routes guarded
by JWT and Casbin. In this scaffold, the equivalent information is expressed in
the API catalog as `access=public|authenticated|permission` while the concrete
middleware remains in `internal/transport/http` and `internal/middleware`.

## Incremental Replacement Order

1. Stabilize the admin shell and table-page visual system.
2. Keep IAM pages aligned with the backend's existing organization, role, user,
   session, security, and audit APIs.
3. Add missing backend management modules only when the Go server exposes real
   models and routes.
4. Preserve the current dependency rule: modules depend on reusable `pkg`
   infrastructure, while `pkg` does not import application modules.
5. Avoid copying Gin-Vue-Admin's code generator, plugin market, or generated
   CRUD surface until this backend has an explicit product requirement for them.

## Implemented Parity Slices

- 2026-06-11: Admin visual cleanup plus static icon bundling.
- 2026-06-11: Server-driven admin menu groups at `/api/v1/system/menus`.
- 2026-06-11: HTTP API catalog at `/api/v1/system/apis`, mapped from the
  current router table.
- 2026-06-11: GVA-style API sync action at `/api/v1/system/apis/sync`, backed by
  `system_apis` when the migration has been applied and safely downgraded to
  live in-memory catalog refresh when the table is not available yet.
- 2026-06-11: API permission dictionary sync at
  `/api/v1/system/apis/permissions/sync`, deriving IAM permission records from
  registered backend routes so the role authorization page can bind them.
- 2026-06-11: Role authorization page changed from a flat permission list to a
  grouped permission matrix with object filters, keyword search, per-group bulk
  selection, and API-management handoff.
- 2026-06-12: Menu management catalog page added at `/admin/menus`, showing the
  server-driven menu groups, route paths, permission bindings, mobile entries,
  icons, and order values that back the admin shell.
- 2026-06-12: Dictionary management slice added with persisted
  `system_dictionaries` and `system_dictionary_items`, CRUD HTTP APIs, IAM
  permissions, role-matrix grouping, a server-driven menu entry, and the
  `/admin/dictionaries` management page.
- 2026-06-12: Operation history slice added after visually inspecting GVA's
  `操作历史` page: protected API requests are recorded into
  `system_operation_records`, surfaced through `/api/v1/system/operation-records`,
  wired into IAM permissions and server-driven menus, and managed from
  `/admin/operation-records` with GVA-style filters, selection, table layout, and
  pagination.
- 2026-06-12: Parameter management slice added after checking GVA's
  `参数管理` / `sys_params` model and service: persisted `system_parameters`
  records expose name, key, value, description, created timestamps, list filters,
  single and batch delete, key lookup, IAM permissions, server-driven menus, and
  the `/admin/parameters` management page.
- 2026-06-12: System configuration slice added after checking GVA's
  `系统配置` page and `/system/getSystemConfig` route: this scaffold now exposes
  a permission-protected `/api/v1/system/config` read-only runtime snapshot,
  masks secrets, wires `config:read` into IAM/menu/API catalogs, and adds the
  `/admin/system` grouped configuration page. GVA-style config write and service
  reload remain a later, higher-risk parity slice.
- 2026-06-12: Server status slice added after checking GVA's
  `/system/getServerInfo` service shape: this scaffold now exposes
  `/api/v1/system/server-info` with `server:read`, returns gopsutil-backed
  host CPU/RAM/disk metrics plus Go runtime, memory, GC, OS, uptime, and build
  metadata, wires the server-driven menu and role permission matrix, and adds
  `/admin/server-info`.
- 2026-06-12: Admin visual pollution hardening after visual comparison with
  GVA dashboard and menu-management pages: the admin runtime now clears legacy
  Aoi background/colorful-nav variables, and the admin CSS baseline uses a
  restrained GVA-like palette with solid panels, thin borders, low shadows,
  denser tables, muted login branding, semantic API method badges, isolated
  admin surface tokens, and desktop/mobile visual checks.
- 2026-06-12: GVA `source`/`initialize` parity slice: the System module can
  seed default dictionaries and parameters during startup through
  `system.seed_defaults_on_start`. The seed is idempotent, skips unavailable
  tables, and never overwrites existing user-edited parameter values.
- 2026-06-12: Login log parity slice after inspecting GVA
  `#/layout/admin/loginLog`: the demo currently exposes the menu item and tab
  but keeps dashboard content in the work surface, so this scaffold implements a
  usable `/admin/login-logs` page backed by IAM `auth.login` audit records and
  adds the server-driven menu entry under Security Audit.
- 2026-06-12: Error log parity slice after inspecting GVA
  `#/layout/admin/errorLog`: the public demo currently renders the unassigned
  route/permission page for the admin account, so this scaffold implements a
  usable `/admin/error-logs` page over `system_operation_records`. The backend
  keeps the existing operation-record table and adds optional `statusClass`
  filtering (`4xx`, `5xx`, or `error`) to `/api/v1/system/operation-records`;
  exact `status` filters still take priority when supplied.
- 2026-06-12: API catalog access-mode parity slice based on GVA's public vs
  JWT/Casbin-protected router groups: route catalog entries now expose
  `access` as `public`, `authenticated`, or `permission`, and the API management
  page can summarize and filter by that access mode without changing the
  append-only `system_apis` schema.
- 2026-06-12: Login captcha parity slice based on the GVA demo login screen:
  IAM now exposes public `GET /api/v1/auth/captcha`, validates optional
  `captchaId`/`captchaCode` during login when `auth.login_captcha_enabled=true`,
  keeps short-lived challenges in service memory, and renders the admin login
  captcha row only when the backend reports it enabled.
- 2026-06-12: API Token management parity slice after inspecting the GVA demo
  page and upstream `sys_api_token` source: this scaffold now stores
  organization-scoped opaque API tokens by hash, supports one-time plaintext
  display on issue, list/status filtering, revoke, API-token Bearer auth with
  role permission scope, server-driven menu/API catalog entries, and the
  `/admin/api-tokens` management page. During visual QA, the shared
  `AoiSelect` component was fixed so Material select values render reliably in
  desktop and mobile dialogs.
- 2026-06-12: Version management parity slice after inspecting the GVA demo
  `sysVersion` page and upstream `sys_version` source: this scaffold now stores
  versioned system release packages for selected menus, APIs, and dictionaries,
  exposes `/api/v1/system/versions` source/export/import/download/delete
  workflows, wires `version:*` IAM permissions and menu/API catalogs, adds the
  `/admin/versions` management page, and documents the safe local import rule
  where dictionaries are persisted while menu/API entries remain code-owned.
- 2026-06-12: Media library upload/download parity slice after inspecting the
  GVA demo upload page and upstream file/category source: this scaffold now
  stores media categories and assets, supports normal storage-backed upload,
  external URL import, keyword filtering, rename, download/open, delete, IAM
  `media:*` permissions, server-driven menus/API catalog entries, and the
  `/admin/media` management page. Breakpoint/chunk upload remains the preferred
  next slice because it needs a separate storage protocol and cleanup design.
