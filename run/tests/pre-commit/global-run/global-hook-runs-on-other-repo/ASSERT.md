## Expected

- The commit in repo B succeeds.
- The global hook executed during the commit in repo B.
- The marker file exists with `GLOBAL_HOOK_RAN`.

## Exit Code

- Commit exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.MarkerExists {
		t.Fatalf("global hook did not run in repo B: marker file %s was not created\ncommit output:\n%s", req.MarkerPath, resp.CommitOutput)
	}
	if !strings.Contains(resp.MarkerData, "GLOBAL_HOOK_RAN") {
		t.Fatalf("global hook ran in repo B but marker content is wrong: got %q, want content containing GLOBAL_HOOK_RAN\ncommit output:\n%s", resp.MarkerData, resp.CommitOutput)
	}
	if resp.CommitExit != 0 {
		t.Fatalf("repo B commit exit code = %d, want 0\ncommit output:\n%s", resp.CommitExit, resp.CommitOutput)
	}
}

```
