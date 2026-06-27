# Scenario

**Feature**: git-hooks CLI manages pre-commit hooks and installs dispatchers into git repositories.

```
isolated HOME -> two temp git repos -> install + pre-commit add -> git commit
```

## Preconditions

- The `git-hooks` command module lives at the repository root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- Every test runs with `HOME` set to a dedicated temporary directory (never the real user home).
- `XDG_CONFIG_HOME` is set to `$HOME/.config` so managed hooks stay inside the fake home.
- The built CLI is placed on `PATH` as `git-hooks` because installed hook scripts invoke that command name.

## Steps

1. Create `fakeHome`, `binDir`, `repoA`, and `repoB` under `t.TempDir()`.
2. Build `git-hooks` into `binDir/git-hooks` before overriding `HOME`.
3. Set `HOME` and `XDG_CONFIG_HOME` to isolated paths.
4. Prepend `binDir` to `PATH`.
5. Initialize both git repositories.
6. Let leaf `Setup` customize `Request` fields if needed.
7. Run the shared `Run` implementation from `DOCTEST.md`.

## Context

- `git-hooks install` (without `--global`) writes dispatcher scripts into `.git/hooks/`.
- `git-hooks pre-commit add` registers managed hook commands.
- Managed hooks must be scoped to the repository where they were added, not shared across all repositories on the machine.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.CommandDir = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	req.FakeHome = filepath.Join(t.TempDir(), "home")
	req.BinDir = filepath.Join(t.TempDir(), "bin")
	req.RepoA = filepath.Join(t.TempDir(), "repo-a")
	req.RepoB = filepath.Join(t.TempDir(), "repo-b")
	req.MarkerPath = filepath.Join(req.FakeHome, "leak.out")
	req.ToolPath = filepath.Join(req.BinDir, "git-hooks")
	req.CaseName = "default"

	if err := os.MkdirAll(req.FakeHome, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.BinDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.RepoA, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.RepoB, 0o755); err != nil {
		return err
	}

	build := exec.Command("go", "build", "-o", req.ToolPath, ".")
	build.Dir = req.CommandDir
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go build git-hooks: %w: %s", err, strings.TrimSpace(string(output)))
	}

	t.Setenv("HOME", req.FakeHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(req.FakeHome, ".config"))

	if err := runGit(nil, req.RepoA, "init"); err != nil {
		return err
	}
	if err := runGit(nil, req.RepoB, "init"); err != nil {
		return err
	}
	return nil
}
```