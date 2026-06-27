# Scenario

**Feature**: `pre-commit list --local` outside a repo produces empty output without error.

```
git-hooks pre-commit list --local (non-repo cwd) -> empty stdout
```

## Preconditions

- A global hook exists but local scope is unavailable outside a git repo.

## Steps

1. Run `git-hooks pre-commit list --local` from `req.FakeHome`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = []string{"--local"}
	req.CaseName = "pre-commit list outside repo local empty"
	return nil
}
```