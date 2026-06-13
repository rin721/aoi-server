---
title: 検証と協作
description: pnpm コマンド、検証境界、Git 協作、未整理の作業ツリーの保護。
order: 70
category: project
navigation:
  icon: check-circle
---

# 検証と協作

リポジトリでは pnpm だけを使います。宣言された package manager は `pnpm@10.22.0` で、よく使うコマンドはプロジェクトルートから実行します。

## コマンド

```bash
pnpm install
pnpm dev
pnpm typecheck
pnpm build
pnpm generate
```

現在このリポジトリにはコミット済みの `lint` script がありません。後で追加されるか、タスクで明示されない限り、lint 検証済みとは言いません。

## 検証タイミング

TypeScript、Vue、ルート、composable、store を変更したら `pnpm typecheck` を実行します。Nuxt 設定、server route、runtime config、ビルドに敏感な module を変更したら `pnpm build` を実行します。Go サービスで静的配信する必要がある場合は `pnpm generate` を実行し、`.output/public/index.html` が存在することを確認します。

見える UI 変更はできるだけブラウザでデスクトップ幅とモバイル幅を確認します。特にテキスト折り返し、フォーカス、ドロワー、オーバーレイ、小画面レイアウトを見ます。

## Git 協作

編集前に作業ツリーの状態を確認します。ユーザー変更や無関係な未整理ファイルは戻しません。ユーザーが明示しない限り、コミット、ブランチ作成、push はしません。
