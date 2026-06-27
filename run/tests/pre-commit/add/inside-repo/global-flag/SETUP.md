# Scenario

**Feature**: `pre-commit add --global` inside a repo auto-installs global dispatchers and stores the hook globally.

```
git-hooks pre-commit add --global global-hook (repo cwd) -> global install -> global pre-commit.d/global-hook
```

## Preconditions

- Fresh repo; no prior install.

## Steps

1. Run `git-hooks pre-commit add --global global-hook echo global` from `repoA`.

```go
func Setup(t *testing.T, req *Request) error {
	req.AddArgs = []string{"--global"}
	req.HookName = "global-hook"
	req.HookCmd = []string{"echo", "global"}
	req.CaseName = "pre-commit add --global"
	return nil
}
```