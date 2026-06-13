# 新增模块

新增应用能力时，优先放在 `internal/modules/<name>`，再通过 `internal/app/initapp` 装配。

## 推荐形态

```text
model -> repository -> service -> handler
```

只使用实际需要的层。很小的模块起步时不一定需要每个文件都齐全。

## 接线步骤

1. 在 `internal/modules/<name>` 下新增模块代码。
2. 如需配置，在 `internal/config` 中新增配置结构。
3. 在 `internal/app/initapp` 中装配 repository、service、handler。
4. 在 `internal/transport/http` 中注册路由。
5. 为 service 和路由行为补充聚焦测试。
6. 如果属于托管任务，同步更新工程文档和运行时证据。

## 身份和访问控制

当前 IAM 能力位于 `internal/modules/iam`，底层 JWT、Casbin、TOTP 和 goose 分别通过 `pkg/token`、`pkg/authorization`、`pkg/mfa`、`pkg/migrator` 封装。新增业务模块如需认证主体，应从 request context 读取 IAM middleware 写入的 Principal；如需权限控制，应通过 service 暴露的授权边界或 HTTP 权限中间件调用，不要在业务代码中直接依赖底层 JWT/Casbin 类型。

新增受保护 HTTP 接口时，需要同时考虑浏览器用户和 API Token 调用方。做法是为路由分配明确的权限码，例如 `report:read` 或 `report:export`，在 `internal/transport/http/router.go` 中通过权限中间件保护，并把权限纳入系统 API/权限同步。API Token 不拥有单独的超级通道，它只继承签发时绑定角色的权限；因此只要接口权限建模清楚，脚本调用和后台用户会走同一套授权边界。

如果新模块需要自己的长期凭据，不要另起一套明文 token 表。优先复用 IAM API Token 的 Bearer 认证，把业务范围表达成角色权限；只有在确实需要第三方回调签名、Webhook secret 或机器身份隔离时，才在模块内新增专用凭据模型，并补充迁移、轮换和泄漏撤销文档。

## 参考：扩展插件入口

Plugins 模块当前只负责读取 manifest、展示插件、执行健康检查和代理请求。新增插件能力时，不要把插件进程生命周期、安装、打包或市场逻辑塞进 Nuxt；先确定 manifest 字段、上游进程编排方式、权限码、健康检查和代理超时，再同步 Go 配置、System 权限和后台页面。
## 参考：扩展版本发布包

System 模块的版本发布包已经覆盖菜单、API 和字典的打包、下载和导入记录。新增模块如果希望参与发布包，优先提供稳定的编码字段，并在 `internal/modules/system/service/version.go` 中把该资源映射为可序列化结构。

当前菜单来自代码内置目录，API 来自路由目录，因此导入阶段不会改写菜单和 API。只有当某类资源已经具备明确的数据库来源、冲突策略和回滚策略时，才应把它加入“导入时可落库”的范围。

## 参考：扩展媒体库

System 媒体库把元数据放在 `system_media_assets`，把断点上传状态放在 `system_media_upload_sessions` 和 `system_media_upload_chunks`，把二进制对象放在 `pkg/storage`。新增图片裁剪、压缩或远程抓取时，不要复用 URL 导入的同步请求直接下载外部内容；应先设计任务状态、大小限制、MIME 校验、清理策略和权限码。

如果其他模块需要挂接媒体资源，优先只保存 `mediaAssetId` 或外链 URL，不要直接拼接 `storage_key` 路径。读取文件应通过 System service 或受保护下载 API 走 IAM 权限和 storage 安全边界。

Demo 模块现在有两个可复用样板：公开 Todo 适合学习最小 CRUD；客户资源示例适合学习后台菜单、IAM 权限、当前 principal、资源归属和分页列表。新增真正业务模块时，优先参考客户资源示例，把可见范围和业务校验放在 service 层，把请求绑定留在 handler 层。
