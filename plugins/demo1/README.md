# demo1 Sidecar Plugin

`demo1` 演示 Aoi Admin v1 的外部 sidecar 插件形态：

- `plugin.yaml` 是主服务读取的插件 manifest；
- `main.go` 启动独立 HTTP sidecar；
- `assets/remote.js` 是 Admin WebUI 动态加载的远程 ESM 微前端；
- `/api/hello` 通过主服务 `/api/v1/plugins/demo1/proxy/api/hello` 代理访问。

## 运行

在仓库根目录启用插件配置和代理签名密钥：

```powershell
$env:RIN_APP_PLUGINS_ENABLED="true"
$env:AOI_DEMO1_PLUGIN_SECRET="dev-demo1-secret-change-me"
```

启动主服务：

```powershell
go run ./cmd/main db migrate up --config=configs/config.yaml
go run ./cmd/main server --config=configs/config.yaml
```

另开一个终端启动 sidecar：

```powershell
cd plugins/demo1
go run .
```

登录后打开 `http://127.0.0.1:9999/admin/plugins/demo1`，即可看到远程模块通过代理读取 sidecar 数据。

## 协议边界

v1 插件协议以 manifest、HTTP 代理和 remote ESM 为主。目录中的 `rpcclient` 保留为旧 JSON-RPC 示例参考，不作为 v1 插件主协议。
