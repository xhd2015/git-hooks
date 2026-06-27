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

	withNonRepoDir(t, func() {
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
	})
}

func TestPreCommitListShowsNamesAndCommands(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
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
	})
}

func TestPreCommitRunSkipsDuplicateSession(t *testing.T) {
	tmp := t.TempDir()
	xdgConfig := filepath.Join(tmp, "config")
	outputFile := filepath.Join(tmp, "hook.out")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)
	t.Setenv("GIT_HOOKS_TEST_OUTPUT", outputFile)
	t.Setenv("GIT_HOOKS_SESSION_ID", "test-session")
	t.Setenv("TMPDIR", tmp)

	withNonRepoDir(t, func() {
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
	})
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

func withNonRepoDir(t *testing.T, fn func()) {
	t.Helper()
	withDir(t, t.TempDir(), fn)
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

func TestExtractHookPrefix(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNum   int
		wantHas   bool
	}{
		{"numeric prefix", "01-binary-check", 1, true},
		{"two digit numeric", "10-detect-words", 10, true},
		{"large number", "100-custom", 100, true},
		{"no prefix", "author-check", 0, false},
		{"number without dash", "123no-dash", 0, false},
		{"dash without number", "-no-number", 0, false},
		{"starts with letter", "a01-something", 0, false},
		{"empty", "", 0, false},
		{"just dash", "-", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, has := extractHookPrefix(tt.input)
			if has != tt.wantHas {
				t.Errorf("hasPrefix: got %v, want %v", has, tt.wantHas)
			}
			if n != tt.wantNum {
				t.Errorf("num: got %d, want %d", n, tt.wantNum)
			}
		})
	}
}

func TestListManagedHooksNumericOrder(t *testing.T) {
	dir := t.TempDir()

	createHook := func(name string) {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\necho "+name+"\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	createHook("10-detect-words")
	createHook("z-author-check")
	createHook("01-binary-check")
	createHook("02-submodule-check")
	createHook("misc-hook")

	hooks, err := listManagedHooks(dir)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"01-binary-check", "02-submodule-check", "10-detect-words", "misc-hook", "z-author-check"}
	if len(hooks) != len(expected) {
		t.Fatalf("got %d hooks, want %d: %v", len(hooks), len(expected), hooks)
	}
	for i, want := range expected {
		if hooks[i] != want {
			t.Errorf("index %d: got %q, want %q\nfull: %v", i, hooks[i], want, hooks)
		}
	}
}

func TestListManagedHooksNumericOrderSamePrefix(t *testing.T) {
	dir := t.TempDir()

	createHook := func(name string) {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\necho "+name+"\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	createHook("01-b")
	createHook("01-a")
	createHook("01-c")

	hooks, err := listManagedHooks(dir)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"01-a", "01-b", "01-c"}
	if len(hooks) != len(expected) {
		t.Fatalf("got %d hooks, want %d: %v", len(hooks), len(expected), hooks)
	}
	for i, want := range expected {
		if hooks[i] != want {
			t.Errorf("index %d: got %q, want %q", i, hooks[i], want)
		}
	}
}

func TestDisplayHookName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01-binary-check", "binary-check"},
		{"10-detect-words", "detect-words"},
		{"100-custom", "custom"},
		{"author-check", "author-check"},
		{"no-prefix", "no-prefix"},
		{"123", "123"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := displayHookName(tt.input)
			if got != tt.expected {
				t.Errorf("displayHookName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveHookName(t *testing.T) {
	dir := t.TempDir()
	createHook := func(name string) {
		path := filepath.Join(dir, name)
		os.WriteFile(path, []byte("#!/bin/sh\necho ok\n"), 0755)
	}
	createHook("01-binary-check")
	createHook("10-detect-words")
	createHook("author-check")

	t.Run("prefixed hook", func(t *testing.T) {
		got, err := resolveHookName(dir, "binary-check")
		if err != nil {
			t.Fatal(err)
		}
		if got != "01-binary-check" {
			t.Errorf("got %q, want 01-binary-check", got)
		}
	})
	t.Run("non-prefixed hook", func(t *testing.T) {
		got, err := resolveHookName(dir, "author-check")
		if err != nil {
			t.Fatal(err)
		}
		if got != "author-check" {
			t.Errorf("got %q, want author-check", got)
		}
	})
	t.Run("not found", func(t *testing.T) {
		_, err := resolveHookName(dir, "nonexistent")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ambiguous", func(t *testing.T) {
		createHook("02-binary-check")
		_, err := resolveHookName(dir, "binary-check")
		if err == nil {
			t.Fatal("expected error for ambiguous name")
		}
	})
}

func TestPreCommitRename(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "01-check", "echo", "test"})

		err := Main([]string{"pre-commit", "rename", "check", "validate"})
		if err != nil {
			t.Fatal(err)
		}

		dir := filepath.Join(xdgConfig, ".git-hooks", "pre-commit.d")
		oldPath := filepath.Join(dir, "01-check")
		newPath := filepath.Join(dir, "01-validate")

		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Fatalf("old file should not exist")
		}
		if _, err := os.Stat(newPath); err != nil {
			t.Fatalf("new file should exist: %v", err)
		}
	})
}

