# Aoi Design Rules

This file is the durable design and architecture rule entry for `admin`.
Keep `design/` focused on constraints that guide future implementation. Do not add temporary research notes, prototypes, phase plans, or stale mockups here.

## Product And Stack

- Aoi Web is the canonical Nuxt 4 admin console for the Go scaffold, mounted by the backend at `/admin`.
- The app uses Vue 3, TypeScript, Pinia, `@nuxt/icon`, optional `@nuxtjs/i18n`, and Material Web behind local Aoi wrappers.
- Use pnpm only. The repository package manager is `pnpm@10.22.0`.
- Current production data comes from the Go backend IAM, probe, and optional demo APIs. Do not add backend-missing management modules inside Nuxt.
- Go serves the static admin console from `web/admin/.output/public` at `/admin`. Use `pnpm generate` to produce that directory; `pnpm build` is a build check and does not replace static generation.
- Gin-Vue-Admin is a structural reference, not a visual clone target. Borrow its admin shell, breadcrumb/tab workflow, settings drawer, dense filters, table pages, and dashboard composition. Do not implement its code-generation assistant or plugin-system surface.

## Repository Boundaries

- Frontend app code lives in `app/`.
- Admin DTOs that mirror Go HTTP responses live in `app/types/admin.ts`.
- Legacy shared DTOs, mock endpoints, and content docs may remain while the framework is being reused, but new admin features should prefer real Go APIs over Nuxt mock routes.
- User-facing locale files live in `i18n/locales/`.
- Generated or dependency directories such as `.nuxt/`, `.output/`, and `node_modules/` are off limits.
- Prefer relative imports or Nuxt auto-imports. Do not add unnecessary global utilities.

## UI Foundation

- Business pages and feature components must not use `md-*` Material Web elements directly.
- Material Web imports stay centralized in `app/plugins/material-web.client.ts`.
- New Material Web behavior must be exposed through `app/components/aoi/` first.
- Aoi component semantics use `appearance` for visual form and `intent` for meaning. Do not add thin one-off wrappers such as secondary/danger/plain buttons when the shared component can express the role through this API.
- `AoiTextField` emits `enter` as a no-payload command event. Use its `keydown` event when a caller needs the raw `KeyboardEvent`.
- Plain text links, card links, tag links, and navigation links use `AoiLink`; business code should not use `NuxtLink` or bare `<a>`.
- Button-style navigation uses `AoiButton` or `AoiIconButton` with `to`/`href`.
- Use `app/assets/css/tokens.css` tokens and shared structure rules in `app/assets/css/main.css`.
- Prefer semantic tokens such as panel/card/control/nav/state/danger variables over one-off color literals.
- Admin status badges, method labels, and table states must use semantic intent tokens instead of page-local color literals so light, dark, and high-contrast modes stay readable.
- Use domain-specific radius tokens: `--aoi-radius-container`, `--aoi-radius-card`, `--aoi-radius-control`, `--aoi-radius-field`, `--aoi-radius-choice`, `--aoi-radius-nav-indicator`, and `--aoi-radius-round`. Reserve `999px` for navigation indicators, single-line pill controls, and true round/circular affordances; wide input/select fields must use capped field radii instead of control radii.
- Preserve responsive behavior, visible focus states, keyboard access, touch targets, text contrast, and `prefers-reduced-motion`.
- Icons should come from the local Lucide collection through `@nuxt/icon`.

## Layout Rules

- Desktop uses a Gin-Vue-Admin-inspired left menu, top toolbar, breadcrumb/title area, visited route tabs, and full-width admin work surface.
- Desktop sidebar may collapse to an icon rail; collapsed labels must remain available through title/accessible labels.
- Mobile uses a compact top bar and 56px bottom nav with four primary destinations: dashboard, organizations, users, and roles.
- Data pages use a filter strip, action toolbar, responsive table, empty/loading/error states, and dialogs or drawers for mutation workflows.
- Settings stay in a right-side drawer with local appearance preferences persisted under `aoi.admin.ui.v1`.
- Admin workflow surfaces must stay visually quiet: solid page backgrounds, solid white or dark panels, thin borders, low shadows, dense tables, and no glass/blur/orb/marketing decoration in core management pages.
- Admin runtime must isolate itself from legacy Aoi presentation settings. User background images, colorful navigation, scroll decoration, and other front-site preferences must not leak into authenticated admin workflows.
- Login and setup pages may show a branded side panel, but it should remain a restrained product shell rather than a promotional hero.

## Viewport And Lazy Loading Rules

- Use `useAoiInViewport()` for browser viewport detection. Keep it SSR safe and fall back to visible when `IntersectionObserver` is unavailable.
- Media-heavy or table-heavy surfaces should defer expensive rendering until the content is relevant to the current route or viewport.
- Long-running visual demos and decorative motion should pause when outside the viewport and respect `prefers-reduced-motion`.

## Layer Stack Rules

