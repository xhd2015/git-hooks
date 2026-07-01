## Expected

- The commit succeeds (exit code 0).
- The global hook executed during commit.
- The marker file at `req.MarkerPath` exists and contains `GLOBAL_HOOK_RAN`.

## Exit Code

- Commit exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.MarkerExists {
		t.Fatalf("global hook did not run: marker file %s was not created\ncommit output:\n%s", req.MarkerPath, resp.CommitOutput)
	}
	if !strings.Contains(resp.MarkerData, "GLOBAL_HOOK_RAN") {
		t.Fatalf("global hook ran but marker content is wrong: got %q, want content containing GLOBAL_HOOK_RAN\ncommit output:\n%s", resp.MarkerData, resp.CommitOutput)
	}
	if resp.CommitExit != 0 {
		t.Fatalf("commit exit code = %d, want 0\ncommit output:\n%s", resp.CommitExit, resp.CommitOutput)
	}
}

```
