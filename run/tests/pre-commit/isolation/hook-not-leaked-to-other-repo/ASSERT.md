# Expected: repo B commit does not run repo A's managed hook

## Expected

- `git commit` in repo B succeeds.
- The marker file at `$HOME/leak.out` does not exist.
- Commit output does not mention `pre-commit: leak-test`.

## Side Effects

- Repo A's managed hook must remain registered only for repo A.

## Exit Code

- Repo B commit exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.CommitExit != 0 {
		t.Fatalf("repo B commit exit code = %d, want 0\n%s", resp.CommitExit, resp.CommitOutput)
	}
	if resp.MarkerExists {
		t.Fatalf("repo B commit leaked repo A hook: marker %s exists with data %q\ncommit output:\n%s", req.MarkerPath, resp.MarkerData, resp.CommitOutput)
	}
	if strings.Contains(resp.CommitOutput, "pre-commit: leak-test") {
		t.Fatalf("repo B commit output mentions repo A hook:\n%s", resp.CommitOutput)
	}
}
```