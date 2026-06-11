# Legacy Admin WebUI

`web/admin` 是旧版 Nuxt 4 管理台源码，当前仅作为迁移回滚和对照参考保留，不再作为后台前端主线。

新的后台前端位于：

```text
web/aoi-web
```

Go 服务默认静态产物目录已经切换为：

```text
web/aoi-web/.output/public
```

如需开发、构建或静态生成当前后台，请进入 `web/aoi-web`：

```bash
cd web/aoi-web
pnpm install
pnpm dev
pnpm typecheck
pnpm build
pnpm generate
```

`pnpm generate` 生成 Go `/admin` 静态托管需要的 `.output/public/index.html`；`pnpm build` 只作为 Nuxt 构建检查。
