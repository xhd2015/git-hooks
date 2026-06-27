## Expected

- Exit code `0`.
- Output lists only the global hook without scope prefixes.

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
	if strings.Contains(got, "[global]") {
		t.Fatalf("single-scope --global must not use scope prefix:\n%s", resp.ListOutput)
	}
}
```