# Scenario

**Feature**: `pre-push list` shows managed hooks from local and global storage with scope flags.

```
# same scope semantics as pre-commit list, different phase directory
git-hooks pre-push add -> pre-push.d
git-hooks pre-push list [--local] [--global] -> stdout
```

## Preconditions

- The CLI binary is built and on `PATH` as `git-hooks`.
- `HOME` and `XDG_CONFIG_HOME` point at an isolated fake home.

## Steps

1. Inherit root setup (fake home, binary, repositories).
2. Set `req.Phase` to `pre-push` and `req.TestKind` to `list`.
3. Descendant nodes seed hooks and set `req.WorkDir` / `req.ListArgs`.

## Context

- Local hooks live under `<repo>/.git/git-hooks/pre-push.d/`.
- Global hooks live under `$XDG_CONFIG_HOME/.git-hooks/pre-push.d/`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "pre-push"
	req.TestKind = "list"
	req.CaseName = "pre-push list"
	return nil
}
```