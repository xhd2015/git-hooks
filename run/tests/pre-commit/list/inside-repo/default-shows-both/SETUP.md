# Scenario

**Feature**: default `pre-commit list` inside a repo shows local then global hooks with scope prefixes.

```
git-hooks pre-commit list (repo cwd) -> [local] lines then [global] lines
```

## Preconditions

- Local and global hooks are seeded (see parent `inside-repo` setup).

## Steps

1. Run `git-hooks pre-commit list` with no scope flags from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = nil
	req.CaseName = "pre-commit list default shows both"
	return nil
}
```