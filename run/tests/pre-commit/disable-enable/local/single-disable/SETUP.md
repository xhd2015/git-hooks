# Scenario

**Feature**: `pre-commit disable <name>` disables one local hook by display name and keeps it listed.

```
# stored hook has an ordering prefix, command target uses display name
pre-commit.d/01-check (0755) <- git-hooks pre-commit disable check

# disabled hook remains listed but runner skips it
git-hooks pre-commit list --local -> check
git commit -> no marker write
```

## Preconditions

- Local hook `01-check` exists and is executable.
- The hook command would write `DISABLED_SHOULD_NOT_RUN` to the marker if executed.

## Steps

1. Seed `01-check` in local pre-commit storage.
2. Run `git-hooks pre-commit disable check`.

## Context

- The command must preserve the stored filename `01-check` and hook script content.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "local single-disable"
	req.CommandArgs = []string{"pre-commit", "disable", "check"}
	return seedHook(req, "local", "pre-commit", "01-check", "DISABLED_SHOULD_NOT_RUN")
}
```
