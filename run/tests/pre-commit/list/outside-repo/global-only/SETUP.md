# Scenario

**Feature**: `pre-commit list --global` outside a repo shows global hooks without prefixes.

```
git-hooks pre-commit list --global (non-repo cwd) -> unprefixed global hook lines
```

## Preconditions

- A global hook is seeded from non-repo cwd (see parent `outside-repo` setup).

## Steps

1. Run `git-hooks pre-commit list --global` from `req.FakeHome`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = []string{"--global"}
	req.CaseName = "pre-commit list outside repo global only"
	return nil
}
```