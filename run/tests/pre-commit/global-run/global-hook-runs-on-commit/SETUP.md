# Scenario

**Feature**: a global pre-commit hook registered with `--global` runs when committing in the same repo.

```
git-hooks pre-commit add --global global-test sh -c 'echo RAN >> $MARKER'
-> git commit in same repo
-> global hook executes
-> marker file exists with "GLOBAL_HOOK_RAN"
```

## Preconditions

- The git-hooks manager is installed globally (via `--global`).
- A global hook is registered.

## Steps

1. The shared `runGlobalHook` adds a global hook that writes to `req.MarkerPath`.
2. It then creates a file, stages it, and commits.

```go
func Setup(t *testing.T, req *Request) error {
	req.TestKind = "global-run"
	req.CaseName = "global-hook-runs-on-commit"
	return nil
}

```
