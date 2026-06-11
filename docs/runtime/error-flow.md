# 错误流

错误从 package/module 代码经过类型化错误和 result helper 转换为 HTTP 响应。

## 模块错误

repository 应返回存储层错误，例如 not found 或数据库失败。service 将这些错误归一化为模块语义，并应用业务校验。handler 再把错误转换为 HTTP
result 响应。

## 响应辅助

`types/result` 负责通用 JSON 响应辅助。模块 handler 使用它保持响应 envelope 一致。

## Panic Recovery

HTTP recovery middleware 捕获 panic，带 trace context 记录日志，并返回错误响应，避免进程崩溃。

## Trace ID 边界

middleware 使用 `traceId` 保存 request trace ID，`types/result` 也优先读取同一个键并兼容历史键 `trace_id`。panic recovery 和普通 result helper 都会把同一 trace ID 写入错误响应，便于 IAM 审计、日志和客户端错误上报关联。

## 就绪检查错误

`GET /ready` 会检查 database manager 是否存在以及能否 ping。它应被本地验证、Docker healthcheck 和部署脚本用于区分进程存活与服务就绪。
