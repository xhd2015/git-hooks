# Scenario

**Feature**: `pre-commit list` inside a git repository can show local hooks, global hooks, or both.

```
# seed hooks in both scopes
git-hooks pre-commit add (repo cwd) -> local-hook in local pre-commit.d
git-hooks pre-commit add (fake home cwd) -> global-hook in global pre-commit.d

# list from repo cwd
git-hooks pre-commit list -> both scopes (default)
```

## Preconditions

- `repoA` is an initialized git repository.
- A local hook and a global hook are registered before `list` runs.

## Steps

1. Set `req.WorkDir` to `req.RepoA`.
2. Add `local-hook` (`echo local`) from `repoA`.
3. Add `global-hook` (`echo global`) from `req.FakeHome` (non-repo cwd).
4. Leaf `Setup` sets `req.ListArgs` for the scope under test.

## Context

- Default `list` from a repo should include both scopes with `[local]` / `[global]` prefixes.
- `--local` or `--global` alone should show unprefixed single-scope output.

```go
import "fmt"

func Setup(t *testing.T, req *Request) error {
	req.WorkDir = req.RepoA
	env := isolatedEnv(req)
	if err := runCLI(env, req.RepoA, req.ToolPath, "pre-commit", "add", "local-hook", "echo", "local"); err != nil {
		return fmt.Errorf("seed local hook: %w", err)
	}
	if err := runCLI(env, req.FakeHome, req.ToolPath, "pre-commit", "add", "global-hook", "echo", "global"); err != nil {
		return fmt.Errorf("seed global hook: %w", err)
	}
	return nil
}
```