# Plugins 模块

Plugins 模块位于 `internal/modules/plugins`，用于把外部 sidecar 插件以受控方式暴露给后台和脚本调用。它不负责安装、打包或启动插件进程；插件进程应由 Compose、systemd、Kubernetes 或其他外部编排系统管理。

## 当前能力

| 能力 | 说明 |
| --- | --- |
| manifest 加载 | 从 `plugins.manifests` 配置的 JSON/YAML 文件读取插件定义。 |
| 插件列表 | `GET /api/v1/plugins` 返回当前用户可见插件。 |
| manifest 详情 | `GET /api/v1/plugins/{id}` 返回单个插件 manifest。 |
| 健康检查 | `GET /api/v1/plugins/{id}/health` 调用插件 sidecar 健康入口。 |
| 代理入口 | `ANY /api/v1/plugins/{id}/proxy/*path` 转发到插件 sidecar API。 |

## 权限和边界

插件接口只在 `plugins.enabled=true` 且 IAM 可用时注册。读取类接口需要 `plugin:read`，代理入口需要 `plugin:proxy`。API Token 调用方和浏览器用户走同一套 Bearer 认证与 Casbin domain RBAC，不存在插件专用超级通道。

manifest 只描述插件元数据、菜单、上游地址和代理/健康检查信息。不要把数据库迁移、密钥明文、进程启动命令或前端安装逻辑塞进 manifest。需要新增插件管理产品能力时，应先设计 manifest 字段、权限码、审计、超时、密钥注入和外部编排方式。

## 配置

常用配置字段：

```yaml
plugins:
  enabled: false
  manifests: []
  health_timeout_seconds: 3
  proxy_timeout_seconds: 30
```

本地和生产示例默认关闭插件。启用后要确保 manifest 路径对运行进程可读，并且插件 sidecar 地址只指向受信任网络。