- Use semantic z-index tokens instead of ad hoc large numbers: background, page, sticky, nav, floating, menu, dialog, loading, and cursor.
- Dynamic overlays should register with `useAoiLayer()` through Aoi wrappers. Business pages should not manually compete with menu, dialog, loading, or cursor z-index values.
- Material Web overlay behavior must stay behind Aoi wrappers. Expose needed menu/dialog/select positioning through `app/components/aoi/` instead of reaching into `md-*` internals.
- Menus and selects should prefer a top-level positioning mode when they need to escape transformed, fixed, sticky, or overflow-hidden ancestors.

## Motion Performance Rules

- Prefer `transform` and opacity for motion. Horizontal and vertical movement should use `translate3d(...)` for card hover, fixed navigation feedback, and reusable motion scenes.
- Apply `will-change: transform` only to elements that actually animate or move, and reset it under reduced-motion where practical.
- Avoid creating persistent extra compositor layers for static decoration or large page sections.
- Use `AoiReveal` or `v-aoi-reveal` for reusable viewport pop-in motion. Content must remain visible before client hydration and when `IntersectionObserver` is unavailable.
- Prefer `AoiReveal` wrappers for cards or controls that already use `transform` for hover/press states. Use `v-aoi-reveal` on ordinary sections, panels, toolbars, and state blocks.
- Do not mix reveal transforms and hover/press transforms on the same element. Keep fixed navigation, menus, dialogs, loading layers, and other overlays outside reveal motion unless a wrapper specifically handles stacking and transform side effects.
- Repeated grids and lists should use reveal `index`/`stagger` sparingly so the UI feels responsive without delaying content access.
- Reveal motion is globally configurable from Settings / Preference. Defaults are enabled, contextual effect, repeat replay, 360ms duration, 18px distance, 35ms stagger, and 280ms max delay.
- Reveal setting ranges are duration 120-800ms, distance 0-48px, stagger 0-120ms, and max delay 0-600ms. Keep these bounds in the settings store and UI controls.
- When global reveal effect is contextual, local component/directive variants decide the effect. When a concrete global effect is selected, it overrides local reveal variants across the app.
- Disabling reveal motion must make reveal-enabled content immediately visible with no hidden state, transform offset, or visual transition. Do not expose engineering-only `rootMargin` or `threshold` controls in user settings.

## Data And API Rules

- Admin API access goes through `useAdminApi()` and must preserve the Go backend endpoints under `/api/v1`.
- Concrete backend endpoint paths live in `app/config/admin-api.ts`. Pages and composables should call named `ADMIN_API_ENDPOINTS` entries instead of adding new ad hoc `/api/v1` strings.
- Shared admin auto-refresh behavior goes through `useAdminAutoRefresh()` and `AdminAutoRefreshControls`. Generic defaults, timing units, clock cadence, manual click cooldown, and shared labels live in `app/config/admin-auto-refresh.ts`; page-specific overrides should be explicit config values, not inline timer numbers or copy.
- Manual refresh cooldown applies only to real browser click events. Programmatic refreshes after filters, pagination, mutations, route watches, and silent auto-refresh timers must not be skipped by click cooldown.
- Auto-refresh control spacing and height live in shared CSS tokens. Keep the control responsive through wrapping first; add new component-local breakpoints only after documenting why token-based spacing and natural wrapping are insufficient.
- Do not add menu/API/dictionary/parameter/code-generation backend concepts unless the Go backend exposes them first.
- JSON uses camelCase keys. Time fields use ISO 8601 UTC strings.
- Frontend error UI expects stable error payloads compatible with `ApiErrorPayload` in `app/types/admin.ts`.
- Browser-local settings are presentation preferences only and must not be sent to the backend.
- Browser-local stores must hydrate only on the client, recover from damaged `localStorage`, and avoid SSR crashes.
- Admin local preferences must never persist credentials, tokens, or private API payloads.
- Gin-Vue-Admin backend parity maps into this repository's existing layers: `api/v1` and router behavior belong in `internal/transport/http` plus module handlers, request validation and transactions belong in module services, persistence belongs in repositories, and reusable infrastructure remains under `pkg`.

## Local State Rules

- Admin UI preferences live under `aoi.admin.ui.v1`.
- Visited route tabs live under `aoi.admin.visited-tabs.v1` and should use session storage.
- Auth token pairs live in session storage through `adminSession.ts`; do not move them to persistent `localStorage`.

## i18n Rules

- Default locale is `zh-CN`; route strategy is `no_prefix`.
- New shared user-facing copy should update `zh-CN.json`, `en.json`, and `ja.json`.
- Existing inline Chinese copy may remain for narrow changes. If touching a large reusable surface, prefer moving reusable copy into locale files.

## Verification Rules

- After TypeScript, Vue, route, composable, or store changes, run `pnpm typecheck`.
- After Nuxt config, server route, runtime config, or build-sensitive changes, run `pnpm build`.
- Before handing off Go-hosted admin changes, run `pnpm generate` and confirm `web/admin/.output/public/index.html` exists.
- Visible UI changes must be checked with Browser/visual inspection at desktop and mobile widths. Use at least `1440x900` and `390x844`, and record the inspected routes, viewport sizes, and any remaining visual risk in the final handoff.
- Backend changes that alter an admin workflow, menu, table, form, permission surface, or response shape must also receive Browser/visual inspection of the affected frontend route before handoff.
- There is currently no committed lint script; do not claim lint was run unless a lint script is added or provided.
