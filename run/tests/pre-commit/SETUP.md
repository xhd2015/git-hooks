# Scenario

We are testing managed `pre-commit` hook behavior (not `pre-push`, not global install).

## Preconditions

- Both repositories are initialized git repos under isolated temp directories.
- The CLI binary is built and available on `PATH` as `git-hooks`.

## Steps

1. Keep the root setup from `run/tests/SETUP.md`.
2. Leaf cases under this node add repo-specific conditions and assertions.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "pre-commit"
	return nil
}
```