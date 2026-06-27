# Scenario

**Feature**: `pre-push add` auto-installs dispatchers before registering managed hooks.

```
# add always runs full install first, then writes managed hook
git-hooks pre-push add <name> <cmd> -> install (local or global) -> pre-push.d/<name>
```

## Preconditions

- The CLI binary is built and on `PATH` as `git-hooks`.
- `HOME` and `XDG_CONFIG_HOME` point at an isolated fake home.

## Steps

1. Keep root setup from `run/tests/SETUP.md`.
2. Set `req.TestKind` to `"add"` and `req.Phase` to `"pre-push"`.
3. Descendant nodes set cwd, flags, and hook payload.

```go
func Setup(t *testing.T, req *Request) error {
	req.TestKind = "add"
	req.Phase = "pre-push"
	req.CaseName = "pre-push add"
	return nil
}
```