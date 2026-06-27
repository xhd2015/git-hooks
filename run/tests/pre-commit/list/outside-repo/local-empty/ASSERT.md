## Expected

- Exit code `0`.
- Output is empty (no hook lines, no directory headers).

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
	if resp.ListOutput != "" {
		t.Fatalf("list output must be empty outside repo with --local, got:\n%q", resp.ListOutput)
	}
}
```