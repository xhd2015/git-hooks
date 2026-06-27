# Scenario

**Feature**: `pre-commit add` from inside a git repository chooses local or global install via flags.

```
# cwd inside repo — default local install, --global selects global install
git-hooks pre-commit add (repo cwd) -> local or global dispatchers + managed hook storage
```

## Preconditions

- `repoA` is an initialized git repository.
- No prior `git-hooks install` or `add` has run unless a leaf overrides.

## Steps

1. Set `req.WorkDir` to `req.RepoA`.
2. Leaf `Setup` sets `AddArgs`, hook name/command, and any preconditions.

```go
func Setup(t *testing.T, req *Request) error {
	req.WorkDir = req.RepoA
	return nil
}
```