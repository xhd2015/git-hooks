# Scenario

**Feature**: `pre-commit list` shows managed hooks from local and global storage with scope flags.

```
# cwd in repo or outside repo determines local availability
git-hooks pre-commit add (repo cwd) -> local pre-commit.d
git-hooks pre-commit add (non-repo cwd) -> global pre-commit.d

# list reads selected scopes and prints hook lines
git-hooks pre-commit list [--local] [--global] [--show-origin] -> stdout
```

## Preconditions

- The CLI binary is built and on `PATH` as `git-hooks`.
- `HOME` and `XDG_CONFIG_HOME` point at an isolated fake home.

## Steps

1. Inherit root setup (fake home, binary, repositories).
2. Set `req.Phase` to `pre-commit` and `req.TestKind` to `list`.
3. Descendant nodes seed hooks and set `req.WorkDir` / `req.ListArgs`.

## Context

- Local hooks live under `<repo>/.git/git-hooks/pre-commit.d/`.
- Global hooks live under `$XDG_CONFIG_HOME/.git-hooks/pre-commit.d/`.
- Adding hooks from inside a git repo writes local; adding from a non-repo directory writes global.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "pre-commit"
	req.TestKind = "list"
	req.CaseName = "pre-commit list"
	return nil
}
```