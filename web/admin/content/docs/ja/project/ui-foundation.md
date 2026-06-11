---
title: UI、デザイントークン、モーション、レイヤー
description: Aoi UI のデザイントークン、レスポンシブ、モーション、レイヤー、アクセシビリティ制約。
order: 40
category: project
navigation:
  icon: palette
---

# UI、デザイントークン、モーション、レイヤー

Aoi UI はローカル wrapper、CSS デザイントークン、共有レイアウトルールを土台にしています。業務ページは既存の Aoi コンポーネントを使い、Material Web の内部実装に直接結合しないようにします。

## デザイントークン

色、角丸、影、サイズ、レイヤー、状態変数は `app/assets/css/tokens.css` と `app/assets/css/main.css` にあります。新しい視覚ルールは孤立した値を追加する前に変数を再利用します。

## Wrapper ルール

Material Web の import は `app/plugins/material-web.client.ts` に集約します。Aoi wrapper はサイズ、視覚的な意図、フォーカス、読み込み状態、リンク挙動、アクセシブルラベルを統一します。

```vue
<AoiButton icon="upload" intent="primary">
  公開
</AoiButton>
```

## モーション

インタラクションのモーションは `prefers-reduced-motion` を尊重し、状態伝達をモーションだけに頼らないようにします。Scroll、Reveal、Skeleton、弾幕、プレイヤー操作は動きを減らした環境でも理解できる必要があります。

## レイヤー

Dialog、Menu、浮遊面、ナビゲーション、読み込みレイヤーはローカルのレイヤールールで調整します。新しい浮遊 UI を作る前に `AoiDialog`、`AoiMenu`、`AoiLightboxGallery`、プレイヤーのオーバーレイを優先します。
