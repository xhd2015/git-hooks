# Scenario

**Feature**: `pre-commit list --local --global` matches default output inside a repo.

```
git-hooks pre-commit list --local --global (repo cwd) -> same as default list
```

## Preconditions

- Local and global hooks are seeded (see parent `inside-repo` setup).

## Steps

1. Run `git-hooks pre-commit list --local --global` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = []string{"--local", "--global"}
	req.CaseName = "pre-commit list both flags same as default"
	return nil
}
```