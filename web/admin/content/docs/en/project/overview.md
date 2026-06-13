---
title: Project Overview
description: admin is a Go-backed Nuxt 4 admin console served by the Go service under /admin by default.
order: 10
category: project
navigation:
  icon: layout-dashboard
---

# Project Overview

`admin` is the current Nuxt 4 admin console for the Go backend. It calls `/api/v1` through `useAdminApi()` and covers setup/login, organizations, users, roles and permissions, API tokens, sessions, security audit, System management, media, versions, and server status. The Go service serves the generated static app under `/admin` by default.

## Stack

- Nuxt 4, Vue 3, TypeScript, and the Composition API.
- Pinia for client state.
- `@nuxtjs/i18n` with three locales, default `zh-CN`, and `no_prefix` routing.
- `@nuxt/icon` with local Lucide icons.
- Material Web exposed only through local Aoi wrappers.
- Nuxt Content for the `/docs` Markdown site.

## Product Boundary

The app does not implement production backend behavior inside Nuxt. New admin capabilities should be exposed by the Go HTTP API first, then wired into pages and DTOs. `server/api/mock/`, `useAoiApi()`, video playback, and danmaku components are legacy prototype or Aoi component-library assets, not the current admin product line.

Long-term product, architecture, UI, API, and interaction constraints belong in `design/rules.md`. Temporary research, prototypes, or phase plans should not accumulate in `design/`.

## Main Flows

The main surface is setup/login, dashboard, organizations, users, roles and permissions, API tokens, sessions, security settings, login logs, audit logs, menus, APIs, dictionaries, parameters, system config, versions, media, and server status. Navigation prefers server menus and falls back to local groups when the API is unavailable.
