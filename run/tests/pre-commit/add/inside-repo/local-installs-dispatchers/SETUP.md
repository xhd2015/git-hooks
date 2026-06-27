# Scenario

**Feature**: first `pre-commit add` inside a repo auto-installs local dispatchers and registers the hook locally.

```
git-hooks pre-commit add test-hook (repo cwd) -> local install both dispatchers -> local pre-commit.d/test-hook
```

## Preconditions

- Fresh repo with no existing dispatcher scripts or managed hooks.

## Steps

1. Run `git-hooks pre-commit add test-hook echo test` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.HookName = "test-hook"
	req.HookCmd = []string{"echo", "test"}
	req.CaseName = "pre-commit add local install"
	return nil
}
```