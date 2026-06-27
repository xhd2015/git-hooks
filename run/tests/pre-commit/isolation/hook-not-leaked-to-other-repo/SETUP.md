# Scenario

Hook added in repo A must not run on commit in repo B.

## Preconditions

- Repo A has `git-hooks install` and `git-hooks pre-commit add leak-test`.
- Repo B has `git-hooks install` only.
- The `leak-test` hook writes `LEAKED_FROM_REPO_A` to a marker file under fake home.

## Steps

1. Run the shared `Run` flow from the root `DOCTEST.md`.
2. The marker file path is `req.MarkerPath` (`$HOME/leak.out`).

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "hook not leaked to other repo"
	return nil
}
```