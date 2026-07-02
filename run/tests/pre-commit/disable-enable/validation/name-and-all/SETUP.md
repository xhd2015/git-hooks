# Scenario

**Feature**: `pre-commit disable --all <name>` is rejected because target selectors conflict.

```
# both selectors are present
git-hooks pre-commit disable --all check -> usage error
```

## Preconditions

- A hook may exist, but validation should reject arguments before applying changes.

## Steps

1. Seed local hook `01-check`.
2. Run `git-hooks pre-commit disable --all check`.

## Context

- Exactly one target is required: name or `--all`, not both.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "validation name and all"
	req.CommandArgs = []string{"pre-commit", "disable", "--all", "check"}
	req.WantUsage = "usage: git-hooks pre-commit disable [<name>|--all]"
	return seedHook(req, "local", "pre-commit", "01-check", "SHOULD_NOT_CHANGE")
}
```
