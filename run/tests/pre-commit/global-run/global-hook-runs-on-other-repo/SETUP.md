# Scenario

**Feature**: a global pre-commit hook runs on commit even in a different repository.

```
git-hooks pre-commit add --global global-test sh -c 'echo RAN >> $MARKER' (in repo A)
-> git commit in repo B (different repo, same system)
-> global hook executes
-> marker file exists with "GLOBAL_HOOK_RAN"
```

## Preconditions

- The git-hooks manager is installed globally.
- A global hook is registered while working in repo A.

## Steps

1. The shared `runGlobalHookOtherRepo` adds a global hook while cwd is repo A.
2. It then creates a file in repo B, stages it, and commits.
3. The global hook should run because it is global — not scoped to repo A.

```go
func Setup(t *testing.T, req *Request) error {
	req.TestKind = "global-run-other"
	req.CaseName = "global-hook-runs-on-other-repo"
	return nil
}

```
