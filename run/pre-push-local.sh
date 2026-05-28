#!/bin/sh
# git-hooks managed pre-push start
if [ -z "${GIT_HOOKS_SESSION_ID:-}" ]; then
	GIT_HOOKS_SESSION_ID="$(date +%s)-$$"
	export GIT_HOOKS_SESSION_ID
fi
git-hooks pre-push run "$@"
# git-hooks managed pre-push end
