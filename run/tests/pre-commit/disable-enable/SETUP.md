# Scenario

**Feature**: `git-hooks <phase> disable|enable` toggles managed hook executable bits without changing hook identity or content.

```
# build isolated CLI, initialize repo, and seed managed hook files through add
git-hooks <phase> add <stored-name> <cmd> -> <scope>/<phase>.d/<stored-name>

# command under test changes executable bits only
git-hooks <phase> disable|enable (<name>|--all) -> chmod managed hook file

# runner observes executable-bit state
git-hooks <phase> run -> executable hooks only
```

## Preconditions

- The module root is four directories above this nested doctest root.
- `HOME` and `XDG_CONFIG_HOME` must point to a fake home for every test.
- The CLI binary must be available on `PATH` as `git-hooks` because dispatcher scripts call it by name.

## Steps

1. Create fake home, temp bin, repo, and marker paths.
2. Build the CLI from the repository root into the temp bin.
3. Initialize the repo.
4. Configure default request fields for `pre-commit` local commands.
5. Descendant scenarios seed hooks and choose command arguments.

## Context

- All hooks are seeded by the existing `add` command.
- Stored filenames may include ordering prefixes, while command targets use display names.
- Assertions compare file modes, file names, and file content before and after the command under test.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	req.CommandDir = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", "..", "..", ".."))
	req.FakeHome = filepath.Join(t.TempDir(), "home")
	req.BinDir = filepath.Join(t.TempDir(), "bin")
	req.RepoA = filepath.Join(t.TempDir(), "repo-a")
	req.MarkerPath = filepath.Join(req.FakeHome, "hook-marker.out")
	req.ToolPath = filepath.Join(req.BinDir, "git-hooks")
	req.Phase = "pre-commit"
	req.WorkDir = req.RepoA
	req.CaseName = "local single-disable"

	for _, dir := range []string{req.FakeHome, req.BinDir, req.RepoA} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
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
	return nil
}
```
