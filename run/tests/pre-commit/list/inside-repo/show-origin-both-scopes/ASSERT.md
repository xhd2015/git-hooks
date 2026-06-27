## Expected

- Exit code `0`.
- Output begins with local then global directory headers:
  - `pre-commit hooks directory (local): <path>`
  - `pre-commit hooks directory (global): <path>`
- Hook lines follow with `[local]` / `[global]` prefixes.

## Side Effects

- None.

## Exit Code

- `0`

```go
import (
	"fmt"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ListExit != 0 {
		t.Fatalf("list exit code = %d, want 0\n%s", resp.ListExit, resp.ListOutput)
	}
	localDir := managedDir(req, "local")
	globalDir := managedDir(req, "global")
	wantHeader := strings.Join([]string{
		fmt.Sprintf("pre-commit hooks directory (local): %s", localDir),
		"",
		fmt.Sprintf("pre-commit hooks directory (global): %s", globalDir),
		"",
		"[local] local-hook\techo local",
		"[global] global-hook\techo global",
	}, "\n") + "\n"
	if resp.ListOutput != wantHeader {
		t.Fatalf("list output mismatch\n\ngot:\n%q\n\nwant:\n%q", resp.ListOutput, wantHeader)
	}
}
```