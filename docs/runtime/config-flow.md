# 配置流程

配置是运行时状态。它先于基础设施和模块加载，并在 server 模式下被监听。

## 首次加载

```text
command flag/env/default path
  -> Manager.Load
  -> .env load
  -> YAML read
  -> placeholder substitution
  -> unmarshal
  -> env override
  -> validation
  -> atomic store
```

配置管理器向装配代码暴露当前配置。各包通常只接收自己需要的配置片段。

## 监听和重载

server 模式会注册配置监听器。文件变化后，管理器会影子加载新文件、应用环境变量覆盖、完成校验、写入存储，并调用变更处理器。

`reloadapp` 对比新旧配置并重载受影响的子系统：

- 日志；
- 数据库；
- 缓存；
- 执行器；
- HTTP 服务；
- 存储。

Demo 表结构应用属于启动行为，普通配置重载不会重新执行。

## 维护清单

1. 更新对应的 `internal/config` 结构体、标签和校验。
2. 更新 `configs` 和 `deploy` 下的示例。
3. 环境变量面向使用者时同步 `.env.example`。
4. 更新运行时文档。
5. 补充环境变量覆盖和配置校验测试。
6. 判断字段是否需要重载行为。
