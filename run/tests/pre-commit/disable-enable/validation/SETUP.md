# Scenario

**Feature**: disable/enable target validation rejects missing or conflicting target selectors.

```
# command must choose exactly one selector
git-hooks pre-commit disable -> usage error
git-hooks pre-commit disable --all check -> usage error
```

## Preconditions

- The test runs inside a git repository.
- No hook needs to exist for argument validation.

## Steps

1. Keep `req.WorkDir` as `repoA`.
2. Descendant scenarios set invalid command arguments.

## Context

- Validation should fail before selecting hook files.

```go
func Setup(t *testing.T, req *Request) error {
	req.Phase = "pre-commit"
	req.WorkDir = req.RepoA
	return nil
}
```
