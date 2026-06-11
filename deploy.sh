#!/usr/bin/env bash
set -euo pipefail

DEFAULT_REPO_URL="https://github.com/rin721/go-scaffold.git"
DEFAULT_REPO_REF="main"

log() {
	printf '[deploy] %s\n' "$*"
}

die() {
	printf '[deploy] ERROR: %s\n' "$*" >&2
	exit 1
}

usage() {
	cat <<'USAGE'
用法:
  bash deploy.sh --docker y --confirm [options]

先克隆仓库:
  git clone <repo-address>
  cd <repo-address>
  bash deploy.sh --docker y --image go-scaffold:local --confirm

直接安装:
  curl -fsSL -o deploy.sh https://raw-githubusercontent-com-gh.helloworlds.eu.org/rin721/go-scaffold/main/script/install.sh
  bash deploy.sh --docker y --image go-scaffold:local --confirm

部署选项:
  --docker <y|n>              使用 Docker Compose 部署；当前仅支持 "y"。
  --repo <url>                脚本不在仓库目录内运行时要克隆的仓库地址。
  --ref <ref>                 要克隆的 Git ref，默认 main。
  --path <path>               运行目录，默认 /opt/go-scaffold。
  --image <image>             要构建或运行的镜像，默认 go-scaffold:local。
  --build <y|n>               是否从源码构建镜像，默认 y。
  --pull <y|n>                Compose up 前是否拉取镜像，默认 n。
  --port <port>               映射到容器 9999 端口的宿主机端口，默认 9999。
  --container-name <name>     Docker 容器名，默认 go-scaffold。
  --env <staging|production>  部署环境标签，默认 production。
  --confirm                   必需的确认标记。

镜像仓库选项:
  --registry-host <host>      镜像仓库地址，默认 ghcr.io。
  --registry-username <name>  docker login 可选用户名。
  --registry-token <token>    docker login 可选令牌。

应用选项:
  --db-driver <value>
  --db-host <value>
  --db-port <value>
  --db-user <value>
  --db-password <value>
  --db-name <value>
  --db-max-open-conns <value>
  --db-max-idle-conns <value>
  --redis-enabled <value>
  --redis-host <value>
  --redis-port <value>
  --redis-password <value>
  --redis-db <value>
  --redis-pool-size <value>
  --redis-min-idle-conns <value>
  --redis-max-retries <value>
  --redis-dial-timeout <value>
  --redis-read-timeout <value>
  --redis-write-timeout <value>
  --server-mode <value>
  --server-read-timeout <value>
  --server-write-timeout <value>
  --log-level <value>
  --log-format <value>
  --log-output <value>
  --i18n-default <value>
  --i18n-supported <value>
  --storage-enabled <value>
  --storage-fs-type <value>
  --storage-base-path <value>
  --storage-enable-watch <value>
  --storage-watch-buffer-size <value>
  --demo-enabled <value>
  --demo-apply-schema-on-start <value>
  --cors-enabled <value>
  --cors-allow-origins <value>
  --cors-allow-methods <value>
  --cors-allow-headers <value>
  --cors-expose-headers <value>
  --cors-allow-credentials <value>
  --cors-max-age <value>

安全提示:
  密码、令牌和密钥类参数可能出现在 shell 历史或进程列表中。
  优先使用受控 shell 会话、CI secret masking 或主机密钥管理服务。
  本脚本不会主动打印敏感值。
USAGE
}

require_arg() {
	local flag="$1"
	local value="${2:-}"
	[ -n "$value" ] || die "$flag requires a value"
}

normalize_yn() {
	local value="${1,,}"
	case "$value" in
	y | yes | true | 1) printf 'y' ;;
	n | no | false | 0) printf 'n' ;;
	*) die "expected y or n, got: $1" ;;
	esac
}

