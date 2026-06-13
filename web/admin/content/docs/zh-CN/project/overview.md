---
title: 项目概览
description: admin 是 Go 后端驱动的 Nuxt 4 后台管理台，默认由 Go 服务挂载到 /admin。
order: 10
category: project
navigation:
  icon: layout-dashboard
---

# 项目概览

`admin` 是当前 Go 后端的 Nuxt 4 后台管理台。它通过 `useAdminApi()` 调用 `/api/v1`，覆盖登录/初始化、组织、用户、角色权限、API Token、会话、安全审计、System 管理、媒体库、版本和服务状态，并由 Go 服务默认挂载到 `/admin`。

## 技术栈

- Nuxt 4、Vue 3、TypeScript 和 Composition API。
- Pinia 管理客户端状态。
- `@nuxtjs/i18n` 提供三语界面，默认语言 `zh-CN`，策略 `no_prefix`。
- `@nuxt/icon` 使用本地 Lucide 图标集合。
- Material Web 只通过本地 Aoi wrapper 暴露给业务页面。
- Nuxt Content 为 `/docs` 渲染 Markdown 静态文档。

## 产品边界

应用不在 Nuxt 内实现生产后端。新增后台业务能力必须先由 Go HTTP API 暴露，再接入页面和 DTO。`server/api/mock/`、`useAoiApi()`、视频播放和弹幕相关组件是历史原型与 Aoi 组件库资产，不代表当前后台主线。

长期产品、架构、UI、API 或交互约束优先记录在 `design/rules.md`。临时研究、阶段计划和一次性说明不应长期堆在 `design/` 目录。

## 主要用户流

当前主路径是初始化/登录、工作台、组织、用户、角色权限、API Token、会话、安全设置、登录日志、审计日志、菜单、API、字典、参数、系统配置、版本、媒体库和服务器状态。导航优先读取服务端菜单；读取失败时使用本地兜底菜单。
