## Expected

- Exit code `0`.
- Output contains only the global hook line without `[global]` prefix.
- Local hooks are not listed.

## Side Effects

- None.

## Exit Code

- `0`

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ListExit != 0 {
		t.Fatalf("list exit code = %d, want 0\n%s", resp.ListExit, resp.ListOutput)
	}
	want := "global-hook\techo global"
	got := strings.TrimSpace(resp.ListOutput)
	if got != want {
		t.Fatalf("list output = %q, want %q", got, want)
	}
	if strings.Contains(got, "[local]") || strings.Contains(got, "[global]") || strings.Contains(got, "local-hook") {
		t.Fatalf("list output must not include scope prefixes or local hooks:\n%s", resp.ListOutput)
	}
}
```