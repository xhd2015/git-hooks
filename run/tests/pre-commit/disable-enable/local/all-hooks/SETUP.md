# Scenario

**Feature**: `pre-commit disable --all` and `enable --all` apply to every local pre-commit hook without renaming files.

```
# all local pre-commit hooks are selected
pre-commit.d/{01-alpha,02-beta} <- git-hooks pre-commit disable --all
pre-commit.d/{01-alpha,02-beta} <- git-hooks pre-commit enable --all
```

## Preconditions

- Local hooks `01-alpha` and `02-beta` exist and are executable.

## Steps

1. Seed two local pre-commit hooks.
2. Run `git-hooks pre-commit disable --all`.
3. Run `git-hooks pre-commit enable --all`.

## Context

- The success output is a count summary, not one line per hook.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "local all-hooks"
	if err := seedHook(req, "local", "pre-commit", "01-alpha", "ALPHA"); err != nil {
		return err
	}
	return seedHook(req, "local", "pre-commit", "02-beta", "BETA")
}
```
