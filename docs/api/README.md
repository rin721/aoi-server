# API 参考

本目录保存当前服务重新生成的 API 契约。内容基于当前有效的路由注册、
handler、service 和 model 类型重建，不沿用旧路由记录或历史遗留说明。

## 文件

| 文件 | 用途 |
| --- | --- |
| `http-api.md` | 面向人阅读的中文 HTTP API 说明。 |
| `rpc-api.md` | JSON-RPC 独立入口和内置方法说明。 |
| `openapi.yaml` | health、ready、演示 Todo、客户资源示例、IAM 和 System 路由的 OpenAPI 3.0 契约。 |

## 事实来源

- 路由：`internal/transport/http/router.go`
- 演示 Todo 请求与响应：`internal/modules/demo/handler/todo.go`,
  `internal/modules/demo/model/todo.go`
- 客户资源示例请求与响应：`internal/modules/demo/handler/customer.go`,
  `internal/modules/demo/model/customer.go`
- IAM 请求与响应：`internal/modules/iam/handler/handler.go`,
  `internal/modules/iam/service/service.go`,
  `internal/modules/iam/model/model.go`
- 响应信封与错误码：`types/result`、`types/errors`
- JSON-RPC：`internal/transport/rpc`、`pkg/rpcserver`

## 当前接口面

- 公开探针：`GET /health`、`GET /ready`
- 演示 Todo CRUD：`/api/v1/demo/todos`
- 受保护客户资源示例：`/api/v1/demo/customers`，使用 `customer:*` 权限
- 公开 IAM 流程：登录、刷新令牌、找回/重置密码、接受邀请
- 受保护 IAM 流程：登出、切换组织、MFA、个人资料、组织、用户分页筛选、
  角色、权限、会话和审计日志
- 受保护 System 流程：菜单、API 目录、版本发布包、媒体库和断点上传
- JSON-RPC 独立入口：`POST /rpc`、`GET /health`，默认关闭

所有受保护的 IAM 路由都使用 `Authorization: Bearer <accessToken>`。
