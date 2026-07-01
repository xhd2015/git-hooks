# Scenario

**Feature**: global pre-commit hooks run on commit in any repository

```
# --- current state (bug) ---
global hook registered via --global flag -> git commit -> hook does NOT run

# --- expected ---
global hook registered via --global flag -> git commit -> hook runs (marker file created)
```

## Preconditions

- The `git-hooks` command module lives at the repository root.
- `HOME` is set to an isolated temporary directory.
- `XDG_CONFIG_HOME` is set to `$HOME/.config`.

## Steps

1. Initialize the repository (handled by root SETUP).
2. The leaf case sets `req.CaseName`.
3. Run the shared `runGlobalHook` flow from the root `DOCTEST.md`.

```go
func Setup(t *testing.T, req *Request) error {
	req.TestKind = "global-run"
	return nil
}

```
