# System 模块

`internal/modules/system` 承载后台系统管理能力，仍然遵循 `model -> repository -> service -> handler` 分层。HTTP 路由在 `internal/transport/http` 注册，权限码通过 API catalog 和 IAM 权限同步进入角色矩阵。

## 能力范围

| 能力 | 说明 |
| --- | --- |
| 菜单目录 | `/api/v1/system/menus` 返回当前用户可见菜单；菜单源在 `service` 内置目录中维护。 |
| API 目录 | `/api/v1/system/apis` 来自当前进程真实注册路由，可同步到 `system_apis` 表。 |
| 字典管理 | `system_dictionaries` 和 `system_dictionary_items` 存储可维护字典。 |
| 参数管理 | `system_parameters` 存储运行期可读参数。 |
| 操作记录 | `system_operation_records` 记录受保护 API 请求。 |
| 服务器状态 | `/api/v1/system/server-info` 暴露运行时、主机指标和构建信息快照。 |
| 版本发布包 | `system_versions` 存储菜单、API、字典的发布包 JSON，用于发版留痕、下载和跨环境导入。 |
| 媒体库 | `system_media_categories`、`system_media_assets`、断点上传会话和分片表存储分类、上传文件、外链资源和分片状态，二进制对象通过 `pkg/storage` 写入。 |

## 版本发布包

版本发布包按本项目边界管理菜单、API 和字典的配置快照：

- 导出时选择菜单、API 和字典，生成一份 JSON 包并保存到 `system_versions.version_data`。
- 菜单由代码内置目录生成，API 由路由目录生成，二者在导入时只记录和预览，不直接改写运行时代码或路由。
- 字典是数据库可维护数据，导入时会按字典 `code` 和字典项 `value` 幂等补齐缺失内容。
- 列表页只返回元数据和计数；详情、下载接口才返回完整包内容。

常用权限：

| 权限 | 用途 |
| --- | --- |
| `version:read` | 查看版本列表、详情和来源目录。 |
| `version:create` | 创建发版包。 |
| `version:import` | 导入发版包。 |
| `version:download` | 下载发版包 JSON。 |
| `version:delete` | 删除发版包记录。 |

## 媒体库

媒体库按本项目的 storage/IAM 边界实现上传、外链导入和资产管理：

- 分类是软删除树形数据，删除分类前必须没有子分类和媒体资源。
- 普通上传由服务端生成 `media/YYYY/MM/<id>.<ext>` 存储 key，原始文件名只作为展示字段保存，不参与路径拼接。
- 断点上传复用媒体库资产模型，先创建 `system_media_upload_sessions` 会话，再把分片写到 `media/chunks/<session-id>/`，完成时校验整文件 SHA-256 并合并为普通媒体资产。
- URL 导入只登记外链元数据，不抓取远程内容，避免把远程下载、类型探测和安全扫描混进导入请求。
- `storage.enabled=false` 时仍可查看数据库中的媒体记录和导入外链；普通上传、断点上传、本地文件下载和本地对象删除会返回 storage unavailable。
- 下载本地文件需要 IAM 鉴权，前端通过带 Bearer Token 的 blob 请求保存文件，不暴露匿名静态下载链接。

常用权限：

| 权限 | 用途 |
| --- | --- |
| `media:read` | 查看分类和媒体资源。 |
| `media:upload` | 普通上传和断点上传本地媒体文件。 |
| `media:import` | 导入外链媒体记录。 |
| `media:update` | 更新媒体名称和分类。 |
| `media:download` | 下载本地媒体对象。 |
| `media:delete` | 删除媒体资源。 |

## 维护注意

- 新增 System 持久化能力时，迁移应追加到 `internal/migrations`，不要改已共享迁移。
- 如果新增配置字段，同步更新 `configs/*.example.yaml`、`.env.example` 和 `docs/environment/configuration.md`。版本发布包没有新增配置开关；媒体库和断点上传复用现有 storage 配置。
- 断点上传会话默认 24 小时过期。过期会话不会继续接收分片；运维清理时应同时处理 `system_media_upload_sessions`、`system_media_upload_chunks` 和 `media/chunks/<session-id>/` 临时对象。
- 如果让菜单或 API 变成可导入的数据库数据，需要先重新设计来源优先级、冲突处理和启动同步策略，再改变版本导入行为。
- 管理页应保持低噪声：紧凑筛选区、清晰表格、薄边框、少装饰，避免把前台视觉风格带进后台工作流。
