---
title: Aoi 文档
description: Go 后端管理台、协作规则和 Aoi wrapper 组件库的静态文档内容入口。
order: 1
category: docs
navigation:
  icon: book-open
---

# Aoi 文档

这里是 `admin` 的长期文档内容入口。内容分为两组：项目系统文档解释 Go-backed Nuxt 管理台、仓库边界、状态、API、i18n 和验证流程；组件库文档覆盖 `app/components/aoi/` 下的全部 Aoi wrapper。

## 快速入口

- [项目概览](/docs/project/overview) 说明应用目标、技术栈和当前 Go API 边界。
- [仓库边界](/docs/project/repository) 说明哪些目录可以承载长期代码，哪些目录属于生成产物。
- [组件总览](/docs/components/overview) 说明 Aoi wrapper 的使用原则和分类。
- [动作组件](/docs/components/actions) 展示按钮、链接和命令型导航。
- [表单组件](/docs/components/forms) 展示输入、选择、上传和编辑器。

## 文档约定

所有语言版本保持相同 slug。默认语言是 `zh-CN`，项目的 i18n 策略是 `no_prefix`，所以切换语言时路径保持不变，页面内容随当前 locale 查询对应 Markdown collection。

::docs-callout{title="静态优先" intent="info" icon="sparkles"}
这些 Markdown 内容由 Nuxt Content 消费，当前静态生成会产出 `/docs`。它们不新增生产 API，也不改变 Go 后端的 `/api/v1` 契约；新增或依赖 `/docs/**` 子路由前需要用 `pnpm generate` 验证实际预渲染结果。
::

## 编写方式

Markdown 负责解释、示例和跨页链接。组件 API、事件、插槽和 demo 入口来自结构化元数据，避免同一张表在多种语言里重复维护。
