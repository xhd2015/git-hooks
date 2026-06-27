# Scenario

**Feature**: second `pre-commit add` reuses existing dispatchers without duplicating managed blocks.

```
# first add installs dispatchers; second add reports already-installed
git-hooks pre-commit add hook-one -> install
git-hooks pre-commit add hook-two -> already-installed messages, no duplicate markers
```

## Preconditions

- Fresh repo with no prior install.

## Steps

1. Run `git-hooks pre-commit add hook-one echo one` from `repoA`.
2. Run `git-hooks pre-commit add hook-two echo two` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.HookName = "hook-one"
	req.HookCmd = []string{"echo", "one"}
	req.RunAddTwice = true
	req.SecondHookName = "hook-two"
	req.SecondHookCmd = []string{"echo", "two"}
	req.CaseName = "pre-commit add idempotent"
	return nil
}
```