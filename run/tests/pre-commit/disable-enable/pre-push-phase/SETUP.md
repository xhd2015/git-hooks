# Scenario

**Feature**: pre-push disable/enable operates on pre-push managed storage, not pre-commit storage.

```
# phase selects the managed directory
git-hooks pre-push disable|enable push-check -> repo/.git/git-hooks/pre-push.d/01-push-check
```

## Preconditions

- The test runs from inside `repoA`.
- Both pre-commit and pre-push managed directories may exist.

## Steps

1. Set `req.Phase` to `pre-push`.
2. Descendant scenario seeds the pre-push hook.

## Context

- The pre-push runner is invoked directly through `git-hooks pre-push run`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "pre-push"
	req.WorkDir = req.RepoA
	return nil
}
```
