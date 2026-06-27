# Scenario

**Feature**: `pre-commit add` outside a repo auto-installs global dispatchers and stores the hook globally.

```
git-hooks pre-commit add outside-hook (non-repo cwd) -> global install -> global pre-commit.d/outside-hook
```

## Preconditions

- Cwd is `fakeHome`, not inside any git repository.

## Steps

1. Run `git-hooks pre-commit add outside-hook echo outside` from `fakeHome`.

```go
func Setup(t *testing.T, req *Request) error {
	req.HookName = "outside-hook"
	req.HookCmd = []string{"echo", "outside"}
	req.CaseName = "pre-commit add outside repo global"
	return nil
}
```