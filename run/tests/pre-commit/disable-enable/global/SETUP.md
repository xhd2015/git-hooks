# Scenario

**Feature**: `--global` selects global managed hook storage even when running inside a repository.

```
# same display name exists in both scopes
local pre-commit.d/01-check
global pre-commit.d/01-check

# --global selects only global storage
git-hooks pre-commit disable --global check -> global pre-commit.d/01-check
```

## Preconditions

- The test runs inside `repoA`.
- Local and global pre-commit hooks can have the same display name.

## Steps

1. Keep `req.Phase` as `pre-commit`.
2. Descendant scenario seeds matching local and global hooks.

## Context

- The local hook is a guard: its executable mode must not change.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "pre-commit"
	req.WorkDir = req.RepoA
	return nil
}
```
