---
title: Validation and Collaboration
description: pnpm commands, verification boundaries, Git collaboration, and dirty worktree protection.
order: 70
category: project
navigation:
  icon: check-circle
---

# Validation and Collaboration

The repository uses pnpm only. The declared package manager version is `pnpm@10.22.0`, and common commands run from the project root.

## Commands

```bash
pnpm install
pnpm dev
pnpm typecheck
pnpm build
pnpm generate
```

The repository does not currently have a committed `lint` script. Do not claim lint verification unless a script is added later or the task explicitly provides one.

## When to Verify

Run `pnpm typecheck` after changing TypeScript, Vue, routes, composables, or stores. Run `pnpm build` after changing Nuxt config, server routes, runtime config, or build-sensitive modules. When the Go service must serve the static app, run `pnpm generate` and confirm `.output/public/index.html` exists.

Visible UI changes should be checked in a browser on desktop and mobile widths, especially text wrapping, focus states, drawers, overlays, and small-screen layouts.

## Git Collaboration

Check worktree status before editing. Do not revert user changes or unrelated dirty files. Do not commit, create branches, or push unless the user explicitly asks.
