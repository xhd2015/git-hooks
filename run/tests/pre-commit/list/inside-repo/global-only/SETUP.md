# Scenario

**Feature**: `pre-commit list --global` inside a repo shows only global hooks without scope prefixes.

```
git-hooks pre-commit list --global (repo cwd) -> unprefixed global hook lines only
```

## Preconditions

- Local and global hooks are seeded (see parent `inside-repo` setup).

## Steps

1. Run `git-hooks pre-commit list --global` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = []string{"--global"}
	req.CaseName = "pre-commit list global only"
	return nil
}
```