validate_value() {
	local key="$1"
	local value="$2"
	if [[ "$value" == *$'\n'* || "$value" == *$'\r'* ]]; then
		die "$key cannot contain newlines"
	fi
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

clone_repo() {
	local repo_url="$1"
	local repo_ref="$2"
	local target_dir="$3"

	require_cmd git
	log "cloning repository"
	if git clone --depth 1 --branch "$repo_ref" "$repo_url" "$target_dir" >/dev/null 2>&1; then
		return 0
	fi

	rm -rf "$target_dir"
	git clone "$repo_url" "$target_dir" >/dev/null
	git -C "$target_dir" checkout "$repo_ref" >/dev/null
}

DEPLOY_DOCKER=""
REPO_URL="$DEFAULT_REPO_URL"
REPO_REF="$DEFAULT_REPO_REF"
DEPLOY_PATH="/opt/go-scaffold"
DEPLOY_IMAGE="go-scaffold:local"
DEPLOY_BUILD="y"
DEPLOY_BUILD_SET="n"
DEPLOY_PULL="n"
APP_PORT="9999"
DEPLOY_CONTAINER_NAME="go-scaffold"
DEPLOY_ENV="production"
DEPLOY_CONFIRM="n"
SOURCE_DIR=""
REGISTRY_HOST="ghcr.io"
REGISTRY_USERNAME=""
REGISTRY_TOKEN=""
APP_ENV_PREFIX="${APP_ENV_PREFIX:-RIN_APP}"

declare -A APP_ENV=()

set_app_env() {
	local key="$1"
	local value="$2"
	validate_value "$key" "$value"
	APP_ENV["${APP_ENV_PREFIX}_${key}"]="$value"
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--docker)
		require_arg "$1" "${2:-}"
		DEPLOY_DOCKER="$(normalize_yn "$2")"
		shift 2
		;;
	--repo)
		require_arg "$1" "${2:-}"
		REPO_URL="$2"
		shift 2
		;;
	--ref)
		require_arg "$1" "${2:-}"
		REPO_REF="$2"
		shift 2
		;;
	--path)
		require_arg "$1" "${2:-}"
		DEPLOY_PATH="$2"
		shift 2
		;;
	--image)
		require_arg "$1" "${2:-}"
		DEPLOY_IMAGE="$2"
		shift 2
		;;
	--build)
		require_arg "$1" "${2:-}"
		DEPLOY_BUILD="$(normalize_yn "$2")"
		DEPLOY_BUILD_SET="y"
		shift 2
		;;
	--pull)
		require_arg "$1" "${2:-}"
		DEPLOY_PULL="$(normalize_yn "$2")"
		shift 2
		;;
	--port)
		require_arg "$1" "${2:-}"
		APP_PORT="$2"
		shift 2
		;;
	--container-name)
		require_arg "$1" "${2:-}"
		DEPLOY_CONTAINER_NAME="$2"
		shift 2
		;;
	--env)
		require_arg "$1" "${2:-}"
		DEPLOY_ENV="$2"
		shift 2
		;;
	--confirm)
		DEPLOY_CONFIRM="y"
		shift
		;;
	--source-dir)
		require_arg "$1" "${2:-}"
		SOURCE_DIR="$2"
		shift 2
		;;
	--registry-host)
		require_arg "$1" "${2:-}"
		REGISTRY_HOST="$2"
		shift 2
		;;
	--registry-username)
		require_arg "$1" "${2:-}"
		REGISTRY_USERNAME="$2"
		shift 2
		;;
	--registry-token)
		require_arg "$1" "${2:-}"
		REGISTRY_TOKEN="$2"
		shift 2
		;;
	--db-driver)
		require_arg "$1" "${2:-}"
		set_app_env DB_DRIVER "$2"
		shift 2
		;;
	--db-host)
		require_arg "$1" "${2:-}"
		set_app_env DB_HOST "$2"
		shift 2
		;;
	--db-port)
		require_arg "$1" "${2:-}"
		set_app_env DB_PORT "$2"
		shift 2
		;;
	--db-user)
		require_arg "$1" "${2:-}"
		set_app_env DB_USER "$2"
		shift 2
		;;
	--db-password)
		require_arg "$1" "${2:-}"
		set_app_env DB_PASSWORD "$2"
		shift 2
		;;
	--db-name)
		require_arg "$1" "${2:-}"
		set_app_env DB_NAME "$2"
		shift 2
		;;
	--db-max-open-conns)
		require_arg "$1" "${2:-}"
		set_app_env DB_MAX_OPEN_CONNS "$2"
		shift 2
		;;
	--db-max-idle-conns)
		require_arg "$1" "${2:-}"
		set_app_env DB_MAX_IDLE_CONNS "$2"
		shift 2
		;;
	--redis-enabled)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_ENABLED "$2"
		shift 2
		;;
	--redis-host)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_HOST "$2"
		shift 2
		;;
	--redis-port)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_PORT "$2"
		shift 2
		;;
	--redis-password)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_PASSWORD "$2"
		shift 2
		;;
	--redis-db)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_DB "$2"
		shift 2
		;;
	--redis-pool-size)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_POOL_SIZE "$2"
		shift 2
		;;
	--redis-min-idle-conns)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_MIN_IDLE_CONNS "$2"
		shift 2
		;;
	--redis-max-retries)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_MAX_RETRIES "$2"
		shift 2
		;;
	--redis-dial-timeout)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_DIAL_TIMEOUT "$2"
		shift 2
		;;
	--redis-read-timeout)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_READ_TIMEOUT "$2"
		shift 2
		;;
	--redis-write-timeout)
		require_arg "$1" "${2:-}"
		set_app_env REDIS_WRITE_TIMEOUT "$2"
		shift 2
		;;
	--server-mode)
		require_arg "$1" "${2:-}"
		set_app_env SERVER_MODE "$2"
		shift 2
		;;
	--server-read-timeout)
		require_arg "$1" "${2:-}"
		set_app_env SERVER_READ_TIMEOUT "$2"
		shift 2
		;;
	--server-write-timeout)
		require_arg "$1" "${2:-}"
		set_app_env SERVER_WRITE_TIMEOUT "$2"
		shift 2
		;;
	--log-level)
		require_arg "$1" "${2:-}"
		set_app_env LOG_LEVEL "$2"
		shift 2
		;;
	--log-format)
		require_arg "$1" "${2:-}"
		set_app_env LOG_FORMAT "$2"
		shift 2
		;;
	--log-output)
		require_arg "$1" "${2:-}"
		set_app_env LOG_OUTPUT "$2"
		shift 2
		;;
	--i18n-default)
		require_arg "$1" "${2:-}"
		set_app_env I18N_DEFAULT "$2"
		shift 2
		;;
	--i18n-supported)
		require_arg "$1" "${2:-}"
		set_app_env I18N_SUPPORTED "$2"
		shift 2
		;;
	--storage-enabled)
		require_arg "$1" "${2:-}"
		set_app_env STORAGE_ENABLED "$2"
		shift 2
		;;
	--storage-fs-type)
		require_arg "$1" "${2:-}"
		set_app_env STORAGE_FS_TYPE "$2"
		shift 2
		;;
	--storage-base-path)
		require_arg "$1" "${2:-}"
		set_app_env STORAGE_BASE_PATH "$2"
		shift 2
		;;
	--storage-enable-watch)
		require_arg "$1" "${2:-}"
		set_app_env STORAGE_ENABLE_WATCH "$2"
		shift 2
		;;
	--storage-watch-buffer-size)
		require_arg "$1" "${2:-}"
		set_app_env STORAGE_WATCH_BUFFER_SIZE "$2"
		shift 2
		;;
	--demo-enabled)
		require_arg "$1" "${2:-}"
		set_app_env DEMO_ENABLED "$2"
		shift 2
		;;
	--demo-apply-schema-on-start)
		require_arg "$1" "${2:-}"
		set_app_env DEMO_APPLY_SCHEMA_ON_START "$2"
		shift 2
		;;
	--cors-enabled)
		require_arg "$1" "${2:-}"
		set_app_env CORS_ENABLED "$2"
		shift 2
		;;
	--cors-allow-origins)
		require_arg "$1" "${2:-}"
		set_app_env CORS_ALLOW_ORIGINS "$2"
		shift 2
		;;
	--cors-allow-methods)
		require_arg "$1" "${2:-}"
		set_app_env CORS_ALLOW_METHODS "$2"
		shift 2
		;;
	--cors-allow-headers)
		require_arg "$1" "${2:-}"
		set_app_env CORS_ALLOW_HEADERS "$2"
		shift 2
		;;
	--cors-expose-headers)
		require_arg "$1" "${2:-}"
		set_app_env CORS_EXPOSE_HEADERS "$2"
		shift 2
		;;
	--cors-allow-credentials)
		require_arg "$1" "${2:-}"
		set_app_env CORS_ALLOW_CREDENTIALS "$2"
		shift 2
		;;
	--cors-max-age)
		require_arg "$1" "${2:-}"
		set_app_env CORS_MAX_AGE "$2"
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		die "unknown argument: $1"
		;;
	esac
