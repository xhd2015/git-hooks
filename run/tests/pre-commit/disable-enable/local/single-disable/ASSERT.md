## Expected

- Exit code `0`.
- Stdout is exactly `Disabled pre-commit hook: check`.
- `01-check` is still present and listed as `check`.
- File content is unchanged.

## Side Effects

- The executable bits are removed from `01-check`.
- Read/write bits remain unchanged.
- The following commit does not execute the disabled hook.

## Errors

- None.

## Exit Code

- `0`

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("disable exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "Disabled pre-commit hook: check\n" {
		t.Fatalf("disable output = %q", resp.Output)
	}
	before := resp.BeforeModes["01-check"]
	after := resp.AfterModes["01-check"]
	if before&0o111 == 0 {
		t.Fatalf("test setup expected executable hook, mode %v", before)
	}
	if after&0o111 != 0 {
		t.Fatalf("disabled hook still executable: before %v after %v", before, after)
	}
	if before&0o666 != after&0o666 {
		t.Fatalf("disable changed read/write bits: before %v after %v", before, after)
	}
	if strings.Join(resp.BeforeNames, ",") != "01-check" || strings.Join(resp.AfterNames, ",") != "01-check" {
		t.Fatalf("hook filename changed: before %v after %v", resp.BeforeNames, resp.AfterNames)
	}
	if resp.BeforeData["01-check"] != resp.AfterData["01-check"] {
		t.Fatalf("hook content changed")
	}
	if resp.ListExit != 0 {
		t.Fatalf("list exit code = %d\n%s", resp.ListExit, resp.ListOutput)
	}
	if !strings.Contains(resp.ListOutput, "check\t") {
		t.Fatalf("disabled hook should remain listed:\n%s", resp.ListOutput)
	}
	if resp.RunExit != 0 {
		t.Fatalf("commit exit code = %d\n%s", resp.RunExit, resp.RunOutput)
	}
	if resp.MarkerExists {
		t.Fatalf("disabled hook ran unexpectedly: %q", resp.MarkerData)
	}
}
```
