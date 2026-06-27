# Scenario

**Feature**: `pre-commit add` outside a git repository always auto-installs globally.

```
# cwd outside repo — global install always
git-hooks pre-commit add (non-repo cwd) -> global dispatchers + global pre-commit.d
```

## Preconditions

- `req.FakeHome` is not inside a git repository.

## Steps

1. Set `req.WorkDir` to `req.FakeHome`.
2. Leaf `Setup` sets hook name/command.

```go
func Setup(t *testing.T, req *Request) error {
	req.WorkDir = req.FakeHome
	return nil
}
```