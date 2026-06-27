# Scenario

**Feature**: `pre-commit add` appends the managed block when a custom local pre-commit hook already exists.

```
# custom pre-commit hook without marker -> add appends managed block + installs pre-push dispatcher
existing .git/hooks/pre-commit -> git-hooks pre-commit add -> append block, fresh pre-push install
```

## Preconditions

- Repo has a custom executable `.git/hooks/pre-commit` without a git-hooks marker.

## Steps

1. Write custom pre-commit hook content to `.git/hooks/pre-commit`.
2. Run `git-hooks pre-commit add append-hook echo appended` from `repoA`.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	hookPath := localDispatcherPath(req, "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
		return err
	}
	custom := "#!/bin/sh\necho custom-pre-commit\n"
	if err := os.WriteFile(hookPath, []byte(custom), 0o755); err != nil {
		return err
	}
	req.HookName = "append-hook"
	req.HookCmd = []string{"echo", "appended"}
	req.CaseName = "pre-commit add preserves existing hook"
	return nil
}
```