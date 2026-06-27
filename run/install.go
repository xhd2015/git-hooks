package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const globalManagedBlockStart = "# git-hooks managed global pre-commit start"
const globalPrePushManagedBlockStart = "# git-hooks managed global pre-push start"
const localManagedBlockStart = "# git-hooks managed pre-commit start"
const localPrePushManagedBlockStart = "# git-hooks managed pre-push start"

func globalPreCommitHookBlock() string {
	return strings.TrimPrefix(preCommitGlobalScript, "#!/bin/sh\n")
}

func globalPrePushHookBlock() string {
	return strings.TrimPrefix(prePushGlobalScript, "#!/bin/sh\n")
}

func localPreCommitHookBlock() string {
	return strings.TrimPrefix(preCommitLocalScript, "#!/bin/sh\n")
}

func localPrePushHookBlock() string {
	return strings.TrimPrefix(prePushLocalScript, "#!/bin/sh\n")
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
	prePushPath := filepath.Join(hooksDir, "pre-push")

	if dryRun {
		fmt.Println("Dry run: would install global git hooks")
		userRoot, _ := userConfigRootDir()
		fmt.Printf("Config root: %s\n", userRoot)
		fmt.Printf("Would create directory: %s\n", hooksDir)
		fmt.Printf("Would write executable file: %s\n", preCommitPath)
		fmt.Printf("Would write executable file: %s\n", prePushPath)
		fmt.Printf("Would run: git config --global core.hooksPath %s\n", hooksDir)
		fmt.Println("Pre-commit hook script:")
		fmt.Print(preCommitGlobalScript)
		fmt.Println("Pre-push hook script:")
		fmt.Print(prePushGlobalScript)
		return nil
	}

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(preCommitPath, []byte(preCommitGlobalScript), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(prePushPath, []byte(prePushGlobalScript), 0755); err != nil {
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
	prePushPath := filepath.Join(hooksDir, "pre-push")
	preCommitBlock := globalPreCommitHookBlock()
	prePushBlock := globalPrePushHookBlock()
	if dryRun {
		fmt.Println("Dry run: would install global git hooks")
		fmt.Printf("Using configured core.hooksPath: %s\n", hooksDir)
		fmt.Printf("Would ensure directory: %s\n", hooksDir)
		printHookDryRun(preCommitPath, preCommitBlock, preCommitGlobalScript, globalManagedBlockStart, "Pre-commit")
		printHookDryRun(prePushPath, prePushBlock, prePushGlobalScript, globalPrePushManagedBlockStart, "Pre-push")
		return nil
	}

	if err := installHookFile(preCommitPath, preCommitGlobalScript, preCommitBlock, globalManagedBlockStart, "Global pre-commit"); err != nil {
		return err
	}
	if err := installHookFile(prePushPath, prePushGlobalScript, prePushBlock, globalPrePushManagedBlockStart, "Global pre-push"); err != nil {
		return err
	}
	fmt.Printf("Installed global git hooks at %s\n", hooksDir)
	return nil
}

func installLocalHooks(dryRun bool) error {
	preCommitPath, err := localHookPath("pre-commit")
	if err != nil {
		return err
	}
	prePushPath, err := localHookPath("pre-push")
	if err != nil {
		return err
	}
	preCommitBlock := localPreCommitHookBlock()
	prePushBlock := localPrePushHookBlock()
	if dryRun {
		fmt.Println("Dry run: would install local git hooks")
		printHookDryRun(preCommitPath, preCommitBlock, preCommitLocalScript, localManagedBlockStart, "Pre-commit")
		printHookDryRun(prePushPath, prePushBlock, prePushLocalScript, localPrePushManagedBlockStart, "Pre-push")
		return nil
	}

	if err := installHookFile(preCommitPath, preCommitLocalScript, preCommitBlock, localManagedBlockStart, "Local pre-commit"); err != nil {
		return err
	}
	if err := installHookFile(prePushPath, prePushLocalScript, prePushBlock, localPrePushManagedBlockStart, "Local pre-push"); err != nil {
		return err
	}
	fmt.Printf("Installed local hooks at %s, %s\n", preCommitPath, prePushPath)
	return nil
}

func printHookDryRun(path string, block string, fullScript string, marker string, label string) {
	action, err := inspectHookInstall(path, marker)
	if err != nil {
		fmt.Printf("  Error inspecting %s hook: %v\n", label, err)
		return
	}
	switch action {
	case hookAlreadyInstalled:
		fmt.Printf("%s hook already contains git-hooks block: %s\n", label, path)
	case hookAppend:
		fmt.Printf("Would append executable file: %s\n", path)
		fmt.Println("Managed block:")
		fmt.Print(block)
	case hookWrite:
		fmt.Printf("Would write executable file: %s\n", path)
		fmt.Printf("%s hook script:\n", label)
		fmt.Print(fullScript)
	}
}

func installHookFile(path string, fullScript string, block string, marker string, label string) error {
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
			fmt.Printf("%s hook already contains git-hooks block: %s\n", label, path)
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

type hookInstallAction int

const (
	hookWrite hookInstallAction = iota
	hookAppend
	hookAlreadyInstalled
)

func inspectHookInstall(path string, marker string) (hookInstallAction, error) {
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return hookWrite, nil
		}
		return 0, err
	}
	if strings.Contains(string(existing), marker) {
		return hookAlreadyInstalled, nil
	}
	return hookAppend, nil
}
