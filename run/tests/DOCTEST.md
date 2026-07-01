# git-hooks CLI — Doc-Style Test Tree

Tests for the `git-hooks` CLI: local/global install, auto-install on `add`, managed
pre-commit and pre-push hooks, repository isolation, and `list` scope flags
(`--local` / `--global`).

## Version

0.0.2

# DSN (Domain Specific Notion)

- **CLI binary** — built from the repository root (`main.go` + `run/` package).
- **Fake home** — isolated `HOME` temp directory; never the real user home.
- **Repository** — temporary git repo with local hook storage under `<git-common-dir>/git-hooks/`.
- **Global config** — user-scoped hook storage under `$XDG_CONFIG_HOME/.git-hooks/`.
- **Auto-install on add** — `git-hooks <phase> add` always runs full dispatcher install
  first (both pre-commit and pre-push dispatchers), then registers the managed hook.
  Inside a repo the default is local install; `--global` selects global install and
  global storage; outside a repo global install is always used.
- **Managed hooks** — hook scripts registered with `git-hooks <phase> add`; destination
  depends on cwd and `--global` (repo default → local, repo `--global` or non-repo → global).
- **List command** — `git-hooks <phase> list` prints managed hooks; scope flags choose
  local only, global only, or both (default), with `[local]` / `[global]` prefixes when
  both scopes are shown.
- **Isolation** — hooks added while working in one repo must not run on commit in another repo.

## How to Run

```sh
doctest vet ./
doctest test -v ./
```

## Test Tree

```
run/tests/
├── pre-commit/
│   ├── add/
│   │   ├── inside-repo/
│   │   │   ├── local-installs-dispatchers
│   │   │   ├── global-flag
│   │   │   ├── idempotent-second-add
│   │   │   └── preserves-existing-hook
│   │   └── outside-repo/
│   │       └── global-install
│   ├── isolation/
│   │   └── hook-not-leaked-to-other-repo
│   └── list/
│       ├── inside-repo/
│       │   ├── default-shows-both
│       │   ├── local-only
│       │   ├── global-only
│       │   ├── both-flags-same-as-default
│       │   └── show-origin-both-scopes
│       └── outside-repo/
│           ├── default-global-only
│           ├── local-empty
│           └── global-only
│   ├── global-run/
│   │   ├── global-hook-runs-on-commit
│   │   └── global-hook-runs-on-other-repo
└── pre-push/
    ├── add/
    │   ├── inside-repo/
    │   │   └── local-installs-dispatchers
    │   └── outside-repo/
    │       └── global-install
    └── list/
        └── inside-repo/
            └── default-shows-both
```

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type Request struct {
	CommandDir string
	BinDir     string
	ToolPath   string
	FakeHome   string
	RepoA      string
	RepoB      string
	MarkerPath string
	CaseName   string

	TestKind string
	Phase    string
	WorkDir  string

	// List tests (TestKind == "list")
	ListArgs []string

	// Add tests (TestKind == "add")
	AddArgs      []string
	HookName     string
	HookCmd      []string
	RunAddTwice  bool
	SecondHookName string
	SecondHookCmd  []string
}

type Response struct {
	CommitOutput string
	CommitExit   int
	MarkerExists bool
	MarkerData   string

	ListOutput string
	ListExit   int

	AddOutput       string
	AddExit         int
	SecondAddOutput string
	SecondAddExit   int
	CoreHooksPath   string
}

func isolatedEnv(req *Request) []string {
	return append(os.Environ(),
		"HOME="+req.FakeHome,
		"XDG_CONFIG_HOME="+filepath.Join(req.FakeHome, ".config"),
		"PATH="+req.BinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func runCLI(env []string, dir string, tool string, args ...string) error {
	cmd := exec.Command(tool, args...)
	cmd.Dir = dir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", filepath.Base(tool), strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runCLICapture(env []string, dir string, tool string, args ...string) (string, int, error) {
	cmd := exec.Command(tool, args...)
	cmd.Dir = dir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", 0, err
		}
	}
	return string(output), exitCode, nil
}

