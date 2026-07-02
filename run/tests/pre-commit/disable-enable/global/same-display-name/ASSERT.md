## Expected

- Exit code `0`.
- Stdout is exactly `Disabled pre-commit hook: check`.

## Side Effects

- Global `01-check` loses executable bits.
- Local `01-check` remains executable.
- Global filename and content are unchanged.

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
		t.Fatalf("global disable exit code = %d\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "Disabled pre-commit hook: check\n" {
		t.Fatalf("global disable output = %q", resp.Output)
	}
	if resp.AfterModes["01-check"]&0o111 != 0 {
		t.Fatalf("global hook still executable: %v", resp.AfterModes["01-check"])
	}
	if resp.AfterModes["local:01-check"]&0o111 == 0 {
		t.Fatalf("local hook should remain executable: before %v after %v", resp.BeforeModes["local:01-check"], resp.AfterModes["local:01-check"])
	}
	if strings.Join(resp.AfterNames, ",") != "01-check" {
		t.Fatalf("global filename changed: %v", resp.AfterNames)
	}
	if resp.BeforeData["01-check"] != resp.AfterData["01-check"] {
		t.Fatalf("global hook content changed")
	}
}
```
