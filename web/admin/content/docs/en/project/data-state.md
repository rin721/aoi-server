---
title: API and Local State
description: useAdminApi, admin DTOs, session storage, Pinia, and localStorage hydration rules.
order: 50
category: project
navigation:
  icon: database
---

# API and Local State

The admin console is driven by the Go API and browser display state. Code should preserve the `/api/v1` contract and avoid spreading temporary display shapes as backend models.

## API Access

Admin business pages use `useAdminApi()`. Endpoint paths are centralized in `app/config/admin-api.ts`, and request, response, and entity types are centralized in `app/types/admin.ts`. Pages may map data for display, but should not scatter new `/api/v1` strings.

`useAoiApi()`, `useAoiApiTelemetry()`, `shared/`, and `server/api/mock/` are retained for legacy Aoi prototypes, component demos, or offline examples. New admin business behavior should not extend those mock entry points.

## Shared DTOs

Go backend request, response, and entity shapes belong in `app/types/admin.ts`. If multiple pages share a response shape, add the type first and then wire the pages.

## Local State

The auth token pair is written to session storage through `adminSession.ts`, not persistent `localStorage`. UI preferences use `aoi.admin.ui.v1`, and visited tabs use `aoi.admin.visited-tabs.v1`. Pinia stores must hydrate safely on the client, recover from damaged browser storage, and avoid SSR crashes.

## Errors and Diagnostics

Errors should be exposed to pages rather than disappearing into console output. Shared user-facing copy belongs in all three locale files; small page-local copy may remain inline until the area is touched more broadly.
