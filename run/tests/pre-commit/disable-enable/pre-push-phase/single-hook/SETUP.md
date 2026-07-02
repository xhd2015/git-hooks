# Scenario

**Feature**: `pre-push disable <name>` and `enable <name>` toggle a pre-push hook without touching pre-commit hooks.

```
# seed a pre-push hook and a pre-commit hook with different storage
pre-push.d/01-push-check <- git-hooks pre-push add
pre-commit.d/01-commit-check <- git-hooks pre-commit add

# command targets pre-push only
git-hooks pre-push disable push-check -> pre-push.d/01-push-check non-executable
git-hooks pre-push enable push-check -> pre-push.d/01-push-check executable
```

## Preconditions

- `01-push-check` exists in pre-push storage.
- `01-commit-check` exists in pre-commit storage as a guard against phase leakage.

## Steps

1. Seed both hooks.
2. Run pre-push disable and pre-push enable for `push-check`.
3. Run the pre-push hook runner after each command.

## Context

- Only the enabled pre-push run should write the marker.

```go
func Setup(t *testing.T, req *Request) error {
	req.CaseName = "pre-push disable-enable"
	if err := seedHook(req, "local", "pre-push", "01-push-check", "PUSH_ENABLED_RAN"); err != nil {
		return err
	}
	return seedHook(req, "local", "pre-commit", "01-commit-check", "COMMIT_SHOULD_NOT_RUN")
}
```
