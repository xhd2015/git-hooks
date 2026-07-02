# Scenario

**Feature**: `pre-commit enable <name>` restores executable bits for readable classes and lets the hook run.

```
# hook starts disabled by mode, not by deletion or rename
pre-commit.d/01-check (0644) <- git-hooks pre-commit enable check

# enabled hook is executable again
git commit -> marker write
```

## Preconditions

- Local hook `01-check` exists with mode `0644`.
- The hook command writes `ENABLED_RAN` to the marker when executed.

## Steps

1. Seed `01-check` in local pre-commit storage.
2. Remove executable bits manually to model a previously disabled hook.
3. Run `git-hooks pre-commit enable check`.

## Context

- Enabling a `0644` file should produce executable bits for user, group, and other because all three readable classes are set.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "local single-enable"
	req.CommandArgs = []string{"pre-commit", "enable", "check"}
	if err := seedHook(req, "local", "pre-commit", "01-check", "ENABLED_RAN"); err != nil {
		return err
	}
	return chmodHook(req, "local", "pre-commit", "01-check", 0o644)
}
```
