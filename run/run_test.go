package run

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallGlobal(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	if err := Main([]string{"install", "--global"}); err != nil {
		t.Fatal(err)
	}

	hooksDir := filepath.Join(xdgConfig, ".git-hooks", "hooks")
	preCommit := filepath.Join(hooksDir, "pre-commit")
	content := mustRead(t, preCommit)
	if !strings.Contains(content, "git-hooks pre-commit run") {
		t.Fatalf("pre-commit hook does not call git-hooks pre-commit run:\n%s", content)
	}
	if content != preCommitGlobalScript {
		t.Fatalf("pre-commit hook does not match embedded global script\nwant:\n%s\ngot:\n%s", preCommitGlobalScript, content)
	}
	if !strings.Contains(content, "git rev-parse --git-common-dir") {
		t.Fatalf("global pre-commit hook does not call local hook first:\n%s", content)
	}
	if _, err := os.Stat(filepath.Join(xdgConfig, ".git-hooks", "pre-commit.d")); !os.IsNotExist(err) {
		t.Fatalf("global install should not create pre-commit.d, stat err: %v", err)
	}

	cmd := exec.Command("git", "config", "--global", "--get", "core.hooksPath")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git config --global --get core.hooksPath: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != hooksDir {
		t.Fatalf("core.hooksPath mismatch\nwant: %s\n got: %s", hooksDir, got)
	}
}

func TestInstallGlobalDryRun(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	if err := Main([]string{"install", "--global", "--dry-run"}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(xdgConfig, ".git-hooks")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create config root, stat err: %v", err)
	}
	if _, err := os.Stat(gitConfig); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write global git config, stat err: %v", err)
	}
}

func TestInstallGlobalAppendsConfiguredHooksPath(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	hooksDir := filepath.Join(t.TempDir(), "existing-hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	preCommit := filepath.Join(hooksDir, "pre-commit")
	existing := "#!/bin/sh\necho global\n"
	if err := os.WriteFile(preCommit, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}
	mustRun(t, "", "git", "config", "--global", "core.hooksPath", hooksDir)

	if err := Main([]string{"install", "--global"}); err != nil {
		t.Fatal(err)
	}

	content := mustRead(t, preCommit)
	if !strings.HasPrefix(content, existing) {
		t.Fatalf("existing global hook content was not preserved at the beginning:\n%s", content)
	}
	if !strings.Contains(content, globalManagedBlockStart) {
		t.Fatalf("global hook missing managed block:\n%s", content)
	}
	if !strings.Contains(content, "git rev-parse --git-common-dir") {
		t.Fatalf("global managed block does not call local hook first:\n%s", content)
	}
	if _, err := os.Stat(filepath.Join(xdgConfig, ".git-hooks", "hooks")); !os.IsNotExist(err) {
		t.Fatalf("install with configured core.hooksPath should not create default hooks dir, stat err: %v", err)
	}

	cmd := exec.Command("git", "config", "--global", "--get", "core.hooksPath")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git config --global --get core.hooksPath: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != hooksDir {
		t.Fatalf("core.hooksPath mismatch\nwant: %s\n got: %s", hooksDir, got)
	}

	if err := Main([]string{"install", "--global"}); err != nil {
		t.Fatal(err)
	}
	content = mustRead(t, preCommit)
	if strings.Count(content, globalManagedBlockStart) != 1 {
		t.Fatalf("global managed block was duplicated:\n%s", content)
	}
}

func TestInstallGlobalDryRunUsesConfiguredHooksPath(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)

	hooksDir := filepath.Join(t.TempDir(), "existing-hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	preCommit := filepath.Join(hooksDir, "pre-commit")
	existing := "#!/bin/sh\necho global\n"
	if err := os.WriteFile(preCommit, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}
	mustRun(t, "", "git", "config", "--global", "core.hooksPath", hooksDir)

	if err := Main([]string{"install", "--global", "--dry-run"}); err != nil {
		t.Fatal(err)
	}

	if content := mustRead(t, preCommit); content != existing {
		t.Fatalf("dry-run should not modify configured global hook\nwant:\n%s\ngot:\n%s", existing, content)
	}
	if _, err := os.Stat(filepath.Join(xdgConfig, ".git-hooks", "hooks")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create default hooks dir, stat err: %v", err)
	}
}

func TestInstallLocalAppendsExistingHook(t *testing.T) {
	repo := t.TempDir()
	mustRun(t, repo, "git", "init")
	hookPath := filepath.Join(repo, ".git", "hooks", "pre-commit")
	existing := "#!/bin/sh\necho local\n"
	if err := os.WriteFile(hookPath, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}
	withDir(t, repo, func() {
		if err := Main([]string{"install"}); err != nil {
			t.Fatal(err)
		}
	})

	content := mustRead(t, hookPath)
	if !strings.HasPrefix(content, existing) {
		t.Fatalf("existing hook content was not preserved at the beginning:\n%s", content)
	}
	if !strings.Contains(content, localManagedBlockStart) {
		t.Fatalf("local hook missing managed block:\n%s", content)
	}

	withDir(t, repo, func() {
		if err := Main([]string{"install"}); err != nil {
			t.Fatal(err)
		}
	})
	content = mustRead(t, hookPath)
	if strings.Count(content, localManagedBlockStart) != 1 {
		t.Fatalf("managed block was duplicated:\n%s", content)
	}
}

func TestInstallLocalCreatesEmbeddedScript(t *testing.T) {
	repo := t.TempDir()
	mustRun(t, repo, "git", "init")
	hookPath := filepath.Join(repo, ".git", "hooks", "pre-commit")
	withDir(t, repo, func() {
		if err := Main([]string{"install"}); err != nil {
			t.Fatal(err)
		}
	})
	if content := mustRead(t, hookPath); content != preCommitLocalScript {
		t.Fatalf("local hook does not match embedded local script\nwant:\n%s\ngot:\n%s", preCommitLocalScript, content)
	}
}

func TestInstallLocalDryRun(t *testing.T) {
	repo := t.TempDir()
	mustRun(t, repo, "git", "init")
	hookPath := filepath.Join(repo, ".git", "hooks", "pre-commit")
	withDir(t, repo, func() {
		if err := Main([]string{"install", "--dry-run"}); err != nil {
			t.Fatal(err)
		}
	})
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create local hook, stat err: %v", err)
	}
}

