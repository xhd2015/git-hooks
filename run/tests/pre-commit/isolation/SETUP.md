# Scenario

Managed pre-commit hooks added in one repository must not run when committing in another repository.

## Preconditions

- Two separate git repositories (`repoA` and `repoB`) share the same isolated `HOME`.
- Both repositories have local hook dispatchers installed.

## Steps

1. Add a managed hook while cwd is `repoA`.
2. Commit staged changes in `repoB`.
3. Assert repo B's commit does not execute repo A's managed hook.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "pre-commit isolation"
	return nil
}
```