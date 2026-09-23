package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellCompletion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, "workspaces")
	if err := os.MkdirAll(filepath.Join(root, "issue-1", "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "issue-2"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Git worktrees have a .git file, rather than a .git directory.
	if err := os.WriteFile(filepath.Join(root, "issue-1", "alpha", ".git"), []byte("gitdir: elsewhere"), 0o644); err != nil {
		t.Fatal(err)
	}
	configDir := filepath.Join(home, ".config", "wp")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := fmt.Sprintf("workspace_root = %q\n\n[repos]\nalpha = \"/tmp/alpha\"\nbeta = \"/tmp/beta\"\n", root)
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	complete := func(args ...string) string {
		t.Helper()
		app := newApp()
		var output bytes.Buffer
		app.Writer = &output
		argv := append([]string{"wp"}, args...)
		argv = append(argv, "--generate-shell-completion")
		if err := app.Run(context.Background(), argv); err != nil {
			t.Fatal(err)
		}
		return output.String()
	}

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"start"}, ""},
		{[]string{"start", "issue-3"}, "alpha\nbeta\n"},
		{[]string{"start", "issue-3", "alpha"}, "beta\n"},
		{[]string{"add"}, "issue-1\nissue-2\n"},
		{[]string{"add", "issue-1"}, "beta\n"},
		{[]string{"remove"}, "issue-1\nissue-2\n"},
		{[]string{"remove", "issue-1"}, "alpha\n"},
		{[]string{"status"}, "issue-1\nissue-2\n"},
		{[]string{"cleanup"}, "issue-1\nissue-2\n"},
		{[]string{"status", "issue-1"}, ""},
		{[]string{"add", "missing"}, ""},
	}
	for _, test := range tests {
		if got := complete(test.args...); got != test.want {
			t.Errorf("completion for %q = %q; want %q", test.args, got, test.want)
		}
	}

	// Every request reads current config and workspace directories.
	config += "gamma = \"/tmp/gamma\"\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "issue-3"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := complete("start", "issue-4"); got != "alpha\nbeta\ngamma\n" {
		t.Errorf("completion after config edit = %q", got)
	}
	if got := complete("status"); got != "issue-1\nissue-2\nissue-3\n" {
		t.Errorf("completion after workspace creation = %q", got)
	}

	if got := complete("add", "issue-1", "--he"); !strings.Contains(got, "--help") {
		t.Errorf("flag completion after a workspace = %q", got)
	}

	t.Setenv("HOME", t.TempDir())
	if got := complete("status"); got != "" {
		t.Errorf("completion without config = %q", got)
	}
}

func TestCompletionShownInUsage(t *testing.T) {
	app := newApp()
	var output bytes.Buffer
	app.Writer = &output
	if err := app.Run(context.Background(), []string{"wp"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "completion") || !strings.Contains(output.String(), "Generate shell completion scripts") {
		t.Fatalf("completion missing from usage:\n%s", output.String())
	}
}
