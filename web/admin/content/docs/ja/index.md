---
title: Aoi ドキュメント
description: Go バックエンド管理台、協作ルール、Aoi wrapper コンポーネントライブラリの静的ドキュメント内容入口。
order: 1
category: docs
navigation:
  icon: book-open
---

# Aoi ドキュメント

ここは `admin` の長期ドキュメント内容入口です。プロジェクトシステムでは Go-backed Nuxt 管理台、リポジトリ境界、状態、API、i18n、検証フローを説明します。コンポーネントライブラリでは `app/components/aoi/` のすべての Aoi wrapper を扱います。

## 入口

- [プロジェクト概要](/docs/project/overview) はアプリの目的、技術スタック、現在の Go API 境界を説明します。
- [リポジトリ境界](/docs/project/repository) は長期コードを置く場所と生成物の扱いを説明します。
- [コンポーネント概要](/docs/components/overview) は wrapper の原則と分類を説明します。
- [アクション](/docs/components/actions) はボタン、リンク、コマンド型ナビゲーションを扱います。
- [フォーム](/docs/components/forms) は入力、選択、アップロード、エディタを扱います。

## ドキュメント規約

すべての言語で同じ slug を使います。デフォルト言語は `zh-CN` で、i18n は `no_prefix` 戦略です。言語を切り替えても URL は変わらず、現在の locale に対応する Markdown collection を問い合わせます。

::docs-callout{title="静的レンダリング優先" intent="info" icon="sparkles"}
これらの Markdown は Nuxt Content が利用し、現在の静的ビルドは `/docs` を出力します。本番 API を追加せず、Go バックエンドの `/api/v1` 契約も変えません。`/docs/**` child route を追加または前提にする前に、`pnpm generate` で実際の prerender 結果を確認します。
::

## 執筆モデル

Markdown は説明、例、リンクを担当します。コンポーネント API、イベント、スロット、demo 入口は構造化メタデータから生成し、言語ごとに同じ表を重複管理しないようにします。
