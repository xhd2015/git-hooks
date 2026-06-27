# Scenario

**Feature**: `pre-push add` outside a repo auto-installs global dispatchers and stores the hook globally.

```
git-hooks pre-push add outside-push (non-repo cwd) -> global install -> global pre-push.d/outside-push
```

## Preconditions

- Cwd is `fakeHome`, not inside any git repository.

## Steps

1. Run `git-hooks pre-push add outside-push echo outside` from `fakeHome`.

```go
func Setup(t *testing.T, req *Request) error {
	req.HookName = "outside-push"
	req.HookCmd = []string{"echo", "outside"}
	req.CaseName = "pre-push add outside repo global"
	return nil
}
```