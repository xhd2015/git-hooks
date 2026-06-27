## Expected

- Both commands exit `0`.
- First stdout includes local install message.
- Second stdout includes already-installed messages for dispatchers (e.g. `already contains git-hooks block`).
- Second stdout confirms second hook added.

## Side Effects

- Both managed hooks exist in local `pre-commit.d`.
- Local dispatcher files contain exactly one managed marker each (no duplication).

## Errors

- None.

## Exit Code

- `0` for both runs.

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
		t.Fatalf("first add exit code = %d, want 0\n%s", resp.AddExit, resp.AddOutput)
	}
	if resp.SecondAddExit != 0 {
		t.Fatalf("second add exit code = %d, want 0\n%s", resp.SecondAddExit, resp.SecondAddOutput)
	}
	if !strings.Contains(resp.AddOutput, "Installed local hooks at") {
		t.Fatalf("first add missing install message:\n%s", resp.AddOutput)
	}
	if !strings.Contains(resp.SecondAddOutput, "already contains git-hooks block") {
		t.Fatalf("second add missing already-installed message:\n%s", resp.SecondAddOutput)
	}
	if !strings.Contains(resp.SecondAddOutput, "Added pre-commit hook: hook-two") {
		t.Fatalf("second add missing added hook message:\n%s", resp.SecondAddOutput)
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
	if countSubstring(preCommitContent, localPreCommitMarker) != 1 {
		t.Fatalf("pre-commit marker count = %d, want 1 in:\n%s", countSubstring(preCommitContent, localPreCommitMarker), preCommitContent)
	}
	if countSubstring(prePushContent, localPrePushMarker) != 1 {
		t.Fatalf("pre-push marker count = %d, want 1 in:\n%s", countSubstring(prePushContent, localPrePushMarker), prePushContent)
	}

	for _, name := range []string{"hook-one", "hook-two"} {
		path := filepath.Join(managedDir(req, "local"), name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("managed hook %s missing at %s: %v", name, path, err)
		}
	}
}
```