## Expected

- Disable stdout is exactly `Disabled pre-push hook: push-check`.
- Enable stdout is exactly `Enabled pre-push hook: push-check`.

## Side Effects

- Pre-push hook executable bits are removed, then restored.
- Pre-push hook filename and content are unchanged.
- The pre-push hook runs only after enable.
- The pre-commit hook is not modified by pre-push operations.

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
		t.Fatalf("pre-push disable exit code = %d\n%s", resp.ExitCode, resp.Output)
	}
	if resp.SecondExitCode != 0 {
		t.Fatalf("pre-push enable exit code = %d\n%s", resp.SecondExitCode, resp.SecondOutput)
	}
	if resp.Output != "Disabled pre-push hook: push-check\n" {
		t.Fatalf("disable output = %q", resp.Output)
	}
	if resp.SecondOutput != "Enabled pre-push hook: push-check\n" {
		t.Fatalf("enable output = %q", resp.SecondOutput)
	}
	if resp.AfterModes["01-push-check"]&0o111 != 0 {
		t.Fatalf("pre-push hook still executable after disable: %v", resp.AfterModes["01-push-check"])
	}
	if resp.FinalModes["01-push-check"]&0o111 == 0 {
		t.Fatalf("pre-push hook not executable after enable: %v", resp.FinalModes["01-push-check"])
	}
	if strings.Join(resp.AfterNames, ",") != "01-push-check" || strings.Join(resp.FinalNames, ",") != "01-push-check" {
		t.Fatalf("pre-push filename changed: after=%v final=%v", resp.AfterNames, resp.FinalNames)
	}
	if resp.BeforeData["01-push-check"] != resp.AfterData["01-push-check"] {
		t.Fatalf("pre-push hook content changed")
	}
	if !resp.MarkerExists || !strings.Contains(resp.MarkerData, "PUSH_ENABLED_RAN") {
		t.Fatalf("enabled pre-push hook did not run, exists=%v data=%q", resp.MarkerExists, resp.MarkerData)
	}
	if strings.Contains(resp.MarkerData, "COMMIT_SHOULD_NOT_RUN") {
		t.Fatalf("pre-push operation ran pre-commit hook: %q", resp.MarkerData)
	}
}
```
