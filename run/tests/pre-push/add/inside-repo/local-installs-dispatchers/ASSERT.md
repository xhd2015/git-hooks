## Expected

- Exit code `0`.
- Stdout mentions local install locations (e.g. `Installed local hooks at`).
- Stdout confirms hook added (`Added pre-push hook: push-hook`).

## Side Effects

- `.git/hooks/pre-commit` and `.git/hooks/pre-push` exist and contain git-hooks managed markers.
- Managed hook script exists at `<git-common-dir>/git-hooks/pre-push.d/push-hook`.

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
	if !strings.Contains(resp.AddOutput, "Added pre-push hook: push-hook") {
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

	managedPath := filepath.Join(managedDir(req, "local"), "push-hook")
	if _, err := os.Stat(managedPath); err != nil {
		t.Fatalf("managed hook missing at %s: %v", managedPath, err)
	}
}
```