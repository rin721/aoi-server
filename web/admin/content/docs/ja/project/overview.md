---
title: プロジェクト概要
description: admin は Go サービスが既定で /admin に配信する Go-backed Nuxt 4 管理台です。
order: 10
category: project
navigation:
  icon: layout-dashboard
---

# プロジェクト概要

`admin` は現在の Go バックエンド向け Nuxt 4 管理台です。`useAdminApi()` を通して `/api/v1` を呼び出し、初期化/ログイン、組織、ユーザー、ロールと権限、API Token、セッション、セキュリティ監査、System 管理、メディア、バージョン、サーバー状態を扱います。Go サービスは生成済み静的アプリを既定で `/admin` に配信します。

## 技術スタック

- Nuxt 4、Vue 3、TypeScript、Composition API。
- クライアント状態は Pinia。
- `@nuxtjs/i18n` は三言語、デフォルト `zh-CN`、`no_prefix` ルーティング。
- `@nuxt/icon` はローカル Lucide アイコンを利用。
- Material Web はローカル Aoi wrapper からのみ公開。
- Nuxt Content が `/docs` の Markdown サイトを描画。

## プロダクト境界

Nuxt 内には本番バックエンド機能を実装しません。新しい管理機能は先に Go HTTP API で公開し、その後ページと DTO に接続します。`server/api/mock/`、`useAoiApi()`、動画再生、弾幕コンポーネントは過去のプロトタイプまたは Aoi コンポーネントライブラリ資産であり、現在の管理台主線ではありません。

長期的なプロダクト、アーキテクチャ、UI、API、インタラクション制約は `design/rules.md` に置きます。一時的な調査、プロトタイプ、段階計画を `design/` に残し続けないようにします。

## 主な流れ

現在の主な面は初期化/ログイン、ダッシュボード、組織、ユーザー、ロールと権限、API Token、セッション、セキュリティ設定、ログインログ、監査ログ、メニュー、API、辞書、パラメータ、システム設定、バージョン、メディア、サーバー状態です。ナビゲーションはサーバーメニューを優先し、取得できない場合はローカルのフォールバックを使います。
