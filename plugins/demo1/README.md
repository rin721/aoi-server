# demo1 RPC Plugin Example

`demo1` 是一个独立 Go module，用来演示外部插件如何通过 JSON-RPC 调用主服务。

## 目录结构

```text
plugins/demo1
|-- main.go              # 示例入口，编排一次 health/ping/methods 调用
|-- go.mod               # 插件自己的 Go module
`-- rpcclient
    |-- client.go        # HTTP 传输、/health、JSON-RPC 调用通用逻辑
    |-- protocol.go      # JSON-RPC 请求、响应、错误模型
    `-- system.go        # system.ping 和 system.methods 的类型化封装
```

## 运行

先启用主服务 RPC 配置：

```yaml
rpc:
  enabled: true
  host: 127.0.0.1
  port: 10099
```

启动主服务：

```powershell
go run ./cmd/main server
```

运行示例插件：

```powershell
cd plugins/demo1
go run .
```

可指定 RPC 地址和超时：

```powershell
go run . -rpc-url http://127.0.0.1:10099 -timeout 5s
```

## 设计意图

- `rpcclient/protocol.go` 只放 JSON-RPC 协议结构。
- `rpcclient/client.go` 只处理 HTTP、请求 ID、响应校验和错误包装。
- `rpcclient/system.go` 放具体方法的类型化封装。
- `main.go` 只展示插件如何使用客户端，不关心协议细节。