func TestPreCommitRenameNonPrefixed(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "check", "echo", "test"})

		err := Main([]string{"pre-commit", "rename", "check", "verify"})
		if err != nil {
			t.Fatal(err)
		}

		dir := filepath.Join(xdgConfig, ".git-hooks", "pre-commit.d")
		newPath := filepath.Join(dir, "verify")
		if _, err := os.Stat(newPath); err != nil {
			t.Fatalf("renamed file should exist: %v", err)
		}
	})
}

func TestPreCommitRenameNotFound(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		err := Main([]string{"pre-commit", "rename", "nonexistent", "newname"})
		if err == nil {
			t.Fatal("expected error for nonexistent hook")
		}
	})
}

func TestPreCommitUpDown(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "a", "echo", "content-a"})
		Main([]string{"pre-commit", "add", "b", "echo", "content-b"})
		Main([]string{"pre-commit", "add", "c", "echo", "content-c"})

		dir := filepath.Join(xdgConfig, ".git-hooks", "pre-commit.d")

		readContent := func(name string) string {
			t.Helper()
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			return string(data)
		}

		err := Main([]string{"pre-commit", "up", "b"})
		if err != nil {
			t.Fatal(err)
		}

		// up swaps positions: b <-> a
		if s := readContent("a"); !strings.Contains(s, "content-b") {
			t.Fatalf("after up b, file a should contain content-b: %s", s)
		}
		if s := readContent("b"); !strings.Contains(s, "content-a") {
			t.Fatalf("after up b, file b should contain content-a: %s", s)
		}

		// b is now at display name "a", so down a moves it back
		err = Main([]string{"pre-commit", "down", "a"})
		if err != nil {
			t.Fatal(err)
		}

		if s := readContent("a"); !strings.Contains(s, "content-a") {
			t.Fatalf("after down a, file a should contain content-a again: %s", s)
		}
		if s := readContent("b"); !strings.Contains(s, "content-b") {
			t.Fatalf("after down a, file b should contain content-b again: %s", s)
		}
	})
}

func TestPreCommitUpNonPrefixedWithPrefixed(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "01-first", "echo", "first-content"})
		Main([]string{"pre-commit", "add", "plain", "echo", "plain-content"})

		dir := filepath.Join(xdgConfig, ".git-hooks", "pre-commit.d")

		readContent := func(name string) string {
			t.Helper()
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			return string(data)
		}

		err := Main([]string{"pre-commit", "up", "plain"})
		if err != nil {
			t.Fatal(err)
		}

		if s := readContent("01-first"); !strings.Contains(s, "plain-content") {
			t.Fatalf("after up, 01-first should contain plain-content: %s", s)
		}
		if s := readContent("plain"); !strings.Contains(s, "first-content") {
			t.Fatalf("after up, plain should contain first-content: %s", s)
		}
	})
}

func TestPreCommitUpFirstError(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "01-first", "echo", "first"})
		Main([]string{"pre-commit", "add", "10-second", "echo", "second"})

		err := Main([]string{"pre-commit", "up", "first"})
		if err == nil {
			t.Fatal("expected error when moving first hook up")
		}
	})
}

func TestPreCommitDownLastError(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "01-first", "echo", "first"})
		Main([]string{"pre-commit", "add", "10-second", "echo", "second"})

		err := Main([]string{"pre-commit", "down", "second"})
		if err == nil {
			t.Fatal("expected error when moving last hook down")
		}
	})
}

func TestPreCommitTop(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "01-first", "echo", "first"})
		Main([]string{"pre-commit", "add", "10-second", "echo", "second"})
		Main([]string{"pre-commit", "add", "plain", "echo", "plain"})

		dir := filepath.Join(xdgConfig, ".git-hooks", "pre-commit.d")

		err := Main([]string{"pre-commit", "top", "plain"})
		if err != nil {
			t.Fatal(err)
		}

		hooks, _ := listManagedHooks(dir)
		expected := []string{"00-plain", "01-first", "10-second"}
		for i, w := range expected {
			if i >= len(hooks) || hooks[i] != w {
				t.Fatalf("position %d: got %v, want %v (full: %v)", i, hooks, w, hooks)
			}
		}
	})
}

func TestPreCommitTopAlreadyFirst(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "01-first", "echo", "first"})
		Main([]string{"pre-commit", "add", "10-second", "echo", "second"})

		err := Main([]string{"pre-commit", "top", "first"})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPreCommitListStripsPrefix(t *testing.T) {
	xdgConfig := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfig)

	withNonRepoDir(t, func() {
		Main([]string{"pre-commit", "add", "01-binary-check", "echo", "binary"})
		Main([]string{"pre-commit", "add", "author-check", "echo", "author"})

		output := captureStdout(t, func() {
			Main([]string{"pre-commit", "list"})
		})

		if !strings.Contains(output, "binary-check") {
			t.Errorf("list should show display name, got:\n%s", output)
		}
		if strings.Contains(output, "01-binary-check") {
			t.Errorf("list should NOT show numeric prefix, got:\n%s", output)
		}
		if !strings.Contains(output, "author-check") {
			t.Errorf("list should show non-prefixed name, got:\n%s", output)
		}
	})
}
