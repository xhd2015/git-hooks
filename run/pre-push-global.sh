#!/bin/sh
# git-hooks managed global pre-push start
set -eu

if [ -z "${GIT_HOOKS_SESSION_ID:-}" ]; then
	GIT_HOOKS_SESSION_ID="$(date +%s)-$$"
	export GIT_HOOKS_SESSION_ID
fi

git_hooks_common_dir="$(git rev-parse --git-common-dir 2>/dev/null || true)"
if [ -n "$git_hooks_common_dir" ]; then
	case "$git_hooks_common_dir" in
		/*) git_hooks_local_hook="$git_hooks_common_dir/hooks/pre-push" ;;
		*) git_hooks_local_hook="$(pwd)/$git_hooks_common_dir/hooks/pre-push" ;;
	esac
	if [ -x "$git_hooks_local_hook" ]; then
		"$git_hooks_local_hook" "$@"
	fi
fi

exec git-hooks pre-push run "$@"
# git-hooks managed global pre-push end
