---
title: Nuxt 路由与布局
description: 说明后台页面路由、认证中间件、导航入口、布局壳和 docs 静态路由的关系。
order: 30
category: project
navigation:
  icon: route
---

# Nuxt 路由与布局

应用使用 Nuxt 文件路由。`login`、`setup`、注册、邀请和密码找回/重置是公开或半公开入口；其他后台页面需要 `auth.global.ts` 校验会话。`default` 布局承载后台壳层，`auth` 布局承载认证类页面。

## 主导航

`useAdminNavigation()` 优先从 Go 后端读取系统菜单，失败时回退到本地工作台、安全审计、系统管理和可选 Demo 分组。移动入口来自当前菜单中标记为 `mobile` 的前四项。

普通链接、卡片链接、标签链接和导航链接统一使用 `AoiLink`。按钮式导航使用 `AoiButton` 或 `AoiIconButton` 的 `to` / `href` 能力，由它们委托给 `AoiLink`。

## 文档内容

当前仓库保留 `web/admin/content/docs/**` 三语 Markdown、docs 渲染组件和 `nuxt.config.ts` 中 `/docs`、`/docs/**` 的 prerender 规则。当前静态生成会产出 `/docs`；如果要依赖 `/docs/project/...` 等子路由，应先确认页面入口和 `pnpm generate` 的实际预渲染输出。

`DocsPage` 使用以下 collection 映射。页面根据当前 locale 选择 collection，找不到对应语言时回退到中文 collection。

```ts
const collectionByLocale = {
  "zh-CN": "docsZhCn",
  en: "docsEn",
  ja: "docsJa"
}
```

## 静态渲染

`nuxt.config.ts` 对 `/docs` 与 `/docs/**` 设置 `prerender: true`。docs 页面在服务端收集导航路径并调用 `prerenderRoutes()`，让动态 Markdown slug 能进入静态构建。
