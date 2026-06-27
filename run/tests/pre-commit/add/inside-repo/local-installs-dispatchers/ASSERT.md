## Expected

- Exit code `0`.
- Stdout mentions local install locations (e.g. `Installed local hooks at`).
- Stdout confirms hook added (`Added pre-commit hook: test-hook`).

## Side Effects

- `.git/hooks/pre-commit` and `.git/hooks/pre-push` exist and contain git-hooks managed markers.
- Managed hook script exists at `<git-common-dir>/git-hooks/pre-commit.d/test-hook`.
- No global `core.hooksPath` is configured.
- Global dispatcher directory under fake home is not created by this command.

## Errors

- None.

## Exit Code

- `0`

```go
import (
	"os"
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.AddExit != 0 {
		t.Fatalf("add exit code = %d, want 0\n%s", resp.AddExit, resp.AddOutput)
	}
	if !strings.Contains(resp.AddOutput, "Installed local hooks at") {
		t.Fatalf("stdout missing local install message:\n%s", resp.AddOutput)
	}
	if !strings.Contains(resp.AddOutput, "Added pre-commit hook: test-hook") {
		t.Fatalf("stdout missing added hook message:\n%s", resp.AddOutput)
	}

	preCommitPath := localDispatcherPath(req, "pre-commit")
	prePushPath := localDispatcherPath(req, "pre-push")
	preCommitContent, err := readFileOrEmpty(preCommitPath)
	if err != nil {
		t.Fatal(err)
	}
	prePushContent, err := readFileOrEmpty(prePushPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preCommitContent, localPreCommitMarker) {
		t.Fatalf("local pre-commit dispatcher missing marker at %s:\n%s", preCommitPath, preCommitContent)
	}
	if !strings.Contains(prePushContent, localPrePushMarker) {
		t.Fatalf("local pre-push dispatcher missing marker at %s:\n%s", prePushPath, prePushContent)
	}

	managedPath := filepath.Join(managedDir(req, "local"), "test-hook")
	if _, err := os.Stat(managedPath); err != nil {
		t.Fatalf("managed hook missing at %s: %v", managedPath, err)
	}

	if resp.CoreHooksPath != "" {
		t.Fatalf("core.hooksPath should be unset, got %q", resp.CoreHooksPath)
	}
	if _, err := os.Stat(globalHooksDir(req)); !os.IsNotExist(err) {
		t.Fatalf("global dispatcher dir should not exist: %v", err)
	}
}
```