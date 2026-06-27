# Scenario

**Feature**: `pre-commit list` outside a git repository has no local scope; global hooks still list.

```
# only global hook seeded from non-repo cwd
git-hooks pre-commit add (fake home cwd) -> global-hook in global pre-commit.d

# list from non-repo cwd — no local hooks available
git-hooks pre-commit list -> global only (default)
git-hooks pre-commit list --local -> empty
```

## Preconditions

- `req.FakeHome` is not inside a git repository.
- A global hook is registered before `list` runs.

## Steps

1. Set `req.WorkDir` to `req.FakeHome`.
2. Add `global-hook` (`echo global`) from `req.FakeHome`.
3. Leaf `Setup` sets `req.ListArgs` for the scope under test.

## Context

- Outside a repo, local scope is unavailable; default listing should not emit `[local]` lines.
- `--local` alone should produce empty output without error.

```go
import "fmt"

func Setup(t *testing.T, req *Request) error {
	req.WorkDir = req.FakeHome
	env := isolatedEnv(req)
	if err := runCLI(env, req.FakeHome, req.ToolPath, "pre-commit", "add", "global-hook", "echo", "global"); err != nil {
		return fmt.Errorf("seed global hook: %w", err)
	}
	return nil
}
```