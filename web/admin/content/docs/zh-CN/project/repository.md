---
title: 仓库边界
description: 说明 app、admin API 类型、Aoi 组件、mock 遗留资产、i18n 和 design 等目录的职责。
order: 20
category: project
navigation:
  icon: folder-tree
---

# 仓库边界

仓库按后台应用、Go API 类型、Aoi 组件库、历史 mock 资产和长期设计文档分层。新增代码应先落到最贴近职责的目录。

## 应用代码

`app/` 是 Nuxt 后台应用主体，包含页面、布局、组件、composable、store、插件、样式和本地类型。业务页面优先使用 Nuxt 自动导入和本地 composable，避免引入不必要的全局工具。

`app/components/aoi/` 是 Material Web 与 Aoi 设计系统的边界。业务组件和页面不要直接使用 `md-*` 元素；需要新的 Material 行为时，先扩展 Aoi wrapper。

## 后台 API 契约

`app/config/admin-api.ts` 集中维护 Go 后端 endpoint，`app/types/admin.ts` 集中维护后台 DTO。已有类型时，不要在页面里临时拼接响应结构。

## Mock 和组件资产

`shared/`、`server/api/mock/`、视频播放和弹幕组件资料属于历史 Aoi 原型或组件库 demo。它们可以保留用于文档和组件展示，但新增后台业务功能优先接 Go API，不把持久化、权限或生产逻辑藏进 mock 层。

## 本地化与设计

`i18n/locales/` 维护 `zh-CN`、`en`、`ja` 三份用户可见文案。新增共享文案时三份同步。

`design/rules.md` 保存长期规则。短期实验、计划草案和调研结果应留在任务上下文或临时位置，避免污染长期设计目录。

## 生成目录

不要编辑 `.nuxt/`、`.output/`、`node_modules/` 等生成目录或依赖目录。依赖变化只通过 pnpm 有意产生，并保留对应 lockfile 更新。
