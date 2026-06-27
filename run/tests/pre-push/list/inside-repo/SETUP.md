# Scenario

**Feature**: default `pre-push list` inside a git repository shows both local and global hooks.

```
git-hooks pre-push add (repo cwd) -> local-hook
git-hooks pre-push add (fake home cwd) -> global-hook
git-hooks pre-push list (repo cwd) -> [local] then [global] lines
```

## Preconditions

- `repoA` is an initialized git repository.
- A local hook and a global hook are registered before `list` runs.

## Steps

1. Set `req.WorkDir` to `req.RepoA`.
2. Add `local-hook` (`echo local`) from `repoA`.
3. Add `global-hook` (`echo global`) from `req.FakeHome`.
4. Run `git-hooks pre-push list` with no scope flags.

```go
import "fmt"

func Setup(t *testing.T, req *Request) error {
	req.WorkDir = req.RepoA
	req.ListArgs = nil
	req.CaseName = "pre-push list inside repo default shows both"
	env := isolatedEnv(req)
	if err := runCLI(env, req.RepoA, req.ToolPath, "pre-push", "add", "local-hook", "echo", "local"); err != nil {
		return fmt.Errorf("seed local hook: %w", err)
	}
	if err := runCLI(env, req.FakeHome, req.ToolPath, "pre-push", "add", "global-hook", "echo", "global"); err != nil {
		return fmt.Errorf("seed global hook: %w", err)
	}
	return nil
}
```