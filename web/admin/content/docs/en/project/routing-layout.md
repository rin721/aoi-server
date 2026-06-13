---
title: Nuxt Routing and Layout
description: Admin page routing, auth middleware, navigation entries, layout shells, and the static docs route.
order: 30
category: project
navigation:
  icon: route
---

# Nuxt Routing and Layout

The app uses Nuxt file-based routing. `login`, `setup`, signup, invitations, and password recovery/reset are public or semi-public entry points. Other admin pages require the `auth.global.ts` session check. The `default` layout hosts the admin shell, and the `auth` layout hosts authentication pages.

## Main Navigation

`useAdminNavigation()` prefers system menus from the Go backend and falls back to local workspace, security, system, and optional Demo groups when the API is unavailable. Mobile entries come from the first four current menu items marked as `mobile`.

Text links, card links, tag links, and navigation links use `AoiLink`. Button-style navigation uses `AoiButton` or `AoiIconButton` with `to` / `href`, delegated through `AoiLink`.

## Docs Content

The repository keeps `web/admin/content/docs/**` Markdown for three locales, docs rendering components, and `nuxt.config.ts` prerender rules for `/docs` and `/docs/**`. The current static build emits `/docs`; if `/docs/project/...` child routes are needed, confirm the page entry and the actual `pnpm generate` prerender output first.

`DocsPage` uses this collection mapping. The page chooses a collection from the active locale and falls back to the Chinese collection when a localized document is missing.

```ts
const collectionByLocale = {
  "zh-CN": "docsZhCn",
  en: "docsEn",
  ja: "docsJa"
}
```

## Static Rendering

`nuxt.config.ts` sets `prerender: true` for `/docs` and `/docs/**`. The docs page also collects navigation paths on the server and calls `prerenderRoutes()` so dynamic Markdown slugs are discovered during static builds.
