# Demo 模块

Demo 模块用于给开发者提供可运行的业务分层样板。当前包含两类示例：

- **Demo Todo**：公开 CRUD 示例，不要求登录，适合验证 Result 响应、路由、service/repository/model 分层。
- **客户列表（资源示例）**：受 IAM 保护的资源 CRUD 示例，用于演示后台菜单、权限码、当前主体、资源归属和可见范围。

生产配置应默认关闭 Demo，除非明确需要给内部开发环境保留示例入口。

## 位置

```text
internal/modules/demo
```

## 结构

| 文件类型 | 职责 |
| --- | --- |
| `model` | GORM model 和表结构 |
| `repository` | 通过 `pkg/database` 做持久化查询、分页、软删除 |
| `service` | 校验、事务编排、资源可见规则和领域错误 |
| `handler` | HTTP 绑定、当前 principal 读取、状态码选择和统一响应 |

## 路由

| 方法 | 路径 | 认证 | 权限 | 用途 |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/demo/todos` | 否 | 无 | 创建 Todo |
| GET | `/api/v1/demo/todos` | 否 | 无 | 查询 Todo 列表 |
| GET | `/api/v1/demo/todos/:id` | 否 | 无 | 查询单个 Todo |
| PUT | `/api/v1/demo/todos/:id` | 否 | 无 | 更新 Todo |
| DELETE | `/api/v1/demo/todos/:id` | 否 | 无 | 删除 Todo |
| POST | `/api/v1/demo/customers` | 是 | `customer:create` | 创建客户资源 |
| GET | `/api/v1/demo/customers` | 是 | `customer:read` | 查询客户资源列表 |
| GET | `/api/v1/demo/customers/:id` | 是 | `customer:read` | 查询单个客户资源 |
| PATCH | `/api/v1/demo/customers/:id` | 是 | `customer:update` | 更新客户资源 |
| DELETE | `/api/v1/demo/customers/:id` | 是 | `customer:delete` | 删除客户资源 |

后台页面：

```text
/admin/todos
/admin/customers
```

## 客户资源规则

客户资源会记录创建者信息：

- `owner_user_id`
- `owner_username`
- `owner_role_code`
- `org_id`

列表和单条查询只返回当前组织内可见的数据：

- 创建者本人始终可见；
- 当请求 principal 带有 `roleCode` 时，同组织、同 `owner_role_code` 的客户也可见；
- 普通网页登录 principal 通常不带 `roleCode`，因此不会因为空角色误看到其他人的客户。

这个规则使用当前 IAM 主体和角色上下文实现资源可见性，不引入额外的角色数据权限关联模型。

## Schema

Demo schema 通过 `pkg/sqlgen` 生成，由以下配置控制是否在服务启动时应用：

```yaml
demo:
  enabled: true
  apply_schema_on_start: true
```

当前会创建：

- `demo_todos`
- `demo_customers`

二次开发新增简单业务模块时，可以参考本模块：先定义 model，再补 repository、service、handler、路由、权限、菜单、测试和文档，避免把校验和事务写进 handler。
