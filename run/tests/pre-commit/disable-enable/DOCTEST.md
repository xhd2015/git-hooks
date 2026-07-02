# Disable and Enable Managed Hooks

Tests for `git-hooks <phase> disable|enable`, covering local and global scope,
single-hook and all-hook targets, pre-commit and pre-push phases, and invalid
target combinations.

## Version

0.0.2

# DSN (Domain Specific Notion)

- **CLI binary** — built from the repository root and executed as `git-hooks`
  with an isolated `PATH`.
- **Fake home** — isolated `HOME` and `XDG_CONFIG_HOME`; global managed hooks
  live under `$XDG_CONFIG_HOME/.git-hooks/<phase>.d/`.
- **Repository** — temporary git repo with local managed hooks under
  `<git-common-dir>/git-hooks/<phase>.d/`.
- **Managed hook file** — executable script created by `git-hooks <phase> add`;
  the filename is the stored identity and the displayed name removes an ordering
  prefix such as `01-`.
- **Disable command** — selects one hook by display name or every hook with
  `--all`, removes executable bits only, and leaves the file listed.
- **Enable command** — selects one hook by display name or every hook with
  `--all`, restores executable bits for readable classes, and leaves the file
  content and name unchanged.
- **Hook runner** — `pre-commit run` and `pre-push run` execute only hook files
  that have executable bits.
- **Validation** — disable/enable requires exactly one target: either a display
  name or `--all`.

## How to Run

```sh
doctest vet ./run/tests/pre-commit/disable-enable
doctest test ./run/tests/pre-commit/disable-enable
doctest test ./run/tests/pre-commit/disable-enable/...
```

## Test Tree

```
disable-enable/
├── local/
│   ├── single-disable
│   ├── single-enable
│   └── all-hooks
├── pre-push-phase/
│   └── single-hook
├── global/
│   └── same-display-name
└── validation/
    ├── missing-target
    └── name-and-all
```

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type Request struct {
	CommandDir string
	BinDir     string
	ToolPath   string
	FakeHome   string
	RepoA      string
	MarkerPath string

	CaseName string
	Phase    string
	WorkDir  string

	CommandArgs []string
	WantUsage   string
}

type Response struct {
	Output   string
	ExitCode int

	SecondOutput   string
	SecondExitCode int

	ListOutput string
	ListExit   int
	RunOutput  string
	RunExit    int

	BeforeModes map[string]os.FileMode
	AfterModes  map[string]os.FileMode
	FinalModes  map[string]os.FileMode
	BeforeData  map[string]string
	AfterData   map[string]string
	BeforeNames []string
	AfterNames  []string
	FinalNames  []string

	MarkerExists bool
	MarkerData   string
}

func isolatedEnv(req *Request) []string {
	return append(os.Environ(),
		"HOME="+req.FakeHome,
		"XDG_CONFIG_HOME="+filepath.Join(req.FakeHome, ".config"),
		"PATH="+req.BinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
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

func runCLI(env []string, dir string, tool string, args ...string) error {
	output, exitCode, err := runCLICapture(env, dir, tool, args...)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("%s %s exited %d: %s", filepath.Base(tool), strings.Join(args, " "), exitCode, strings.TrimSpace(output))
	}
	return nil
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
	output, exitCode, err := runGitCapture(env, dir, args...)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("git %s exited %d: %s", strings.Join(args, " "), exitCode, strings.TrimSpace(output))
	}
	return nil
}

func writeFile(root string, rel string, content string) error {
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func setupRepoIdentity(env []string, repo string) error {
	if err := runGit(env, repo, "config", "user.email", "a@test"); err != nil {
		return err
	}
	return runGit(env, repo, "config", "user.name", "A")
}

func managedDir(req *Request, scope string, phase string) string {
	if scope == "local" {
		return filepath.Join(req.RepoA, ".git", "git-hooks", phase+".d")
	}
	return filepath.Join(req.FakeHome, ".config", ".git-hooks", phase+".d")
}

func hookPath(req *Request, scope string, phase string, file string) string {
	return filepath.Join(managedDir(req, scope, phase), file)
}

func snapshotDir(dir string) (map[string]os.FileMode, map[string]string, []string, error) {
	modes := map[string]os.FileMode{}
	data := map[string]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			return nil, nil, nil, err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		names = append(names, name)
		modes[name] = info.Mode().Perm()
		data[name] = string(content)
	}
	sort.Strings(names)
	return modes, data, names, nil
}

func markerState(path string) (bool, string, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return true, string(data), nil
	}
	if os.IsNotExist(err) {
		return false, "", nil
	}
	return false, "", err
}

func seedHook(req *Request, scope string, phase string, file string, markerText string) error {
	env := isolatedEnv(req)
	name := file
	cmd := fmt.Sprintf("echo %s >> %q", markerText, req.MarkerPath)
	args := []string{phase, "add"}
	if scope == "global" {
		args = append(args, "--global")
	}
	args = append(args, name, "sh", "-c", cmd)
	return runCLI(env, req.RepoA, req.ToolPath, args...)
}

func chmodHook(req *Request, scope string, phase string, file string, mode os.FileMode) error {
	return os.Chmod(hookPath(req, scope, phase, file), mode)
}

func runCommit(req *Request) (string, int, error) {
	env := isolatedEnv(req)
	if err := setupRepoIdentity(env, req.RepoA); err != nil {
		return "", 0, err
	}
	if err := writeFile(req.RepoA, "file.txt", fmt.Sprintf("%s\n", req.CaseName)); err != nil {
		return "", 0, err
	}
	if err := runGit(env, req.RepoA, "add", "file.txt"); err != nil {
		return "", 0, err
	}
	return runGitCapture(env, req.RepoA, "commit", "-m", req.CaseName)
}