func runGitCapture(env []string, dir string, args ...string) (string, int, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", 0, err
		}
	}
	return string(output), exitCode, nil
}

func runGit(env []string, dir string, args ...string) error {
	_, _, err := runGitCapture(env, dir, args...)
	return err
}

func writeFile(root string, rel string, content string) error {
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func managedDir(req *Request, scope string) string {
	phaseDir := req.Phase + ".d"
	var dir string
	if scope == "local" {
		dir = filepath.Join(req.RepoA, ".git", "git-hooks", phaseDir)
	} else {
		dir = filepath.Join(req.FakeHome, ".config", ".git-hooks", phaseDir)
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return dir
}

const (
	localPreCommitMarker  = "# git-hooks managed pre-commit start"
	localPrePushMarker    = "# git-hooks managed pre-push start"
	globalPreCommitMarker = "# git-hooks managed global pre-commit start"
	globalPrePushMarker   = "# git-hooks managed global pre-push start"
)

func localDispatcherPath(req *Request, hook string) string {
	return filepath.Join(req.RepoA, ".git", "hooks", hook)
}

func globalHooksDir(req *Request) string {
	return filepath.Join(req.FakeHome, ".config", ".git-hooks", "hooks")
}

func globalManagedDir(req *Request, phase string) string {
	return filepath.Join(req.FakeHome, ".config", ".git-hooks", phase+".d")
}

func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func countSubstring(s, sub string) int {
	if sub == "" {
		return 0
	}
	return strings.Count(s, sub)
}

func listLines(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}
	return strings.Split(output, "\n")
}

func runList(t *testing.T, req *Request) (*Response, error) {
	env := isolatedEnv(req)
	args := append([]string{req.Phase, "list"}, req.ListArgs...)
	output, exitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, args...)
	if err != nil {
		return nil, err
	}
	return &Response{
		ListOutput: output,
		ListExit:   exitCode,
	}, nil
}

func runIsolation(t *testing.T, req *Request) (*Response, error) {
	env := isolatedEnv(req)

	if err := runGit(env, req.RepoA, "config", "user.email", "a@test"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoA, "config", "user.name", "A"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoB, "config", "user.email", "b@test"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoB, "config", "user.name", "B"); err != nil {
		return nil, err
	}

	if err := writeFile(req.RepoA, "file.txt", "repo-a\n"); err != nil {
		return nil, err
	}
	if err := writeFile(req.RepoB, "file.txt", "repo-b\n"); err != nil {
	}
	if err := runGit(env, req.RepoA, "add", "file.txt"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoB, "add", "file.txt"); err != nil {
		return nil, err
	}

	if err := runCLI(env, req.RepoA, req.ToolPath, "install"); err != nil {
		return nil, fmt.Errorf("repo A install: %w", err)
	}
	addCmd := fmt.Sprintf("echo LEAKED_FROM_REPO_A >> %q", req.MarkerPath)
	if err := runCLI(env, req.RepoA, req.ToolPath, "pre-commit", "add", "leak-test", "sh", "-c", addCmd); err != nil {
		return nil, fmt.Errorf("repo A pre-commit add: %w", err)
	}

	if err := runCLI(env, req.RepoB, req.ToolPath, "install"); err != nil {
		return nil, fmt.Errorf("repo B install: %w", err)
	}

	output, exitCode, err := runGitCapture(env, req.RepoB, "commit", "-m", "test")
	if err != nil {
		return nil, err
	}

	markerData, readErr := os.ReadFile(req.MarkerPath)
	exists := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}

	return &Response{
		CommitOutput: string(output),
		CommitExit:   exitCode,
		MarkerExists: exists,
		MarkerData:   string(markerData),
	}, nil
}

