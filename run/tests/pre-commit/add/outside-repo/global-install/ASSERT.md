## Expected

- Exit code `0`.
- Stdout mentions global install (e.g. `Installed global git hooks at`).
- Stdout confirms hook added (`Added pre-commit hook: outside-hook`).

## Side Effects

- Global dispatchers exist under `$XDG_CONFIG_HOME/.git-hooks/hooks/` with managed markers.
- `git config --global core.hooksPath` points at the global hooks directory.
- Managed hook exists at `$XDG_CONFIG_HOME/.git-hooks/pre-commit.d/outside-hook`.

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
	if !strings.Contains(resp.AddOutput, "Installed global git hooks at") {
		t.Fatalf("stdout missing global install message:\n%s", resp.AddOutput)
	}
	if !strings.Contains(resp.AddOutput, "Added pre-commit hook: outside-hook") {
		t.Fatalf("stdout missing added hook message:\n%s", resp.AddOutput)
	}

	hooksDir := globalHooksDir(req)
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	prePushPath := filepath.Join(hooksDir, "pre-push")
	preCommitContent, err := readFileOrEmpty(preCommitPath)
	if err != nil {
		t.Fatal(err)
	}
	prePushContent, err := readFileOrEmpty(prePushPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preCommitContent, globalPreCommitMarker) {
		t.Fatalf("global pre-commit dispatcher missing marker at %s:\n%s", preCommitPath, preCommitContent)
	}
	if !strings.Contains(prePushContent, globalPrePushMarker) {
		t.Fatalf("global pre-push dispatcher missing marker at %s:\n%s", prePushPath, prePushContent)
	}

	wantHooksPath := hooksDir
	if resolved, err := filepath.EvalSymlinks(wantHooksPath); err == nil {
		wantHooksPath = resolved
	}
	gotHooksPath := resp.CoreHooksPath
	if resolved, err := filepath.EvalSymlinks(gotHooksPath); err == nil {
		gotHooksPath = resolved
	}
	if gotHooksPath != wantHooksPath {
		t.Fatalf("core.hooksPath = %q, want %q", resp.CoreHooksPath, wantHooksPath)
	}

	managedPath := filepath.Join(globalManagedDir(req, "pre-commit"), "outside-hook")
	if _, err := os.Stat(managedPath); err != nil {
		t.Fatalf("global managed hook missing at %s: %v", managedPath, err)
	}
}
```