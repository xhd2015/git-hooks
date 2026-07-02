## Expected

- Exit code `0`.
- Stdout is exactly `Enabled pre-commit hook: check`.
- `01-check` remains present with unchanged content.

## Side Effects

- Executable bits are added for every readable class.
- The following commit executes the enabled hook.

## Errors

- None.

## Exit Code

- `0`

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("enable exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "Enabled pre-commit hook: check\n" {
		t.Fatalf("enable output = %q", resp.Output)
	}
	before := resp.BeforeModes["01-check"]
	after := resp.AfterModes["01-check"]
	if before != 0o644 {
		t.Fatalf("test setup disabled mode = %v, want 0644", before)
	}
	if after != 0o755 {
		t.Fatalf("enabled mode = %v, want 0755", after)
	}
	if resp.BeforeData["01-check"] != resp.AfterData["01-check"] {
		t.Fatalf("hook content changed")
	}
	if strings.Join(resp.AfterNames, ",") != "01-check" {
		t.Fatalf("hook filename changed: %v", resp.AfterNames)
	}
	if resp.RunExit != 0 {
		t.Fatalf("commit exit code = %d\n%s", resp.RunExit, resp.RunOutput)
	}
	if !resp.MarkerExists || !strings.Contains(resp.MarkerData, "ENABLED_RAN") {
		t.Fatalf("enabled hook did not run, exists=%v data=%q", resp.MarkerExists, resp.MarkerData)
	}
}
```
