---
title: Aoi Docs
description: Static documentation content for the Go-backed admin console, collaboration rules, and Aoi wrapper component library.
order: 1
category: docs
navigation:
  icon: book-open
---

# Aoi Docs

This is the long-lived documentation content entry for `admin`. The project system pages explain the Go-backed Nuxt admin console, repository boundaries, state, API, i18n, and validation workflow. The component library pages cover every Aoi wrapper under `app/components/aoi/`.

## Fast Paths

- [Project overview](/docs/project/overview) explains the app goals, stack, and current Go API boundary.
- [Repository boundaries](/docs/project/repository) explains where long-lived code belongs and which folders are generated.
- [Component overview](/docs/components/overview) explains the wrapper rules and categories.
- [Actions components](/docs/components/actions) covers buttons, links, and command navigation.
- [Forms components](/docs/components/forms) covers inputs, selections, uploads, and editors.

## Documentation Rules

All locales keep the same slugs. The default locale is `zh-CN`, and the app uses the i18n `no_prefix` strategy, so switching language keeps the route stable while the page queries a different Markdown collection.

::docs-callout{title="Static first" intent="info" icon="sparkles"}
This Markdown content is consumed by Nuxt Content, and the current static build emits `/docs`. It does not add production APIs or change the Go backend `/api/v1` contract; before adding or relying on `/docs/**` child routes, verify the actual prerender output with `pnpm generate`.
::

## Authoring Model

Markdown carries narrative, examples, and links. Component APIs, events, slots, and demo entry points come from structured metadata so tables stay consistent across locales.
