# Scenario

**Feature**: `pre-commit disable` without a name or `--all` is rejected with usage.

```
# no target selector
git-hooks pre-commit disable -> usage error
```

## Preconditions

- The command has no positional target and no `--all` flag.

## Steps

1. Run `git-hooks pre-commit disable`.

## Context

- The command should not silently select all hooks.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "validation missing target"
	req.CommandArgs = []string{"pre-commit", "disable"}
	req.WantUsage = "usage: git-hooks pre-commit disable [<name>|--all]"
	return nil
}
```