done

[ "$DEPLOY_CONFIRM" = "y" ] || die "--confirm is required"
[ "$DEPLOY_DOCKER" = "y" ] || die "--docker y is required; non-Docker deployment is not implemented"

case "$DEPLOY_ENV" in
staging | production) ;;
*) die "--env must be staging or production" ;;
esac

[[ "$DEPLOY_PATH" = /* ]] || die "--path must be absolute"
[[ "$APP_PORT" =~ ^[0-9]+$ ]] || die "--port must be numeric"
validate_value REPO_URL "$REPO_URL"
validate_value REPO_REF "$REPO_REF"
validate_value DEPLOY_PATH "$DEPLOY_PATH"
validate_value DEPLOY_IMAGE "$DEPLOY_IMAGE"
validate_value DEPLOY_CONTAINER_NAME "$DEPLOY_CONTAINER_NAME"
validate_value APP_PORT "$APP_PORT"
validate_value SOURCE_DIR "$SOURCE_DIR"
validate_value REGISTRY_HOST "$REGISTRY_HOST"
validate_value REGISTRY_USERNAME "$REGISTRY_USERNAME"
validate_value REGISTRY_TOKEN "$REGISTRY_TOKEN"

if [ "$DEPLOY_PULL" = "y" ] && [ "$DEPLOY_BUILD_SET" = "n" ]; then
	DEPLOY_BUILD="n"
fi

if [ "$DEPLOY_BUILD" = "y" ] && [ "$DEPLOY_PULL" = "y" ]; then
	die "--build y and --pull y cannot be used together"
fi

require_cmd docker
if docker compose version >/dev/null 2>&1; then
	compose=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
	compose=(docker-compose)
else
	die "docker compose or docker-compose is required"
fi

cleanup_dir=""
if [ -n "$SOURCE_DIR" ]; then
	[ -f "$SOURCE_DIR/Dockerfile" ] || die "--source-dir does not look like a go-scaffold checkout"
elif [ -f "./Dockerfile" ] && [ -f "./deploy/docker-compose.production.example.yml" ]; then
	SOURCE_DIR="$(pwd)"
else
	cleanup_dir="$(mktemp -d "${TMPDIR:-/tmp}/go-scaffold.XXXXXX")"
	clone_repo "$REPO_URL" "$REPO_REF" "$cleanup_dir"
	SOURCE_DIR="$cleanup_dir"
fi

if [ -n "$cleanup_dir" ]; then
	trap 'rm -rf "$cleanup_dir"' EXIT
fi

COMPOSE_SOURCE="${SOURCE_DIR}/deploy/docker-compose.production.example.yml"
CONFIG_SOURCE="${SOURCE_DIR}/deploy/config.production.example.yaml"
COMPOSE_FILE="docker-compose.yml"
SERVICE_NAME="go-scaffold"
HEALTH_URL="http://127.0.0.1:${APP_PORT}/health"
READY_URL="http://127.0.0.1:${APP_PORT}/ready"

[ -f "$COMPOSE_SOURCE" ] || die "compose template not found: $COMPOSE_SOURCE"
[ -f "$CONFIG_SOURCE" ] || die "config template not found: $CONFIG_SOURCE"

mkdir -p "$DEPLOY_PATH/configs" "$DEPLOY_PATH/data" "$DEPLOY_PATH/logs"
if [ ! -f "$DEPLOY_PATH/configs/config.yaml" ]; then
	cp "$CONFIG_SOURCE" "$DEPLOY_PATH/configs/config.yaml"
	log "wrote default config template"
else
	log "kept existing config file"
fi
cp "$COMPOSE_SOURCE" "$DEPLOY_PATH/$COMPOSE_FILE"

if [ "$(id -u)" = "0" ]; then
	chown -R 10001:10001 "$DEPLOY_PATH/data" "$DEPLOY_PATH/logs"
elif command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
	sudo chown -R 10001:10001 "$DEPLOY_PATH/data" "$DEPLOY_PATH/logs"
else
	log "warning: cannot chown data/logs to 10001:10001 without passwordless sudo"
fi

export DEPLOY_IMAGE DEPLOY_CONTAINER_NAME APP_PORT
for key in "${!APP_ENV[@]}"; do
	export "$key=${APP_ENV[$key]}"
done

if [ -n "$REGISTRY_USERNAME" ] || [ -n "$REGISTRY_TOKEN" ]; then
	[ -n "$REGISTRY_USERNAME" ] || die "--registry-username is required with --registry-token"
	[ -n "$REGISTRY_TOKEN" ] || die "--registry-token is required with --registry-username"
	printf '%s' "$REGISTRY_TOKEN" | docker login "$REGISTRY_HOST" -u "$REGISTRY_USERNAME" --password-stdin >/dev/null
	log "registry login completed"
fi

log "source: $SOURCE_DIR"
log "target: $DEPLOY_PATH"
log "environment: $DEPLOY_ENV"
log "image: $DEPLOY_IMAGE"
if [ "${#APP_ENV[@]}" -gt 0 ]; then
	keys=("${!APP_ENV[@]}")
	IFS=,
	log "application env keys: ${keys[*]}"
	unset IFS
fi

if [ "$DEPLOY_BUILD" = "y" ]; then
	log "building image"
	docker build -t "$DEPLOY_IMAGE" "$SOURCE_DIR"
fi

cd "$DEPLOY_PATH"
if [ "$DEPLOY_PULL" = "y" ]; then
	log "pulling image"
	docker pull "$DEPLOY_IMAGE"
	"${compose[@]}" -f "$COMPOSE_FILE" pull "$SERVICE_NAME"
fi

log "starting service"
"${compose[@]}" -f "$COMPOSE_FILE" up -d "$SERVICE_NAME"

if command -v curl >/dev/null 2>&1; then
	check_url() {
		local name="$1"
		local url="$2"

		for _ in $(seq 1 30); do
			if curl -fsS --max-time 5 "$url" >/dev/null; then
				log "$name check passed"
				return 0
			fi
			sleep 2
		done

		die "$name check failed: $url"
	}

	check_url health "$HEALTH_URL"
	check_url ready "$READY_URL"
else
	log "warning: curl not found on host; skipped health/ready checks"
fi

log "deployment finished"
