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
	return listHooks("pre-commit", args)
}

func listPrePushHooks(config Config, args []string) error {
	return listHooks("pre-push", args)
}

type hookListScope struct {
	label string
	dir   string
}

func listHooks(phase string, args []string) error {
	var showOrigin, localFlag, globalFlag bool
	_, err := flags.
		Bool("--show-origin", &showOrigin).
		Bool("--local", &localFlag).
		Bool("--global", &globalFlag).
		Help("-h,--help", "git-hooks "+phase+" list").
		Parse(args)
	if err != nil {
		return err
	}

	if localFlag && globalFlag {
		localFlag = false
		globalFlag = false
	}

	inRepo := isInsideGitRepo()
	showLocal, showGlobal := resolveListScopes(localFlag, globalFlag, inRepo)
	usePrefixes := showLocal && showGlobal

	var scopes []hookListScope
	if showLocal {
		dir, err := localManagedDir(phase)
		if err != nil {
			return err
		}
		scopes = append(scopes, hookListScope{label: "local", dir: dir})
	}
	if showGlobal {
		dir, err := globalManagedDir(phase)
		if err != nil {
			return err
		}
		scopes = append(scopes, hookListScope{label: "global", dir: dir})
	}

	if showOrigin {
		for i, scope := range scopes {
			fmt.Printf("%s hooks directory (%s): %s\n", phase, scope.label, resolvedPath(scope.dir))
			if i < len(scopes)-1 {
				fmt.Println()
			}
		}
		if len(scopes) > 0 {
			fmt.Println()
		}
	}

	for _, scope := range scopes {
		hooks, err := listManagedHooks(scope.dir)
		if err != nil {
			return err
		}
		for _, hook := range hooks {
			path := filepath.Join(scope.dir, hook)
			command, err := managedHookCommand(path)
			if err != nil {
				return err
			}
			displayName := displayHookName(hook)
			prefix := ""
			if usePrefixes {
				prefix = fmt.Sprintf("[%s] ", scope.label)
			}
			if command == "" {
				fmt.Printf("%s%s\n", prefix, displayName)
				continue
			}
			fmt.Printf("%s%s\t%s\n", prefix, displayName, command)
		}
	}
	return nil
}

func resolveListScopes(localFlag, globalFlag, inRepo bool) (showLocal, showGlobal bool) {
	if localFlag {
		return inRepo, false
	}
	if globalFlag {
		return false, true
	}
	if inRepo {
		return true, true
	}
	return false, true
}

func isInsideGitRepo() bool {
	_, err := gitCommonDirAbs()
	return err == nil
}

func localManagedDir(phase string) (string, error) {
	commonDir, err := gitCommonDirAbs()
	if err != nil {
		return "", err
	}
	return filepath.Join(commonDir, "git-hooks", phase+".d"), nil
}

func globalManagedDir(phase string) (string, error) {
	root, err := userConfigRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, phase+".d"), nil
}

func resolvedPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
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