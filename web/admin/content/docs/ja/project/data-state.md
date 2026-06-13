---
title: API とローカル状態
description: useAdminApi、管理台 DTO、session storage、Pinia、localStorage hydrate のルール。
order: 50
category: project
navigation:
  icon: database
---

# API とローカル状態

管理台は Go API とブラウザ表示状態で動きます。コードは `/api/v1` 契約を保ち、一時的な表示用形状をバックエンドモデルのように広げないことが重要です。

## API アクセス

管理台の業務ページは `useAdminApi()` を使います。endpoint は `app/config/admin-api.ts` に集約し、request、response、entity 型は `app/types/admin.ts` に集約します。ページは表示用 mapping を持てますが、新しい `/api/v1` 文字列を散らさないようにします。

`useAoiApi()`、`useAoiApiTelemetry()`、`shared/`、`server/api/mock/` は過去の Aoi プロトタイプ、コンポーネント demo、オフライン例のために残しています。新しい管理台業務はこれらの mock 入口を拡張しません。

## 共有 DTO

Go バックエンドの request、response、entity 形状は `app/types/admin.ts` に置きます。複数ページで同じレスポンス形状を使う場合は、先に型を追加してからページを接続します。

## ローカル状態

認証 token pair は `adminSession.ts` から session storage に保存し、永続 `localStorage` には移しません。UI 設定は `aoi.admin.ui.v1`、訪問タブは `aoi.admin.visited-tabs.v1` を使います。Pinia store はクライアントで安全に hydrate し、壊れたブラウザストレージから復旧し、SSR crash を避ける必要があります。

## エラーと診断

エラーは console に消すだけでなく、ページに出せるようにします。共有のユーザー向け文言は三つの locale ファイルに置きます。小さなページ内文言は、広く触るまでは inline のままでも構いません。
