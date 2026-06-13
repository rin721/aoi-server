#!/usr/bin/env bash
set -euo pipefail

DEFAULT_REPO_URL="https://github.com/rin721/go-scaffold.git"
DEFAULT_REPO_REF="main"
DEFAULT_REPO_SLUG="rin721/go-scaffold"

die() {
	printf '[install] ERROR: %s\n' "$*" >&2
	exit 1
}

usage() {
	cat <<'USAGE'
Usage:
  curl -fsSL -o install.sh https://raw.githubusercontent.com/rin721/go-scaffold/main/script/install.sh
  bash install.sh --docker y --confirm [deploy options]

GitHub proxy:
  curl -fsSL -o install.sh https://raw-githubusercontent-com-gh.helloworlds.eu.org/rin721/go-scaffold/main/script/install.sh
  bash install.sh --github-proxy-host github-com-gh.helloworlds.eu.org --docker y --confirm [deploy options]

This bootstrap script clones the repository, then delegates to the repository
root deploy.sh with the same arguments. Use --repo, --ref, or --github-proxy-host
to override the default source:
  --repo https://github.com/rin721/go-scaffold.git
  --ref main
  --github-proxy-host github-com-gh.helloworlds.eu.org
USAGE
}

require_arg() {
	local flag="$1"
	local value="${2:-}"
	[ -n "$value" ] || die "$flag requires a value"
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

repo_url_from_github_proxy() {
	local host="$1"

	host="${host#https://}"
	host="${host#http://}"
	host="${host%/}"
	[ -n "$host" ] || die "github proxy host cannot be empty"
	printf 'https://%s/%s.git' "$host" "$DEFAULT_REPO_SLUG"
}

clone_repo() {
	local repo_url="$1"
	local repo_ref="$2"
	local target_dir="$3"

	if git clone --depth 1 --branch "$repo_ref" "$repo_url" "$target_dir" >/dev/null 2>&1; then
		return 0
	fi

	rm -rf "$target_dir"
	git clone "$repo_url" "$target_dir" >/dev/null
	git -C "$target_dir" checkout "$repo_ref" >/dev/null
}

repo_url="${REPO_URL:-${DEPLOY_REPO_URL:-}}"
repo_ref="${REPO_REF:-${DEPLOY_REPO_REF:-$DEFAULT_REPO_REF}}"
github_proxy_host="${GITHUB_PROXY_HOST:-${DEPLOY_GITHUB_PROXY_HOST:-}}"
args=()

while [ "$#" -gt 0 ]; do
	case "$1" in
	--repo)
		require_arg "$1" "${2:-}"
		repo_url="$2"
		args+=("$1" "$2")
		shift 2
		;;
	--ref)
		require_arg "$1" "${2:-}"
		repo_ref="$2"
		args+=("$1" "$2")
		shift 2
		;;
	--github-proxy-host)
		require_arg "$1" "${2:-}"
		github_proxy_host="$2"
		args+=("$1" "$2")
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		args+=("$1")
		shift
		;;
	esac
done

if [ -z "$repo_url" ]; then
	if [ -n "$github_proxy_host" ]; then
		repo_url="$(repo_url_from_github_proxy "$github_proxy_host")"
	else
		repo_url="$DEFAULT_REPO_URL"
	fi
fi

require_cmd git
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/go-scaffold-install.XXXXXX")"
trap 'rm -rf "$work_dir"' EXIT

printf '[install] cloning %s (%s)\n' "$repo_url" "$repo_ref"
clone_repo "$repo_url" "$repo_ref" "$work_dir"
cd "$work_dir"

exec bash ./deploy.sh "${args[@]}"
