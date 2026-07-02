package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

func addPreCommitHook(config Config, args []string) error {
	var global bool
	var err error
	args, err = flags.Bool("--global", &global).StopOnFirstArg().Parse(args)
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: git-hooks pre-commit add [--global] <name> <cmd>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	useGlobal, managedRoot, err := resolveAddScope(config, global)
	if err != nil {
		return err
	}
	if err := ensureHooksInstalledForAdd(config, useGlobal); err != nil {
		return err
	}
	dir := filepath.Join(managedRoot, "pre-commit.d")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\nset -eu\nexec " + shellJoin(args[1:]) + ` "$@"
`
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	printAddedManagedHook("pre-commit", name, path, args[1:])
	return nil
}

func removePreCommitHook(config Config, args []string) error {
	var global bool
	var err error
	args, err = flags.Bool("--global", &global).StopOnFirstArg().Parse(args)
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-commit remove [--global] <name>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	dir, err := removeManagedDir(config, "pre-commit", global)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pre-commit hook %s does not exist", name)
		}
		return err
	}
	fmt.Printf("Removed pre-commit hook: %s\n", name)
	return nil
}

func disablePreCommitHook(config Config, args []string) error {
	return toggleHookEnabled(config, "pre-commit", false, args)
}

func enablePreCommitHook(config Config, args []string) error {
	return toggleHookEnabled(config, "pre-commit", true, args)
}

func addPrePushHook(config Config, args []string) error {
	var global bool
	var err error
	args, err = flags.Bool("--global", &global).StopOnFirstArg().Parse(args)
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: git-hooks pre-push add [--global] <name> <cmd>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	useGlobal, managedRoot, err := resolveAddScope(config, global)
	if err != nil {
		return err
	}
	if err := ensureHooksInstalledForAdd(config, useGlobal); err != nil {
		return err
	}
	dir := filepath.Join(managedRoot, "pre-push.d")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\nset -eu\nexec " + shellJoin(args[1:]) + ` "$@"
`
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	printAddedManagedHook("pre-push", name, path, args[1:])
	return nil
}

func removePrePushHook(config Config, args []string) error {
	var global bool
	var err error
	args, err = flags.Bool("--global", &global).StopOnFirstArg().Parse(args)
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-push remove [--global] <name>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	dir, err := removeManagedDir(config, "pre-push", global)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pre-push hook %s does not exist", name)
		}
		return err
	}
	fmt.Printf("Removed pre-push hook: %s\n", name)
	return nil
}

func removeManagedDir(config Config, phase string, global bool) (string, error) {
	if global {
		return globalManagedHooksDir(phase)
	}
	switch phase {
	case "pre-commit":
		return managedPreCommitDir(config), nil
	case "pre-push":
		return managedPrePushDir(config), nil
	default:
		return "", fmt.Errorf("unknown phase: %s", phase)
	}
}

func disablePrePushHook(config Config, args []string) error {
	return toggleHookEnabled(config, "pre-push", false, args)
}

func enablePrePushHook(config Config, args []string) error {
	return toggleHookEnabled(config, "pre-push", true, args)
}

func toggleHookEnabled(config Config, phase string, enable bool, args []string) error {
	var global bool
	var all bool
	var err error
	action := "disable"
	actionPast := "Disabled"
	if enable {
		action = "enable"
		actionPast = "Enabled"
	}
	usage := fmt.Sprintf("usage: git-hooks %s %s [<name>|--all]", phase, action)
	args, err = flags.
		Bool("--global", &global).
		Bool("--all", &all).
		Parse(args)
	if err != nil {
		return err
	}
	if all == (len(args) == 1) || len(args) > 1 {
		return fmt.Errorf("%s", usage)
	}

	dir, err := managedHookDirForScope(config, phase, global)
	if err != nil {
		return err
	}
	if all {
		hooks, err := listManagedHooks(dir)
		if err != nil {
			return err
		}
		for _, hook := range hooks {
			if err := setHookExecutable(filepath.Join(dir, hook), enable); err != nil {
				return err
			}
		}
		fmt.Printf("%s %d %s hooks\n", actionPast, len(hooks), phase)
		return nil
	}

	file, err := resolveHookName(dir, args[0])
	if err != nil {
		return err
	}
	if err := setHookExecutable(filepath.Join(dir, file), enable); err != nil {
		return err
	}
	fmt.Printf("%s %s hook: %s\n", actionPast, phase, displayHookName(file))
	return nil
}

func managedHookDirForScope(config Config, phase string, global bool) (string, error) {
	if global {
		return globalManagedHooksDir(phase)
	}
	dir, err := localManagedDir(phase)
	if err != nil {
		return "", fmt.Errorf("not inside a git repo: %w", err)
	}
	return dir, nil
}

func setHookExecutable(path string, enable bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode().Perm()
	if !enable {
		return os.Chmod(path, mode&^0111)
	}
	if mode&0400 != 0 {
		mode |= 0100
	}
	if mode&0040 != 0 {
		mode |= 0010
	}
	if mode&0004 != 0 {
		mode |= 0001
	}
	return os.Chmod(path, mode)
}

func renamePreCommitHook(config Config, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: git-hooks pre-commit rename <old-name> <new-name>")
	}
	return renameHook(managedPreCommitDir(config), args[0], args[1])
}

func renamePrePushHook(config Config, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: git-hooks pre-push rename <old-name> <new-name>")
	}
	return renameHook(managedPrePushDir(config), args[0], args[1])
}

func renameHook(dir, oldDisplay, newDisplay string) error {
	oldFile, err := resolveHookName(dir, oldDisplay)
	if err != nil {
		return err
	}
	if err := validateHookName(newDisplay); err != nil {
		return err
	}

	prefixStr, hasPrefix := hookPrefixStr(oldFile)
	var newFile string
	if hasPrefix {
		newFile = prefixStr + newDisplay
	} else {
		newFile = newDisplay
	}

	newPath := filepath.Join(dir, newFile)
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("hook %s already exists", newDisplay)
	}

	oldPath := filepath.Join(dir, oldFile)
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}
	fmt.Printf("Renamed %s -> %s\n", oldDisplay, displayHookName(newFile))
	return nil
}

func upPreCommitHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-commit up <name>")
	}
	return moveHook(managedPreCommitDir(config), args[0], true)
}

func downPreCommitHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-commit down <name>")
	}
	return moveHook(managedPreCommitDir(config), args[0], false)
}

func topPreCommitHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-commit top <name>")
	}
	return topHook(managedPreCommitDir(config), args[0])
}

func upPrePushHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-push up <name>")
	}
	return moveHook(managedPrePushDir(config), args[0], true)
}

func downPrePushHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-push down <name>")
	}
	return moveHook(managedPrePushDir(config), args[0], false)
}

func topPrePushHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-push top <name>")
	}
	return topHook(managedPrePushDir(config), args[0])
}

func moveHook(dir, displayName string, up bool) error {
	file, err := resolveHookName(dir, displayName)
	if err != nil {
		return err
	}

	hooks, err := listManagedHooks(dir)
	if err != nil {
		return err
	}

	myIdx := -1
	for i, h := range hooks {
		if h == file {
			myIdx = i
			break
		}
	}
	if myIdx < 0 {
		return fmt.Errorf("internal: hook %s not found in sorted list", file)
	}

	otherIdx := myIdx - 1
	if !up {
		otherIdx = myIdx + 1
	}
	if otherIdx < 0 || otherIdx >= len(hooks) {
		if up {
			return fmt.Errorf("already first")
		}
		return fmt.Errorf("already last")
	}

	otherFile := hooks[otherIdx]
	if err := swapHookFiles(dir, file, otherFile); err != nil {
		return err
	}
	fmt.Printf("Swapped %s and %s\n", displayHookName(file), displayHookName(otherFile))
	return nil
}

func topHook(dir, displayName string) error {
	file, err := resolveHookName(dir, displayName)
	if err != nil {
		return err
	}

	hooks, err := listManagedHooks(dir)
	if err != nil {
		return err
	}
	if len(hooks) > 0 && hooks[0] == file {
		fmt.Printf("%s is already first\n", displayName)
		return nil
	}

	topFile := "00-" + displayHookName(file)
	topPath := filepath.Join(dir, topFile)
	if _, err := os.Stat(topPath); err == nil {
		return fmt.Errorf("top position already occupied by %s", displayHookName(topFile))
	}

	oldPath := filepath.Join(dir, file)
	if err := os.Rename(oldPath, topPath); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}
	fmt.Printf("Moved %s to top\n", displayName)
	return nil
}

func swapHookFiles(dir, a, b string) error {
	tmp := "__git_hooks_swap_tmp__"
	aPath := filepath.Join(dir, a)
	bPath := filepath.Join(dir, b)
	tmpPath := filepath.Join(dir, tmp)

	if err := os.Rename(bPath, tmpPath); err != nil {
		return fmt.Errorf("rename %s failed: %w", b, err)
	}
	if err := os.Rename(aPath, bPath); err != nil {
		os.Rename(tmpPath, bPath)
		return fmt.Errorf("rename %s failed: %w", a, err)
	}
	if err := os.Rename(tmpPath, aPath); err != nil {
		return fmt.Errorf("rename %s failed: %w", tmp, err)
	}
	return nil
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

func printAddedManagedHook(phase, name, path string, cmdArgs []string) {
	fmt.Printf("Added %s hook: %s at %s\n", phase, name, path)
	fmt.Printf("  command: %s\n", displayShellCommand(shellJoin(cmdArgs)))
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

func resolveAddScope(config Config, globalFlag bool) (useGlobal bool, managedRoot string, err error) {
	inRepo := isInsideGitRepo()
	if globalFlag || !inRepo {
		root, err := userConfigRootDir()
		if err != nil {
			return false, "", err
		}
		return true, root, nil
	}
	return false, config.ConfigRoot, nil
}

func ensureHooksInstalledForAdd(config Config, useGlobal bool) error {
	if useGlobal {
		return installGlobalHooks(config, false)
	}
	return installLocalHooks(false)
}
