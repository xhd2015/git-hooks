## Expected

- Exit code `0`.
- Output matches default `list` (both scopes, local first, prefixed lines).

## Side Effects

- None.

## Exit Code

- `0`

```go
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
}
```