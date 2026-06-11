# Gin-Vue-Admin Parity Notes

Verified on 2026-06-11 from the public demo and upstream documentation.
Sources:

- https://demo.gin-vue-admin.com
- https://www.gin-vue-admin.com/guide/server/
- https://github.com/flipped-aurora/gin-vue-admin/tree/main/server

## Visual Reference

- Shell: fixed left menu, top toolbar, visited tabs, dense white work surface.
- Navigation: menu groups expand in place; active item uses a strong blue block.
- Data pages: filters stay at the top, actions sit above the table, and tables
  favor compact row height with operation controls on the right.
- Dashboard: summary cards, chart/table regions, quick-entry panels, and notice
  lists are arranged as operational widgets rather than marketing cards.
- Styling: the demo keeps backgrounds mostly solid white, uses thin borders, and
  avoids blurred or translucent surfaces inside core management workflows.

## Visual Review Rule

For future parity work, use screenshot or browser-based visual inspection before
and after implementation whenever a frontend change affects the UI, or a backend
change affects an admin workflow that users can see. Record the route, viewport,
and remaining risk in the handoff or final note.

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
