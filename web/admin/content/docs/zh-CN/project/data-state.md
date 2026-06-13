---
title: API 与本地状态
description: 说明 useAdminApi、后台 DTO、会话存储、Pinia 和 localStorage hydrate 规则。
order: 50
category: project
navigation:
  icon: database
---

# API 与本地状态

当前后台以 Go API 和浏览器展示状态为主。代码要保持 `/api/v1` 契约意识，避免把临时 UI 结构当作后端响应结构扩散。

## API 访问

后台业务页面统一走 `useAdminApi()`。具体 endpoint 集中维护在 `app/config/admin-api.ts`，响应、请求和实体类型集中维护在 `app/types/admin.ts`。页面可以做展示层映射，但不要散落新的 `/api/v1` 字符串。

`useAoiApi()`、`useAoiApiTelemetry()`、`shared/` 和 `server/api/mock/` 只服务历史 Aoi 原型、组件 demo 或离线示例。新增后台业务能力不要扩展这些 mock 入口。

## 共享 DTO

面向 Go 后端的响应、请求和实体形状应放在 `app/types/admin.ts`。如果多个页面复用同一响应结构，先补类型，再接页面。

## 本地状态

认证 token pair 通过 `adminSession.ts` 写入 session storage，不迁移到持久 `localStorage`。本地 UI 偏好使用 `aoi.admin.ui.v1`，访问标签使用 `aoi.admin.visited-tabs.v1`。Pinia store 在客户端 hydrate 时必须能处理损坏的浏览器存储，并避免 SSR 崩溃。

## 错误与诊断

错误状态应暴露给页面，而不是只在 console 中丢失。用户可见共享文案需要进入三份 locale 文件；页面局部文案仍可在小范围内保持内联，但大幅触碰时应迁移到 locale。