func TestPreCommitAddRunRemove(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	outputFile := filepath.Join(t.TempDir(), "hook.out")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)
	t.Setenv("GIT_HOOKS_TEST_OUTPUT", outputFile)

	err := Main([]string{
		"pre-commit",
		"add",
		"write-output",
		"sh",
		"-c",
		`echo ok >> "$GIT_HOOKS_TEST_OUTPUT"`,
	})
	if err != nil {
		t.Fatal(err)
	}

	hookPath := filepath.Join(xdgConfig, ".git-hooks", "pre-commit.d", "write-output")
	if _, err := os.Stat(hookPath); err != nil {
		t.Fatalf("expected hook file: %v", err)
	}

	if err := Main([]string{"pre-commit", "run"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(mustRead(t, outputFile)); got != "ok" {
		t.Fatalf("hook output mismatch: %q", got)
	}

	if err := Main([]string{"pre-commit", "remove", "write-output"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Fatalf("expected hook to be removed, stat err: %v", err)
	}
}

func TestPreCommitListShowsNamesAndCommands(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	if err := Main([]string{"pre-commit", "add", "check-keywords", "git-hook-detect-words", "shopee", "seamoney"}); err != nil {
		t.Fatal(err)
	}
	if err := Main([]string{"pre-commit", "add", "author", "git-hook-author-check", "--email", "ends-with:@example.com"}); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		if err := Main([]string{"pre-commit", "list"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "author\tgit-hook-author-check --email ends-with:@example.com\n") {
		t.Fatalf("missing author hook command:\n%s", output)
	}
	if !strings.Contains(output, "check-keywords\tgit-hook-detect-words shopee seamoney\n") {
		t.Fatalf("missing check-keywords hook command:\n%s", output)
	}
}

func TestPreCommitRunSkipsDuplicateSession(t *testing.T) {
	tmp := t.TempDir()
	xdgConfig := filepath.Join(tmp, "config")
	outputFile := filepath.Join(tmp, "hook.out")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)
	t.Setenv("GIT_HOOKS_TEST_OUTPUT", outputFile)
	t.Setenv("GIT_HOOKS_SESSION_ID", "test-session")
	t.Setenv("TMPDIR", tmp)

	err := Main([]string{
		"pre-commit",
		"add",
		"write-output",
		"sh",
		"-c",
		`echo ok >> "$GIT_HOOKS_TEST_OUTPUT"`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := Main([]string{"pre-commit", "run"}); err != nil {
		t.Fatal(err)
	}
	if err := Main([]string{"pre-commit", "run"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(mustRead(t, outputFile)); got != "ok" {
		t.Fatalf("duplicate session should run hook once, got: %q", got)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}

func withDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(old); err != nil {
			t.Fatal(err)
		}
	}()
	fn()
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}
