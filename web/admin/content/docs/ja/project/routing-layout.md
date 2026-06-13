---
title: Nuxt ルーティングとレイアウト
description: 管理台ページルート、認証 middleware、ナビゲーション、レイアウトシェル、静的 docs ルートの関係。
order: 30
category: project
navigation:
  icon: route
---

# Nuxt ルーティングとレイアウト

アプリは Nuxt のファイルベースルーティングを使います。`login`、`setup`、signup、invitation、password recovery/reset は公開または半公開入口です。それ以外の管理台ページは `auth.global.ts` のセッション確認を通ります。`default` layout は管理台 shell、`auth` layout は認証ページを担当します。

## メインナビゲーション

`useAdminNavigation()` は Go バックエンドのシステムメニューを優先し、API が使えない場合はローカルの workspace、security、system、任意の Demo グループにフォールバックします。モバイル入口は現在メニューで `mobile` が付いた最初の四項目です。

テキストリンク、カードリンク、タグリンク、ナビゲーションリンクは `AoiLink` を使います。ボタン型ナビゲーションは `AoiButton` または `AoiIconButton` の `to` / `href` を使い、内部で `AoiLink` に委譲します。

## ドキュメント内容

現在のリポジトリには `web/admin/content/docs/**` の三言語 Markdown、docs rendering component、`nuxt.config.ts` の `/docs` と `/docs/**` prerender rule が残っています。現在の静的ビルドは `/docs` を出力します。`/docs/project/...` などの child route を前提にする場合は、page entry と `pnpm generate` の実際の prerender 出力を先に確認します。

`DocsPage` は次の collection mapping を使います。現在の locale から collection を選び、ローカライズ済み文書がない場合は中国語 collection にフォールバックします。

```ts
const collectionByLocale = {
  "zh-CN": "docsZhCn",
  en: "docsEn",
  ja: "docsJa"
}
```

## 静的レンダリング

`nuxt.config.ts` は `/docs` と `/docs/**` に `prerender: true` を設定します。docs ページはサーバー側でナビゲーションパスを集め、`prerenderRoutes()` を呼び出して Markdown slug を静的ビルドに含めます。
