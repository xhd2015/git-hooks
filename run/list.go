package run

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
		if command == "" {
			fmt.Println(hook)
			continue
		}
		fmt.Printf("%s\t%s\n", hook, command)
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
	sort.Strings(hooks)
	return hooks, nil
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
