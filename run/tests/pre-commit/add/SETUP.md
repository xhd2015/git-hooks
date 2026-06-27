# Scenario

**Feature**: `pre-commit add` auto-installs dispatchers before registering managed hooks.

```
# add always runs full install first, then writes managed hook
git-hooks pre-commit add <name> <cmd> -> install (local or global) -> pre-commit.d/<name>
```

## Preconditions

- The CLI binary is built and on `PATH` as `git-hooks`.
- `HOME` and `XDG_CONFIG_HOME` point at an isolated fake home.

## Steps

1. Keep root setup from `run/tests/SETUP.md`.
2. Set `req.TestKind` to `"add"` and `req.Phase` to `"pre-commit"`.
3. Descendant nodes set cwd, flags, and hook payload.

## Context

- Auto-install installs **both** pre-commit and pre-push dispatchers (full install).
- Default inside a repo uses local install; `--global` uses global install and global storage.
- Outside a repo always uses global install regardless of flags.

```go
func Setup(t *testing.T, req *Request) error {
	req.TestKind = "add"
	req.Phase = "pre-commit"
	req.CaseName = "pre-commit add"
	return nil
}
```