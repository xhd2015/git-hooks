# Scenario

**Feature**: `pre-push add` from inside a git repository auto-installs local dispatchers by default.

```
# cwd inside repo — default local install
git-hooks pre-push add (repo cwd) -> local dispatchers + local pre-push.d
```

## Preconditions

- `repoA` is an initialized git repository.
- No prior `git-hooks install` or `add` has run.

## Steps

1. Set `req.WorkDir` to `req.RepoA`.
2. Leaf `Setup` sets hook name/command.

```go
func Setup(t *testing.T, req *Request) error {
	req.WorkDir = req.RepoA
	return nil
}
```