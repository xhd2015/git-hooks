## Expected

- Exit code `0`.
- Output lists local hooks first, then global hooks.
- Each line is prefixed with `[local]` or `[global]`.
- Lines use tab-separated `hook-name<TAB>command` format after the prefix.

## Side Effects

- No hooks are added or removed by `list`.

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
	want := []string{
		"[local] local-hook\techo local",
		"[global] global-hook\techo global",
	}
	got := listLines(resp.ListOutput)
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(got), len(want), resp.ListOutput)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("line %d = %q, want %q\nfull output:\n%s", i, got[i], w, resp.ListOutput)
		}
	}
	if strings.Contains(resp.ListOutput, "\t\n") {
		t.Fatalf("unexpected blank lines in output:\n%s", resp.ListOutput)
	}
}
```