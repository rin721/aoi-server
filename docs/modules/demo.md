# Demo 模块

Demo 模块是一个小型 Todo 领域，用来展示推荐的模块形态。它适合脚手架、测试和示例，但生产配置应默认关闭，除非有明确需求。

## 位置

```text
internal/modules/demo
```

## 结构

| 文件类型 | 职责 |
| --- | --- |
| model | GORM model 和表结构 |
| repository | 通过 `pkg/database` 做 CRUD 持久化 |
| service | 校验、事务编排和领域规则 |
| handler | HTTP 绑定和响应转换 |

## 路由

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| POST | `/api/v1/demo/todos` | 创建 Todo |
| GET | `/api/v1/demo/todos` | 查询 Todo 列表 |
| GET | `/api/v1/demo/todos/:id` | 查询单个 Todo |
| PUT | `/api/v1/demo/todos/:id` | 更新 Todo |
| DELETE | `/api/v1/demo/todos/:id` | 删除 Todo |

## Schema

demo schema 通过 `pkg/sqlgen` 生成，可以由 DB CLI 打印或应用。启动时是否应用由以下配置控制：

```yaml
demo:
  enabled: true
  apply_schema_on_start: true
```

新增简单业务模块时，可以把此模块作为参考：model、repository、service、handler、测试、router 注册和文档都要闭环。
