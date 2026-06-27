## Expected

- Exit code `0`.
- Stdout mentions local install (pre-push freshly written; pre-commit appended or reported).
- Stdout confirms hook added.

## Side Effects

- `.git/hooks/pre-commit` still contains original custom content and now includes the managed block marker.
- `.git/hooks/pre-push` exists with managed marker.
- Managed hook exists in local `pre-commit.d`.

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
	if !strings.Contains(resp.AddOutput, "Added pre-commit hook: append-hook") {
		t.Fatalf("stdout missing added hook message:\n%s", resp.AddOutput)
	}

	preCommitPath := localDispatcherPath(req, "pre-commit")
	preCommitContent, err := readFileOrEmpty(preCommitPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preCommitContent, "echo custom-pre-commit") {
		t.Fatalf("custom pre-commit content lost at %s:\n%s", preCommitPath, preCommitContent)
	}
	if !strings.Contains(preCommitContent, localPreCommitMarker) {
		t.Fatalf("managed block not appended at %s:\n%s", preCommitPath, preCommitContent)
	}

	prePushPath := localDispatcherPath(req, "pre-push")
	prePushContent, err := readFileOrEmpty(prePushPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prePushContent, localPrePushMarker) {
		t.Fatalf("local pre-push dispatcher missing marker at %s:\n%s", prePushPath, prePushContent)
	}

	managedPath := filepath.Join(managedDir(req, "local"), "append-hook")
	if _, err := os.Stat(managedPath); err != nil {
		t.Fatalf("managed hook missing at %s: %v", managedPath, err)
	}
}
```