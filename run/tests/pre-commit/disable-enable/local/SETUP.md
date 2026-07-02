# Scenario

**Feature**: local pre-commit disable/enable operates on the current repository's managed pre-commit directory.

```
# local scope is selected by running inside a git repository without --global
repo cwd -> git-hooks pre-commit disable|enable -> repo/.git/git-hooks/pre-commit.d
```

## Preconditions

- The test runs from inside `repoA`.
- Local managed pre-commit storage is isolated to `repoA`.

## Steps

1. Keep `req.Phase` as `pre-commit`.
2. Keep `req.WorkDir` as `repoA`.
3. Descendant scenarios choose single-hook or all-hook target mode.

## Context

- Local operations must not need global configuration.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "pre-commit"
	req.WorkDir = req.RepoA
	return nil
}
```
