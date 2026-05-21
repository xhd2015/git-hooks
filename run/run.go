package run

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xhd2015/less-gen/flags"
)

//go:embed pre-commit-global.sh
var preCommitGlobalScript string

//go:embed pre-commit-local.sh
var preCommitLocalScript string

const help = `
Usage: git-hooks <command> [OPTIONS]

Commands:
  install [--global] [--dry-run] install git hook dispatcher
  pre-commit list               list managed pre-commit hooks
  pre-commit add <name> <cmd>   add a managed pre-commit hook
  pre-commit remove <name>      remove a managed pre-commit hook
  pre-commit run                run managed pre-commit hooks

Options:
  -h,--help                     show help message
`

const preCommitHelp = `
Usage: git-hooks pre-commit <command> [OPTIONS]

Commands:
  list                          list managed pre-commit hooks
  add <name> <cmd>              add a managed pre-commit hook
  remove <name>                 remove a managed pre-commit hook
  run                           run managed pre-commit hooks
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
		if len(args) != 1 {
			return fmt.Errorf("usage: git-hooks pre-commit list")
		}
		return listPreCommitHooks(config)
	case "add":
		return addPreCommitHook(config, args[1:])
	case "remove":
		return removePreCommitHook(config, args[1:])
	case "run":
		return runPreCommitHooks(config, args[1:])
	case "help", "--help", "-h":
		fmt.Print(strings.TrimPrefix(preCommitHelp, "\n"))
		return nil
	default:
		return fmt.Errorf("unknown pre-commit command: %s", args[0])
	}
}

func installGlobalHooks(config Config, dryRun bool) error {
	configuredHooksDir, hasConfiguredHooksDir, err := configuredGlobalHooksPath()
	if err != nil {
		return err
	}
	if hasConfiguredHooksDir {
		return installGlobalHooksAtConfiguredPath(configuredHooksDir, dryRun)
	}

	hooksDir := defaultGlobalHooksDir(config)
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	content := preCommitGlobalScript

	if dryRun {
		fmt.Println("Dry run: would install global git hooks")
		fmt.Printf("Config root: %s\n", config.ConfigRoot)
		fmt.Printf("Would create directory: %s\n", hooksDir)
		fmt.Printf("Would write executable file: %s\n", preCommitPath)
		fmt.Printf("Would run: git config --global core.hooksPath %s\n", hooksDir)
		fmt.Println("Pre-commit hook script:")
		fmt.Print(content)
		return nil
	}

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(preCommitPath, []byte(content), 0755); err != nil {
		return err
	}
	if err := runCmd("", "git", "config", "--global", "core.hooksPath", hooksDir); err != nil {
		return err
	}
	fmt.Printf("Installed global git hooks at %s\n", hooksDir)
	return nil
}

func installGlobalHooksAtConfiguredPath(hooksDir string, dryRun bool) error {
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	block := globalPreCommitHookBlock()
	if dryRun {
		fmt.Println("Dry run: would install global git hooks")
		fmt.Printf("Using configured core.hooksPath: %s\n", hooksDir)
		fmt.Printf("Would ensure directory: %s\n", hooksDir)
		action, err := inspectPreCommitHookInstall(preCommitPath, globalManagedBlockStart)
		if err != nil {
			return err
		}
		switch action {
		case preCommitHookAlreadyInstalled:
			fmt.Printf("Pre-commit hook already contains git-hooks block: %s\n", preCommitPath)
		case preCommitHookAppend:
			fmt.Printf("Would append executable file: %s\n", preCommitPath)
			fmt.Println("Managed block:")
			fmt.Print(block)
		case preCommitHookWrite:
			fmt.Printf("Would write executable file: %s\n", preCommitPath)
			fmt.Println("Pre-commit hook script:")
			fmt.Print(preCommitGlobalScript)
		}
		return nil
	}

	if err := installPreCommitHookFile(preCommitPath, preCommitGlobalScript, block, globalManagedBlockStart, "Global"); err != nil {
		return err
	}
	fmt.Printf("Installed global git hooks at %s\n", hooksDir)
	return nil
}

func installLocalHooks(dryRun bool) error {
	preCommitPath, err := localPreCommitHookPath()
	if err != nil {
		return err
	}
	block := localPreCommitHookBlock()
	if dryRun {
		fmt.Println("Dry run: would install local git pre-commit hook")
		action, err := inspectPreCommitHookInstall(preCommitPath, localManagedBlockStart)
		if err != nil {
			return err
		}
		switch action {
		case preCommitHookAlreadyInstalled:
			fmt.Printf("Pre-commit hook already contains git-hooks block: %s\n", preCommitPath)
		case preCommitHookAppend:
			fmt.Printf("Would append executable file: %s\n", preCommitPath)
			fmt.Println("Managed block:")
			fmt.Print(block)
		case preCommitHookWrite:
			fmt.Printf("Would write executable file: %s\n", preCommitPath)
			fmt.Println("Pre-commit hook script:")
			fmt.Print(preCommitLocalScript)
		}
		return nil
	}

	if err := installPreCommitHookFile(preCommitPath, preCommitLocalScript, block, localManagedBlockStart, "Local"); err != nil {
		return err
	}
	fmt.Printf("Installed local pre-commit hook at %s\n", preCommitPath)
	return nil
}

const globalManagedBlockStart = "# git-hooks managed global pre-commit start"
const localManagedBlockStart = "# git-hooks managed pre-commit start"

func globalPreCommitHookBlock() string {
	return strings.TrimPrefix(preCommitGlobalScript, "#!/bin/sh\n")
}

func localPreCommitHookBlock() string {
	return strings.TrimPrefix(preCommitLocalScript, "#!/bin/sh\n")
}

func installPreCommitHookFile(path string, fullScript string, block string, marker string, label string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var content []byte
	if os.IsNotExist(err) {
		content = []byte(fullScript)
	} else {
		if strings.Contains(string(existing), marker) {
			fmt.Printf("%s pre-commit hook already contains git-hooks block: %s\n", label, path)
			return os.Chmod(path, 0755)
		}
		content = appendPreCommitHookBlock(existing, block)
	}

	if err := os.WriteFile(path, content, 0755); err != nil {
		return err
	}
	return nil
}

func appendPreCommitHookBlock(existing []byte, block string) []byte {
	content := append([]byte{}, existing...)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	content = append(content, '\n')
	content = append(content, []byte(block)...)
	return content
}

type preCommitHookInstallAction int

const (
	preCommitHookWrite preCommitHookInstallAction = iota
	preCommitHookAppend
	preCommitHookAlreadyInstalled
)

func inspectPreCommitHookInstall(path string, marker string) (preCommitHookInstallAction, error) {
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return preCommitHookWrite, nil
		}
		return 0, err
	}
	if strings.Contains(string(existing), marker) {
		return preCommitHookAlreadyInstalled, nil
	}
	return preCommitHookAppend, nil
}

func listPreCommitHooks(config Config) error {
	hooks, err := listManagedHooks(config)
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		path := filepath.Join(managedPreCommitDir(config), hook)
		command, err := managedHookCommand(path)
		if err != nil {
			return err
		}
		if command == "" {
			fmt.Println(hook)
			continue
		}
		fmt.Printf("%s\t%s\n", hook, command)
	}
	return nil
}

func addPreCommitHook(config Config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: git-hooks pre-commit add <name> <cmd>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	dir := managedPreCommitDir(config)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\nset -eu\nexec " + shellJoin(args[1:]) + ` "$@"
`
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	fmt.Printf("Added pre-commit hook: %s\n", name)
	return nil
}

func managedHookCommand(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "exec ") {
			continue
		}
		command := strings.TrimSpace(strings.TrimPrefix(line, "exec "))
		command = strings.TrimSuffix(command, ` "$@"`)
		command = strings.TrimSpace(command)
		return displayShellCommand(command), nil
	}
	return "", nil
}

func removePreCommitHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-commit remove <name>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	path := filepath.Join(managedPreCommitDir(config), name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pre-commit hook %s does not exist", name)
		}
		return err
	}
	fmt.Printf("Removed pre-commit hook: %s\n", name)
	return nil
}

func runPreCommitHooks(config Config, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	shouldRun, err := markPreCommitSession()
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	hooks, err := listManagedHooks(config)
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		path := filepath.Join(managedPreCommitDir(config), hook)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Mode()&0111 == 0 {
			continue
		}
		fmt.Printf("pre-commit: %s\n", hook)
		cmd := exec.Command(path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pre-commit hook %s failed: %w", hook, err)
		}
	}
	return nil
}

func listManagedHooks(config Config) ([]string, error) {
	dir := managedPreCommitDir(config)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var hooks []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		hooks = append(hooks, entry.Name())
	}
	sort.Strings(hooks)
	return hooks, nil
}

func validateHookName(name string) error {
	if name == "" || name == "." || name == ".." {
		return fmt.Errorf("invalid hook name: %s", name)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("hook name must not contain path separators: %s", name)
	}
	return nil
}

func configRootDir() (string, error) {
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
	return filepath.Join(config.ConfigRoot, "hooks")
}

func managedPreCommitDir(config Config) string {
	return filepath.Join(config.ConfigRoot, "pre-commit.d")
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

func localPreCommitHookPath() (string, error) {
	commonDir, err := gitOutput("", "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("not inside a git repo: %w", err)
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
	return filepath.Join(commonDir, "hooks", "pre-commit"), nil
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

func shellJoin(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func displayShellCommand(command string) string {
	words, ok := splitShellWords(command)
	if !ok || len(words) == 0 {
		return command
	}
	display := make([]string, 0, len(words))
	for _, word := range words {
		display = append(display, shellDisplayQuote(word))
	}
	return strings.Join(display, " ")
}

func splitShellWords(command string) ([]string, bool) {
	var words []string
	for i := 0; i < len(command); {
		for i < len(command) && command[i] == ' ' {
			i++
		}
		if i >= len(command) {
			break
		}
		var b strings.Builder
		for i < len(command) && command[i] != ' ' {
			switch command[i] {
			case '\'':
				i++
				for i < len(command) && command[i] != '\'' {
					b.WriteByte(command[i])
					i++
				}
				if i >= len(command) {
					return nil, false
				}
				i++
			case '\\':
				i++
				if i >= len(command) {
					return nil, false
				}
				b.WriteByte(command[i])
				i++
			default:
				b.WriteByte(command[i])
				i++
			}
		}
		words = append(words, b.String())
	}
	return words, true
}

func shellDisplayQuote(s string) string {
	if isShellSafeWord(s) {
		return s
	}
	return shellQuote(s)
}

func isShellSafeWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			continue
		}
		switch r {
		case '_', '@', '%', '+', '=', ':', ',', '.', '/', '-':
			continue
		default:
			return false
		}
	}
	return true
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
