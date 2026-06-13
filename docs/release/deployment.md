# 部署说明

当前部署能力是生产风格示例，不是 v1 发布保证。真实环境使用前需要审查配置、密钥、数据库选择和回滚策略。

## 相关文件

| 路径 | 用途 |
| --- | --- |
| `Dockerfile` | 构建服务镜像 |
| `deploy/config.production.example.yaml` | 生产风格应用配置 |
| `deploy/docker-compose.production.example.yml` | Compose 服务定义 |
| `deploy.sh` | Bash 部署包装脚本 |
| `script/install.sh` | 远程安装入口，克隆仓库后委托仓库内 `deploy.sh` |
| `.github/workflows/deploy-remote.yml` | 手动触发的 GitHub Actions 远程部署 |

## 手动 Docker Compose 路径

```bash
export DEPLOY_IMAGE=go-scaffold:local
export AUTH_SIGNING_KEY=change-me-at-least-32-bytes-long
export AUTH_REFRESH_TOKEN_PEPPER=change-me-refresh-pepper
export AUTH_MFA_SECRET_KEY=change-me-mfa-secret-key-32-bytes
docker compose -f deploy/docker-compose.production.example.yml up -d
```

然后检查：

```bash
curl http://127.0.0.1:9999/health
curl http://127.0.0.1:9999/ready
curl http://127.0.0.1:9999/admin/server-info
```

## deploy.sh 路径

`deploy.sh` 可以克隆仓库或使用本地仓库、准备配置、构建或拉取镜像、运行 Compose，并检查健康、就绪和 Admin WebUI 静态路由。破坏性或类生产操作必须显式传入 `--confirm`。

该脚本应在 Linux Bash 环境运行。

常用 WebUI 参数：

| 参数 | 说明 |
| --- | --- |
| `--webui-mount-path /admin` | Go 静态托管挂载路径，运行时写入 `RIN_APP_WEBUI_MOUNT_PATH`；必须是非根绝对路径，不能是 `/`。 |
| `--webui-public-base-url /admin` | 后台公开基础路径，运行时写入 `RIN_APP_WEBUI_PUBLIC_BASE_URL`。 |
| `--webui-build-base-url /admin/` | Nuxt 构建 baseURL；未传时默认跟随 `--webui-mount-path`。 |
| `--webui-api-base-url ""` | Nuxt public API baseURL；空值表示同源调用。 |
| `--webui-show-demo-todo false` | 构建时是否显示 Demo Todo 兜底入口。 |
| `--webui-check y` | 部署后是否检查 WebUI 静态路由。 |
| `--webui-check-path /admin/server-info` | WebUI 静态路由检查路径；未传时默认 `<mount_path>/server-info`。 |

示例：

```bash
bash deploy.sh \
  --docker y \
  --image go-scaffold:local \
  --build y \
  --webui-mount-path /admin \
  --webui-build-base-url /admin/ \
  --webui-check-path /admin/server-info \
  --confirm
```

远程执行时，GitHub Actions 会把当前仓库的 `script/install.sh` 通过 SSH 传入远端 Bash；该入口会克隆目标 ref，再委托仓库内 `deploy.sh` 执行真实部署。直接在目标机器运行时，也可以下载 `deploy.sh` 后让它自行克隆仓库。

## 发布清单

1. 选择并验证生产数据库驱动。
2. 注入 `AUTH_SIGNING_KEY`、`AUTH_REFRESH_TOKEN_PEPPER`、`AUTH_MFA_SECRET_KEY` 等敏感值。
3. 运行 `db migrate status` 并在维护窗口执行 `db migrate up`。
4. 通过 `iam bootstrap-admin --password-stdin` 创建初始管理员。
5. 除非明确需要，否则关闭 Demo 路由。
6. 审查 CORS origins 和 headers，确保需要浏览器调用 IAM 时允许 `Authorization`。
7. 验证 `/health`、`/ready` 和 `/admin/server-info`。
8. 运行根模块测试。
9. 在干净环境构建 Docker 镜像。
10. 记录回滚、备份和迁移证据。
11. 如果属于托管任务，在对应运行时制品中记录部署证据。
## Admin WebUI 发布说明

生产镜像会从 `web/admin` 构建后台静态产物，并在运行时由 Go 服务挂载到 `/admin`。生产配置应保持：

```yaml
webui:
  enabled: true
  mount_path: /admin
  dist_dir: ./web/admin/.output/public
  public_base_url: ${WEBUI_PUBLIC_BASE_URL:/admin}
```

手动发布非 Docker 产物时，需要先在 `web/admin` 执行 `pnpm generate`，再将 `.output/public` 随服务一起部署。`pnpm build` 只作为构建检查，不替代静态托管产物。后台 UI 使用左侧导航、顶部工具栏、访问标签、筛选表格和管理抽屉；当前不发布代码生成、编程辅助或插件市场能力。

`NUXT_APP_BASE_URL` 是构建期配置。如果把后台从 `/admin` 改到其他非根挂载路径，需要重新构建前端静态产物或 Docker 镜像，并同步更新 `webui.mount_path`、`webui.public_base_url` 和部署脚本的 `--webui-build-base-url`。
