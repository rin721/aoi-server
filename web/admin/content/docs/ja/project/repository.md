---
title: リポジトリ境界
description: app、管理台 API 型、Aoi コンポーネント、mock 遺産、i18n、design、生成ディレクトリの責務。
order: 20
category: project
navigation:
  icon: folder-tree
---

# リポジトリ境界

リポジトリは管理台アプリ、Go API 型、Aoi コンポーネントライブラリ、過去の mock 資産、長期設計ルールで分かれています。新しいコードはもっとも近い責務の場所に置きます。

## アプリコード

`app/` は Nuxt 管理台本体です。ページ、レイアウト、コンポーネント、composable、store、plugin、style、ローカル型を含みます。業務ページでは Nuxt auto import とローカル composable を優先します。

`app/components/aoi/` は Material Web と Aoi デザインシステムの境界です。業務コンポーネントやページで `md-*` 要素を直接使わず、必要なら Aoi wrapper を追加または拡張します。

## 管理台 API 契約

`app/config/admin-api.ts` は Go バックエンド endpoint を集約し、`app/types/admin.ts` は管理台 DTO を集約します。既存の型がある場合、ページ内でレスポンス形状を作り直さないようにします。

## Mock とコンポーネント資産

`shared/`、`server/api/mock/`、動画再生、弾幕ドキュメントは過去の Aoi プロトタイプまたはコンポーネントライブラリ demo 資産です。ドキュメントやコンポーネント例として残せますが、新しい管理台業務機能は Go API を使い、永続化、認可、本番動作を mock route に隠しません。

## ローカライズと設計

`i18n/locales/` は `zh-CN`、`en`、`ja` のユーザー向け文言を管理します。共有文言を追加するときは三つすべてを更新します。

`design/rules.md` は長期ルールの場所です。短期メモや一度きりの計画は `design/` に残しません。

## 生成ディレクトリ

`.nuxt/`、`.output/`、`node_modules/` などの生成物や依存ディレクトリは編集しません。依存変更は pnpm で意図的に行い、対応する lockfile 更新を残します。
