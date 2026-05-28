#!/usr/bin/env bash
# git-hooks managed global pre-commit start
set -eu

if [ -z "${GIT_HOOKS_SESSION_ID:-}" ]; then
	GIT_HOOKS_SESSION_ID="$(date +%s)-$$"
	export GIT_HOOKS_SESSION_ID
fi

# git-hooks check
# see: https://stackoverflow.com/questions/19387073/how-to-detect-commit-amend-by-pre-commit-hook
is_amend=$(ps -ocommand= -p $PPID | grep -e '--amend' || true)
# echo "is amend: $is_amend"
# args is always empty
# echo "args: ${args[@]}"
flags=()
if [[ -n $is_amend ]];then
    flags=("${flags[@]}" --amend)
fi

git_hooks_common_dir="$(git rev-parse --git-common-dir 2>/dev/null || true)"
if [ -n "$git_hooks_common_dir" ]; then
	case "$git_hooks_common_dir" in
		/*) git_hooks_local_hook="$git_hooks_common_dir/hooks/pre-commit" ;;
		*) git_hooks_local_hook="$(pwd)/$git_hooks_common_dir/hooks/pre-commit" ;;
	esac
	if [ -x "$git_hooks_local_hook" ]; then
		"$git_hooks_local_hook" "${flags[@]}"
	fi
fi

exec git-hooks pre-commit run "${flags[@]}"
# git-hooks managed global pre-commit end
