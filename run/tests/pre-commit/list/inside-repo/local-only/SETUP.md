# Scenario

**Feature**: `pre-commit list --local` inside a repo shows only local hooks without scope prefixes.

```
git-hooks pre-commit list --local (repo cwd) -> unprefixed local hook lines only
```

## Preconditions

- Local and global hooks are seeded (see parent `inside-repo` setup).

## Steps

1. Run `git-hooks pre-commit list --local` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = []string{"--local"}
	req.CaseName = "pre-commit list local only"
	return nil
}
```