func runAdd(t *testing.T, req *Request) (*Response, error) {
	env := isolatedEnv(req)
	args := append([]string{req.Phase, "add"}, req.AddArgs...)
	args = append(args, req.HookName)
	args = append(args, req.HookCmd...)
	output, exitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, args...)
	if err != nil {
		return nil, err
	}
	resp := &Response{
		AddOutput: output,
		AddExit:   exitCode,
	}
	if req.RunAddTwice {
		args2 := append([]string{req.Phase, "add"}, req.AddArgs...)
		args2 = append(args2, req.SecondHookName)
		args2 = append(args2, req.SecondHookCmd...)
		output2, exit2, err := runCLICapture(env, req.WorkDir, req.ToolPath, args2...)
		if err != nil {
			return nil, err
		}
		resp.SecondAddOutput = output2
		resp.SecondAddExit = exit2
	}
	hooksPath, _, err := runGitCapture(env, req.FakeHome, "config", "--global", "--get", "core.hooksPath")
	if err != nil {
		return nil, err
	}
	resp.CoreHooksPath = strings.TrimSpace(hooksPath)
	return resp, nil
}

func runGlobalHookOtherRepo(t *testing.T, req *Request) (*Response, error) {
	env := isolatedEnv(req)

	if err := runGit(env, req.RepoB, "config", "user.email", "b@test"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoB, "config", "user.name", "B"); err != nil {
		return nil, err
	}

	addCmd := fmt.Sprintf("echo GLOBAL_HOOK_RAN >> %q", req.MarkerPath)
	hookArgs := []string{"pre-commit", "add", "--global", "global-test", "sh", "-c", addCmd}
	if err := runCLI(env, req.RepoA, req.ToolPath, hookArgs...); err != nil {
		return nil, fmt.Errorf("global hook add: %w", err)
	}

	if err := writeFile(req.RepoB, "file.txt", "repo-b\n"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoB, "add", "file.txt"); err != nil {
		return nil, err
	}

	output, exitCode, err := runGitCapture(env, req.RepoB, "commit", "-m", "test")
	if err != nil && exitCode == 0 {
		return nil, err
	}

	markerData, readErr := os.ReadFile(req.MarkerPath)
	markerExists := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}

	return &Response{
		CommitOutput: string(output),
		CommitExit:   exitCode,
		MarkerExists: markerExists,
		MarkerData:   string(markerData),
	}, nil
}

func runGlobalHook(t *testing.T, req *Request) (*Response, error) {
	env := isolatedEnv(req)

	if err := runGit(env, req.RepoA, "config", "user.email", "a@test"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoA, "config", "user.name", "A"); err != nil {
		return nil, err
	}

	addCmd := fmt.Sprintf("echo GLOBAL_HOOK_RAN >> %q", req.MarkerPath)
	hookArgs := []string{"pre-commit", "add", "--global", "global-test", "sh", "-c", addCmd}
	if err := runCLI(env, req.RepoA, req.ToolPath, hookArgs...); err != nil {
		return nil, fmt.Errorf("global hook add: %w", err)
	}

	if err := writeFile(req.RepoA, "file.txt", "repo-a\n"); err != nil {
		return nil, err
	}
	if err := runGit(env, req.RepoA, "add", "file.txt"); err != nil {
		return nil, err
	}

	output, exitCode, err := runGitCapture(env, req.RepoA, "commit", "-m", "test")
	if err != nil && exitCode == 0 {
		return nil, err
	}

	markerData, readErr := os.ReadFile(req.MarkerPath)
	markerExists := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}

	return &Response{
		CommitOutput: string(output),
		CommitExit:   exitCode,
		MarkerExists: markerExists,
		MarkerData:   string(markerData),
	}, nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.TestKind {
	case "list":
		return runList(t, req)
	case "add":
		return runAdd(t, req)
	case "global-run-other":
		return runGlobalHookOtherRepo(t, req)
	case "global-run":
		return runGlobalHook(t, req)
	default:
		return runIsolation(t, req)
	}
}
```