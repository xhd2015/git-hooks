# Scenario

**Feature**: default `pre-push list` inside a repo shows local then global hooks with scope prefixes.

```
git-hooks pre-push list (repo cwd) -> [local] lines then [global] lines
```

## Preconditions

- Local and global pre-push hooks are seeded (see parent `inside-repo` setup).

## Steps

1. Run `git-hooks pre-push list` with no scope flags from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "pre-push list default shows both"
	return nil
}
```