func runPhase(req *Request) (string, int, error) {
	env := isolatedEnv(req)
	return runCLICapture(env, req.RepoA, req.ToolPath, req.Phase, "run")
}

func runSingleCommand(req *Request) (*Response, error) {
	env := isolatedEnv(req)
	dir := managedDir(req, "local", req.Phase)
	beforeModes, beforeData, beforeNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	output, exitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, req.CommandArgs...)
	if err != nil {
		return nil, err
	}
	afterModes, afterData, afterNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	listOutput, listExit, err := runCLICapture(env, req.WorkDir, req.ToolPath, req.Phase, "list", "--local")
	if err != nil {
		return nil, err
	}
	runOutput, runExit, err := runCommit(req)
	if err != nil {
		return nil, err
	}
	markerExists, markerData, err := markerState(req.MarkerPath)
	if err != nil {
		return nil, err
	}
	return &Response{
		Output:       output,
		ExitCode:     exitCode,
		ListOutput:   listOutput,
		ListExit:     listExit,
		RunOutput:    runOutput,
		RunExit:      runExit,
		BeforeModes:  beforeModes,
		AfterModes:   afterModes,
		BeforeData:   beforeData,
		AfterData:    afterData,
		BeforeNames:  beforeNames,
		AfterNames:   afterNames,
		MarkerExists: markerExists,
		MarkerData:   markerData,
	}, nil
}

func runAllHooks(req *Request) (*Response, error) {
	env := isolatedEnv(req)
	dir := managedDir(req, "local", req.Phase)
	beforeModes, beforeData, beforeNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	output, exitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, req.Phase, "disable", "--all")
	if err != nil {
		return nil, err
	}
	afterModes, afterData, afterNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	secondOutput, secondExitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, req.Phase, "enable", "--all")
	if err != nil {
		return nil, err
	}
	finalModes, _, finalNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	return &Response{
		Output:         output,
		ExitCode:       exitCode,
		SecondOutput:   secondOutput,
		SecondExitCode: secondExitCode,
		BeforeModes:    beforeModes,
		AfterModes:     afterModes,
		FinalModes:     finalModes,
		BeforeData:     beforeData,
		AfterData:      afterData,
		BeforeNames:    beforeNames,
		AfterNames:     afterNames,
		FinalNames:     finalNames,
	}, nil
}

func runPrePushDisableEnable(req *Request) (*Response, error) {
	env := isolatedEnv(req)
	dir := managedDir(req, "local", "pre-push")
	beforeModes, beforeData, beforeNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	output, exitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, "pre-push", "disable", "push-check")
	if err != nil {
		return nil, err
	}
	afterModes, afterData, afterNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	firstRunOutput, firstRunExit, err := runPhase(req)
	if err != nil {
		return nil, err
	}
	secondOutput, secondExitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, "pre-push", "enable", "push-check")
	if err != nil {
		return nil, err
	}
	finalModes, _, finalNames, err := snapshotDir(dir)
	if err != nil {
		return nil, err
	}
	secondRunOutput, secondRunExit, err := runPhase(req)
	if err != nil {
		return nil, err
	}
	markerExists, markerData, err := markerState(req.MarkerPath)
	if err != nil {
		return nil, err
	}
	return &Response{
		Output:         output,
		ExitCode:       exitCode,
		SecondOutput:   secondOutput,
		SecondExitCode: secondExitCode,
		RunOutput:      firstRunOutput + secondRunOutput,
		RunExit:        firstRunExit + secondRunExit,
		BeforeModes:    beforeModes,
		AfterModes:     afterModes,
		FinalModes:     finalModes,
		BeforeData:     beforeData,
		AfterData:      afterData,
		BeforeNames:    beforeNames,
		AfterNames:     afterNames,
		FinalNames:     finalNames,
		MarkerExists:   markerExists,
		MarkerData:     markerData,
	}, nil
}

func runGlobalDisable(req *Request) (*Response, error) {
	env := isolatedEnv(req)
	globalDir := managedDir(req, "global", "pre-commit")
	localDir := managedDir(req, "local", "pre-commit")
	beforeModes, beforeData, beforeNames, err := snapshotDir(globalDir)
	if err != nil {
		return nil, err
	}
	localBeforeModes, _, _, err := snapshotDir(localDir)
	if err != nil {
		return nil, err
	}
	output, exitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, "pre-commit", "disable", "--global", "check")
	if err != nil {
		return nil, err
	}
	afterModes, afterData, afterNames, err := snapshotDir(globalDir)
	if err != nil {
		return nil, err
	}
	localAfterModes, _, _, err := snapshotDir(localDir)
	if err != nil {
		return nil, err
	}
	for name, mode := range localAfterModes {
		afterModes["local:"+name] = mode
	}
	for name, mode := range localBeforeModes {
		beforeModes["local:"+name] = mode
	}
	return &Response{
		Output:      output,
		ExitCode:    exitCode,
		BeforeModes: beforeModes,
		AfterModes:  afterModes,
		BeforeData:  beforeData,
		AfterData:   afterData,
		BeforeNames: beforeNames,
		AfterNames:  afterNames,
	}, nil
}

func runValidation(req *Request) (*Response, error) {
	env := isolatedEnv(req)
	output, exitCode, err := runCLICapture(env, req.WorkDir, req.ToolPath, req.CommandArgs...)
	if err != nil {
		return nil, err
	}
	return &Response{Output: output, ExitCode: exitCode}, nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.CaseName {
	case "local all-hooks":
		return runAllHooks(req)
	case "pre-push disable-enable":
		return runPrePushDisableEnable(req)
	case "global same display":
		return runGlobalDisable(req)
	case "validation missing target", "validation name and all":
		return runValidation(req)
	default:
		return runSingleCommand(req)
	}
}
```
