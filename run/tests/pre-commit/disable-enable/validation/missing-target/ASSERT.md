## Expected

- The command fails.
- Output contains command-specific usage.

## Side Effects

- No hook files are required or modified.

## Errors

- `usage: git-hooks pre-commit disable [<name>|--all]`

## Exit Code

- Non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("missing target unexpectedly succeeded:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, req.WantUsage) {
		t.Fatalf("missing target output missing usage %q:\n%s", req.WantUsage, resp.Output)
	}
}
```
