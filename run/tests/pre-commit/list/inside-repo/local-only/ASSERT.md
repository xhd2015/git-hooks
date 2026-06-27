## Expected

- Exit code `0`.
- Output contains only the local hook line without `[local]` prefix.
- Global hooks are not listed.

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
	want := "local-hook\techo local"
	got := strings.TrimSpace(resp.ListOutput)
	if got != want {
		t.Fatalf("list output = %q, want %q", got, want)
	}
	if strings.Contains(got, "[local]") || strings.Contains(got, "[global]") || strings.Contains(got, "global-hook") {
		t.Fatalf("list output must not include scope prefixes or global hooks:\n%s", resp.ListOutput)
	}
}
```