# Scenario

**Feature**: first `pre-push add` inside a repo auto-installs local dispatchers and registers the hook locally.

```
git-hooks pre-push add push-hook (repo cwd) -> local install both dispatchers -> local pre-push.d/push-hook
```

## Preconditions

- Fresh repo with no existing dispatcher scripts or managed hooks.

## Steps

1. Run `git-hooks pre-push add push-hook echo push` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.HookName = "push-hook"
	req.HookCmd = []string{"echo", "push"}
	req.CaseName = "pre-push add local install"
	return nil
}
```