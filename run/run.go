package run

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/less-gen/flags"
)

//go:embed pre-commit-global.sh
var preCommitGlobalScript string

//go:embed pre-commit-local.sh
var preCommitLocalScript string

//go:embed pre-push-global.sh
var prePushGlobalScript string

//go:embed pre-push-local.sh
var prePushLocalScript string

const help = `
Usage: git-hooks <command> [OPTIONS]

Commands:
  install [--global] [--dry-run]   install git hook dispatcher
  pre-commit list [--local] [--global] [--show-origin]  list managed pre-commit hooks
  pre-commit add [--global] <name> <cmd>  add a managed pre-commit hook (auto-installs dispatchers)
  pre-commit remove [--global] <name>  remove a managed pre-commit hook
  pre-commit disable [--global] [<name>|--all]  disable managed pre-commit hook(s)
  pre-commit enable [--global] [<name>|--all]   enable managed pre-commit hook(s)
  pre-commit rename <old> <new>    rename a managed pre-commit hook
  pre-commit up <name>             move hook earlier (swap with previous)
  pre-commit down <name>           move hook later (swap with next)
  pre-commit top <name>            move hook to the first position
  pre-commit run [--amend]         run managed pre-commit hooks
  pre-push list [--local] [--global] [--show-origin]    list managed pre-push hooks
  pre-push add [--global] <name> <cmd>    add a managed pre-push hook (auto-installs dispatchers)
  pre-push remove [--global] <name>  remove a managed pre-push hook
  pre-push disable [--global] [<name>|--all]    disable managed pre-push hook(s)
  pre-push enable [--global] [<name>|--all]     enable managed pre-push hook(s)
  pre-push rename <old> <new>      rename a managed pre-push hook
  pre-push up <name>               move hook earlier (swap with previous)
  pre-push down <name>             move hook later (swap with next)
  pre-push top <name>              move hook to the first position
  pre-push run                     run managed pre-push hooks

Options:
  -h,--help                        show help message
`

const preCommitHelp = `
Usage: git-hooks pre-commit <command> [OPTIONS]

Commands:
  list [--local] [--global] [--show-origin]  list managed pre-commit hooks
  add [--global] <name> <cmd>   add a managed pre-commit hook (auto-installs dispatchers)
  disable [--global] [<name>|--all]  disable managed pre-commit hook(s)
  enable [--global] [<name>|--all]   enable managed pre-commit hook(s)
  remove [--global] <name>      remove a managed pre-commit hook
  rename <old> <new>            rename a managed pre-commit hook
  up <name>                     move hook earlier (swap with previous)
  down <name>                   move hook later (swap with next)
  top <name>                    move hook to the first position
  run [--amend]                 run managed pre-commit hooks
`

const prePushHelp = `
Usage: git-hooks pre-push <command> [OPTIONS]

Commands:
  list [--local] [--global] [--show-origin]  list managed pre-push hooks
  add [--global] <name> <cmd>   add a managed pre-push hook (auto-installs dispatchers)
  disable [--global] [<name>|--all]  disable managed pre-push hook(s)
  enable [--global] [<name>|--all]   enable managed pre-push hook(s)
  remove [--global] <name>      remove a managed pre-push hook
  rename <old> <new>            rename a managed pre-push hook
  up <name>                     move hook earlier (swap with previous)
  down <name>                   move hook later (swap with next)
  top <name>                    move hook to the first position
  run                           run managed pre-push hooks
`

type Config struct {
	ConfigRoot string
}

func Main(args []string) error {
	config := Config{}
	args, err := flags.
		Help("-h,--help", help).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return err
	}
	configRoot, err := configRootDir()
	if err != nil {
		return err
	}
	config.ConfigRoot = configRoot
	return Run(config, args)
}

func Run(config Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command")
	}
	switch args[0] {
	case "install":
		return handleInstall(config, args[1:])
	case "pre-commit":
		return handlePreCommit(config, args[1:])
	case "pre-push":
		return handlePrePush(config, args[1:])
	case "help", "--help", "-h":
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func handleInstall(config Config, args []string) error {
	var global bool
	var dryRun bool
	args, err := flags.
		Bool("--global", &global).
		Bool("--dry-run", &dryRun).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	if global {
		return installGlobalHooks(config, dryRun)
	}
	return installLocalHooks(dryRun)
}

func handlePreCommit(config Config, args []string) error {
	args, err := flags.
		Help("-h,--help", preCommitHelp).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return fmt.Errorf("requires pre-commit command")
	}
	switch args[0] {
	case "list":
		return listPreCommitHooks(config, args[1:])
	case "add":
		return addPreCommitHook(config, args[1:])
	case "disable":
		return disablePreCommitHook(config, args[1:])
	case "enable":
		return enablePreCommitHook(config, args[1:])
	case "remove":
		return removePreCommitHook(config, args[1:])
	case "rename":
		return renamePreCommitHook(config, args[1:])
	case "up":
		return upPreCommitHook(config, args[1:])
	case "down":
		return downPreCommitHook(config, args[1:])
	case "top":
		return topPreCommitHook(config, args[1:])
	case "run":
		return runPreCommitHooks(config, args[1:])
	case "help", "--help", "-h":
		fmt.Print(strings.TrimPrefix(preCommitHelp, "\n"))
		return nil
	default:
		return fmt.Errorf("unknown pre-commit command: %s", args[0])
	}
}

