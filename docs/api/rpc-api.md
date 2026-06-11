# JSON-RPC API

RPC 入口随 `server` 进程装配，但使用独立端口监听。默认配置为关闭；启用后监听 `rpc.host:rpc.port`，示例默认值为 `127.0.0.1:10099`。

## 启用配置

```yaml
rpc:
  enabled: true
  host: 127.0.0.1
  port: 10099
  read_timeout: 10
  write_timeout: 10
  idle_timeout: 30
```

## 端点

- `POST /rpc`：JSON-RPC 2.0 单请求入口。
- `GET /health`：RPC 端口存活检查。

当前 MVP 不支持 batch 和 notification，请求必须包含 `id`。

## 内置方法

### system.ping

请求：

```json
{"jsonrpc":"2.0","id":1,"method":"system.ping","params":{"echo":"hi"}}
```

响应：

```json
{"jsonrpc":"2.0","id":1,"result":{"echo":"hi","ok":true}}
```

### system.methods

返回当前注册的方法名列表，按字典序排序。

```json
{"jsonrpc":"2.0","id":2,"method":"system.methods"}
```
