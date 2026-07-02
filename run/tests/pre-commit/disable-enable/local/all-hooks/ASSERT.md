## Expected

- Disable stdout is exactly `Disabled 2 pre-commit hooks`.
- Enable stdout is exactly `Enabled 2 pre-commit hooks`.

## Side Effects

- Both hooks become non-executable after disable.
- Both hooks become executable again after enable.
- Filenames and content are unchanged.

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
		t.Fatalf("disable --all exit code = %d\n%s", resp.ExitCode, resp.Output)
	}
	if resp.SecondExitCode != 0 {
		t.Fatalf("enable --all exit code = %d\n%s", resp.SecondExitCode, resp.SecondOutput)
	}
	if resp.Output != "Disabled 2 pre-commit hooks\n" {
		t.Fatalf("disable --all output = %q", resp.Output)
	}
	if resp.SecondOutput != "Enabled 2 pre-commit hooks\n" {
		t.Fatalf("enable --all output = %q", resp.SecondOutput)
	}
	if strings.Join(resp.BeforeNames, ",") != "01-alpha,02-beta" {
		t.Fatalf("before filenames = %v", resp.BeforeNames)
	}
	if strings.Join(resp.AfterNames, ",") != "01-alpha,02-beta" || strings.Join(resp.FinalNames, ",") != "01-alpha,02-beta" {
		t.Fatalf("filenames changed: after=%v final=%v", resp.AfterNames, resp.FinalNames)
	}
	for _, name := range resp.BeforeNames {
		if resp.BeforeModes[name]&0o111 == 0 {
			t.Fatalf("%s should start executable: %v", name, resp.BeforeModes[name])
		}
		if resp.AfterModes[name]&0o111 != 0 {
			t.Fatalf("%s still executable after disable: %v", name, resp.AfterModes[name])
		}
		if resp.FinalModes[name]&0o111 == 0 {
			t.Fatalf("%s not executable after enable: %v", name, resp.FinalModes[name])
		}
		if resp.BeforeData[name] != resp.AfterData[name] {
			t.Fatalf("%s content changed", name)
		}
	}
}
```
