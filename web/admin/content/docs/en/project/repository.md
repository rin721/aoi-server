---
title: Repository Boundaries
description: Responsibilities for app code, admin API types, Aoi components, legacy mock assets, i18n, design, and generated directories.
order: 20
category: project
navigation:
  icon: folder-tree
---

# Repository Boundaries

The repository is split by admin app code, Go API types, the Aoi component library, legacy mock assets, and long-lived design rules. New code should land in the closest matching boundary.

## App Code

`app/` contains the Nuxt admin app: pages, layouts, components, composables, stores, plugins, styles, and local types. Business pages should prefer Nuxt auto imports and local composables instead of new global utilities.

`app/components/aoi/` is the boundary between Material Web and the Aoi design system. Business components and pages should not use `md-*` elements directly; add or extend an Aoi wrapper first.

## Admin API Contract

`app/config/admin-api.ts` centralizes Go backend endpoints, and `app/types/admin.ts` centralizes admin DTOs. When a type exists, do not rebuild response shapes ad hoc inside pages.

## Mock and Component Assets

`shared/`, `server/api/mock/`, video playback, and danmaku docs are legacy Aoi prototype or component-library demo assets. They may remain for documentation and component examples, but new admin business features should use the Go API and should not hide persistence, authorization, or production behavior in mock routes.

## Localization and Design

`i18n/locales/` maintains user-facing copy for `zh-CN`, `en`, and `ja`. Shared copy changes should update all three files.

`design/rules.md` stores long-term rules. Short-lived notes and one-off plans should stay out of `design/`.

## Generated Directories

Do not edit `.nuxt/`, `.output/`, `node_modules/`, or other generated and dependency directories. Dependency changes should come from intentional pnpm commands and include the matching lockfile update.
