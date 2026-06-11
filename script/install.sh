#!/usr/bin/env bash
set -euo pipefail

DEFAULT_REPO_URL="https://github.com/rin721/go-scaffold.git"
DEFAULT_REPO_REF="main"

die() {
	printf '[install] ERROR: %s\n' "$*" >&2
	exit 1
}

usage() {
	cat <<'USAGE'
Usage:
  curl -fsSL -o deploy.sh https://raw-githubusercontent-com-gh.helloworlds.eu.org/rin721/go-scaffold/main/script/install.sh
  bash deploy.sh --docker y --confirm [deploy options]

This bootstrap script clones the repository, then delegates to the repository
root deploy.sh with the same arguments. Use --repo and --ref to override the
default source:
  --repo https://github.com/rin721/go-scaffold.git
  --ref main
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

repo_url="$DEFAULT_REPO_URL"
repo_ref="$DEFAULT_REPO_REF"
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

require_cmd git
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/go-scaffold-install.XXXXXX")"
trap 'rm -rf "$work_dir"' EXIT

printf '[install] cloning %s (%s)\n' "$repo_url" "$repo_ref"
clone_repo "$repo_url" "$repo_ref" "$work_dir"
cd "$work_dir"

exec bash ./deploy.sh "${args[@]}"
