package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

func addPrePushHook(config Config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: git-hooks pre-push add <name> <cmd>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	dir := managedPrePushDir(config)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\nset -eu\nexec " + shellJoin(args[1:]) + ` "$@"
`
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	fmt.Printf("Added pre-push hook: %s\n", name)
	return nil
}

func removePrePushHook(config Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git-hooks pre-push remove <name>")
	}
	name := args[0]
	if err := validateHookName(name); err != nil {
		return err
	}
	path := filepath.Join(managedPrePushDir(config), name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pre-push hook %s does not exist", name)
		}
		return err
	}
	fmt.Printf("Removed pre-push hook: %s\n", name)
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
