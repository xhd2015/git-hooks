package run

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

func listPreCommitHooks(config Config, args []string) error {
	return listHooks(config, "pre-commit", managedPreCommitDir(config), args)
}

func listPrePushHooks(config Config, args []string) error {
	return listHooks(config, "pre-push", managedPrePushDir(config), args)
}

func listHooks(config Config, phase string, dir string, args []string) error {
	var showOrigin bool
	_, err := flags.
		Bool("--show-origin", &showOrigin).
		Help("-h,--help", "git-hooks "+phase+" list").
		Parse(args)
	if err != nil {
		return err
	}

	if showOrigin {
		fmt.Printf("%s hooks directory: %s\n\n", phase, dir)
	}

	hooks, err := listManagedHooks(dir)
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		path := filepath.Join(dir, hook)
		command, err := managedHookCommand(path)
		if err != nil {
			return err
		}
		displayName := displayHookName(hook)
		if command == "" {
			fmt.Println(displayName)
			continue
		}
		fmt.Printf("%s\t%s\n", displayName, command)
	}
	return nil
}

func listManagedHooks(dir string) ([]string, error) {
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
	sort.Slice(hooks, func(i, j int) bool {
		ni, hasI := extractHookPrefix(hooks[i])
		nj, hasJ := extractHookPrefix(hooks[j])
		if hasI && hasJ {
			if ni != nj {
				return ni < nj
			}
			return hooks[i] < hooks[j]
		}
		if hasI {
			return true
		}
		if hasJ {
			return false
		}
		return hooks[i] < hooks[j]
	})
	return hooks, nil
}

func extractHookPrefix(name string) (int, bool) {
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '-' && i > 0 {
			n, err := strconv.Atoi(name[:i])
			if err == nil {
				return n, true
			}
		}
		break
	}
	return 0, false
}

func displayHookName(filename string) string {
	_, hasPrefix := extractHookPrefix(filename)
	if !hasPrefix {
		return filename
	}
	dash := strings.IndexByte(filename, '-')
	if dash < 0 {
		return filename
	}
	return filename[dash+1:]
}

func hookPrefixStr(name string) (string, bool) {
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '-' && i > 0 {
			return name[:i+1], true
		}
		break
	}
	return "", false
}

func resolveHookName(dir string, displayName string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if displayHookName(entry.Name()) == displayName {
			matches = append(matches, entry.Name())
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("hook not found: %s", displayName)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous hook name %q matches: %s", displayName, strings.Join(matches, ", "))
	}
	return matches[0], nil
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
