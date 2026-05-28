#!/usr/bin/env bash
# git-hooks managed pre-push start
if [ -z "${GIT_HOOKS_SESSION_ID:-}" ]; then
	GIT_HOOKS_SESSION_ID="$(date +%s)-$$"
	export GIT_HOOKS_SESSION_ID
fi

# NOTE: "$@" is "origin ssh://git@github.com/xhd2015/git-hooks"

git-hooks pre-push run "$@"
# git-hooks managed pre-push end
