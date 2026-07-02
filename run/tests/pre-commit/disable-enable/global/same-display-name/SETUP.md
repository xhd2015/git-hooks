# Scenario

**Feature**: `pre-commit disable --global <name>` disables only the global hook when local and global display names collide.

```
# local and global scopes each contain display name "check"
local pre-commit.d/01-check (0755)
global pre-commit.d/01-check (0755)

# --global selects global file only
git-hooks pre-commit disable --global check -> global non-executable, local still executable
```

## Preconditions

- Local and global `01-check` hooks are both executable.

## Steps

1. Seed local `01-check`.
2. Seed global `01-check`.
3. Run `git-hooks pre-commit disable --global check`.

## Context

- The command output uses the display name `check`.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "global same display"
	if err := seedHook(req, "local", "pre-commit", "01-check", "LOCAL_SHOULD_STAY_EXECUTABLE"); err != nil {
		return err
	}
	return seedHook(req, "global", "pre-commit", "01-check", "GLOBAL_DISABLED")
}
```
