#!/bin/sh
# git-hooks managed pre-commit start
if [ -z "${GIT_HOOKS_SESSION_ID:-}" ]; then
	GIT_HOOKS_SESSION_ID="$(date +%s)-$$"
	export GIT_HOOKS_SESSION_ID
fi
git-hooks pre-commit run "$@"
# git-hooks managed pre-commit end
