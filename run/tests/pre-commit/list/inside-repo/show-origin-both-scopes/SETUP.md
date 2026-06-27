# Scenario

**Feature**: `pre-commit list --show-origin` prints both directory headers when both scopes are listed.

```
git-hooks pre-commit list --show-origin (repo cwd) -> local + global directory headers, then prefixed hook lines
```

## Preconditions

- Local and global hooks are seeded (see parent `inside-repo` setup).

## Steps

1. Run `git-hooks pre-commit list --show-origin` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ListArgs = []string{"--show-origin"}
	req.CaseName = "pre-commit list show-origin both scopes"
	return nil
}
```