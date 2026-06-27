# Scenario

**Feature**: default `pre-commit list` outside a repo shows global hooks only (no local section).

```
git-hooks pre-commit list (non-repo cwd) -> unprefixed global hook lines only
```

## Preconditions

- A global hook is seeded from non-repo cwd (see parent `outside-repo` setup).

## Steps

1. Run `git-hooks pre-commit list` with no scope flags from `req.FakeHome`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = nil
	req.CaseName = "pre-commit list outside repo default global only"
	return nil
}
```