func handlePrePush(config Config, args []string) error {
	args, err := flags.
		Help("-h,--help", prePushHelp).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return fmt.Errorf("requires pre-push command")
	}
	switch args[0] {
	case "list":
		return listPrePushHooks(config, args[1:])
	case "add":
		return addPrePushHook(config, args[1:])
	case "disable":
		return disablePrePushHook(config, args[1:])
	case "enable":
		return enablePrePushHook(config, args[1:])
	case "remove":
		return removePrePushHook(config, args[1:])
	case "rename":
		return renamePrePushHook(config, args[1:])
	case "up":
		return upPrePushHook(config, args[1:])
	case "down":
		return downPrePushHook(config, args[1:])
	case "top":
		return topPrePushHook(config, args[1:])
	case "run":
		return runPrePushHooks(config, args[1:])
	case "help", "--help", "-h":
		fmt.Print(strings.TrimPrefix(prePushHelp, "\n"))
		return nil
	default:
		return fmt.Errorf("unknown pre-push command: %s", args[0])
	}
}

func runPreCommitHooks(config Config, args []string) error {
	var amend bool
	for _, arg := range args {
		if arg == "--amend" {
			amend = true
		}
	}
	shouldRun, err := markPreCommitSession()
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	env := []string{
		"GIT_HOOK_PHASE=pre-commit",
	}
	if amend {
		env = append(env, "GIT_HOOK_AMEND=1")
	}

	localDir := managedPreCommitDir(config)
	globalDir, err := globalManagedHooksDir("pre-commit")
	if err != nil {
		return err
	}
	if localDir != globalDir {
		if err := runManagedHooks(config, localDir, "pre-commit", env); err != nil {
			return err
		}
	}
	return runManagedHooks(config, globalDir, "pre-commit", env)
}

func runPrePushHooks(config Config, args []string) error {
	// NOTE: args might look like
	// ["origin", "ssh://git@github.com/xhd2015/git-hooks"]

	env := []string{
		"GIT_HOOK_PHASE=push",
	}

	localDir := managedPrePushDir(config)
	globalDir, err := globalManagedHooksDir("pre-push")
	if err != nil {
		return err
	}
	if localDir != globalDir {
		if err := runManagedHooks(config, localDir, "pre-push", env); err != nil {
			return err
		}
	}
	return runManagedHooks(config, globalDir, "pre-push", env)
}

func runManagedHooks(config Config, dir string, phase string, extraEnv []string) error {
	hooks, err := listManagedHooks(dir)
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		path := filepath.Join(dir, hook)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Mode()&0111 == 0 {
			continue
		}
		fmt.Printf("%s: %s\n", phase, hook)
		cmd := exec.Command(path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), extraEnv...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s hook %s failed: %w", phase, hook, err)
		}
	}
	return nil
}

func configRootDir() (string, error) {
	if commonDir, err := gitCommonDirAbs(); err == nil {
		return filepath.Join(commonDir, "git-hooks"), nil
	}
	return userConfigRootDir()
}

func userConfigRootDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home := os.Getenv("HOME")
		if home != "" {
			base = filepath.Join(home, ".config")
		}
	}
	if base == "" {
		userConfig, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = userConfig
	}
	return filepath.Join(base, ".git-hooks"), nil
}

func defaultGlobalHooksDir(config Config) string {
	root, err := userConfigRootDir()
	if err != nil {
		return filepath.Join(config.ConfigRoot, "hooks")
	}
	return filepath.Join(root, "hooks")
}

func managedPreCommitDir(config Config) string {
	return filepath.Join(config.ConfigRoot, "pre-commit.d")
}

func managedPrePushDir(config Config) string {
	return filepath.Join(config.ConfigRoot, "pre-push.d")
}

func globalManagedHooksDir(phase string) (string, error) {
	root, err := userConfigRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, phase+".d"), nil
}

func configuredGlobalHooksPath() (string, bool, error) {
	cmd := exec.Command("git", "config", "--global", "--get", "core.hooksPath")
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 && strings.TrimSpace(string(output)) == "" {
			return "", false, nil
		}
		return "", false, fmt.Errorf("git config --global --get core.hooksPath failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	hooksPath := strings.TrimSpace(string(output))
	if hooksPath == "" {
		return "", false, nil
	}
	return expandHomePath(hooksPath), true, nil
}

func expandHomePath(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home := os.Getenv("HOME")
	if home == "" {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/"))
}

func gitCommonDirAbs() (string, error) {
	commonDir, err := gitOutput("", "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}
	commonDir = strings.TrimSpace(commonDir)
	if commonDir == "" {
		return "", fmt.Errorf("git common dir is empty")
	}
	if !filepath.IsAbs(commonDir) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		commonDir = filepath.Join(wd, commonDir)
	}
	return commonDir, nil
}

func localHookPath(hookName string) (string, error) {
	commonDir, err := gitCommonDirAbs()
	if err != nil {
		return "", fmt.Errorf("not inside a git repo: %w", err)
	}
	return filepath.Join(commonDir, "hooks", hookName), nil
}

func markPreCommitSession() (bool, error) {
	sessionID := os.Getenv("GIT_HOOKS_SESSION_ID")
	if sessionID == "" {
		sessionID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
		if err := os.Setenv("GIT_HOOKS_SESSION_ID", sessionID); err != nil {
			return false, err
		}
	}
	sessionID = safeSessionID(sessionID)
	dir := filepath.Join(os.TempDir(), "git-hooks-sessions")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return false, err
	}
	path := filepath.Join(dir, "pre-commit-"+sessionID)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}
	return true, f.Close()
}

func safeSessionID(sessionID string) string {
	var b strings.Builder
	for _, r := range sessionID {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func runCmd(dir string, name string, args ...string) error {
	fmt.Println(name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
