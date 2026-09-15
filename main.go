package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v3"
)

type Config struct {
	WorkspaceRoot string            `toml:"workspace_root"`
	Repos         map[string]string `toml:"repos"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "wp", "config.toml")
}

func loadConfig() (*Config, error) {
	path := configPath()
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not read config at %s: %w", path, err)
	}
	cfg.WorkspaceRoot = expandHome(cfg.WorkspaceRoot)
	for k, v := range cfg.Repos {
		cfg.Repos[k] = expandHome(v)
	}
	return &cfg, nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func workspacePath(cfg *Config, name string) string {
	return filepath.Join(cfg.WorkspaceRoot, name)
}

func worktreePath(cfg *Config, workspace, repo string) string {
	return filepath.Join(cfg.WorkspaceRoot, workspace, repo)
}

func repoSource(cfg *Config, repo string) (string, error) {
	src, ok := cfg.Repos[repo]
	if !ok {
		available := make([]string, 0, len(cfg.Repos))
		for k := range cfg.Repos {
			available = append(available, k)
		}
		return "", fmt.Errorf("unknown repo %q (available: %s)", repo, strings.Join(available, ", "))
	}
	return src, nil
}

func activeWorktrees(cfg *Config, workspace string) ([]string, error) {
	wsPath := workspacePath(cfg, workspace)
	entries, err := os.ReadDir(wsPath)
	if err != nil {
		return nil, err
	}
	var repos []string
	for _, e := range entries {
		if e.IsDir() {
			gitPath := filepath.Join(wsPath, e.Name(), ".git")
			if _, err := os.Stat(gitPath); err == nil {
				repos = append(repos, e.Name())
			}
		}
	}
	return repos, nil
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

var version = "dev"

const ansiGreen = "\033[32m"
const ansiYellow = "\033[33m"
const ansiRed = "\033[31m"
const ansiCyan = "\033[36m"
const ansiBold = "\033[1m"
const ansiReset = "\033[0m"

func main() {
	app := &cli.Command{
		Name:    "wp",
		Usage:   "Git worktree workspace manager",
		Version: version,
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cli.ShowAppHelp(cmd)
		},
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Create default config file",
				Action: cmdInit,
			},
			{
				Name:      "start",
				Usage:     "Create a workspace and add repo worktrees",
				ArgsUsage: "<name> <repo> [repo...]",
				Action:    cmdStart,
			},
			{
				Name:      "create",
				Usage:     "Create a new workspace",
				ArgsUsage: "<name>",
				Action:    cmdCreate,
			},
			{
				Name:      "add",
				Usage:     "Add a repo worktree to a workspace",
				ArgsUsage: "<workspace> <repo>",
				Action:    cmdAdd,
			},
			{
				Name:      "remove",
				Usage:     "Remove a repo worktree from a workspace",
				ArgsUsage: "<workspace> <repo>",
				Action:    cmdRemove,
			},
			{
				Name:      "cleanup",
				Usage:     "Remove all worktrees and delete a workspace",
				ArgsUsage: "<name>",
				Action:    cmdCleanup,
			},
			{
				Name:   "list",
				Usage:  "List all workspaces",
				Action: cmdList,
			},
			{
				Name:      "status",
				Usage:     "Show status of a workspace",
				ArgsUsage: "<name>",
				Action:    cmdStatus,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", ansiRed, ansiReset, err)
		os.Exit(1)
	}
}

func cmdInit(_ context.Context, cmd *cli.Command) error {
	path := configPath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %s", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	content := `workspace_root = "~/Workspaces"

[repos]
# name = "/path/to/repo"
# example:
# myapp = "~/Documents/myapp"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Created config at %s\n", path)
	return nil
}

func cmdCreate(_ context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	if name == "" {
		return fmt.Errorf("usage: wp create <name>")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	return createWorkspace(cfg, name)
}

func createWorkspace(cfg *Config, name string) error {
	wsPath := workspacePath(cfg, name)
	if _, err := os.Stat(wsPath); err == nil {
		return fmt.Errorf("workspace %q already exists at %s", name, wsPath)
	}

	if err := os.MkdirAll(wsPath, 0o755); err != nil {
		return err
	}
	fmt.Printf("%s✓%s Created workspace %s%s%s at %s\n", ansiGreen, ansiReset, ansiBold, name, ansiReset, wsPath)
	return nil
}

func cmdAdd(_ context.Context, cmd *cli.Command) error {
	workspace := cmd.Args().Get(0)
	repo := cmd.Args().Get(1)
	if workspace == "" || repo == "" {
		return fmt.Errorf("usage: wp add <workspace> <repo>")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	return addRepo(cfg, workspace, repo)
}

func addRepo(cfg *Config, workspace, repo string) error {
	src, err := repoSource(cfg, repo)
	if err != nil {
		return err
	}

	wsPath := workspacePath(cfg, workspace)
	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		return fmt.Errorf("workspace %q does not exist (run: wp create %s)", workspace, workspace)
	}

	wtPath := worktreePath(cfg, workspace, repo)
	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("repo %q already in workspace %q", repo, workspace)
	}

	branch := workspace
	checkCmd := exec.Command("git", "-C", src, "rev-parse", "--verify", branch)
	branchExists := checkCmd.Run() == nil

	if branchExists {
		err = runGit("-C", src, "worktree", "add", wtPath, branch)
	} else {
		err = runGit("-C", src, "worktree", "add", "-b", branch, wtPath)
	}
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	fmt.Printf("%s✓%s Added %s%s%s to workspace %s (branch: %s)\n",
		ansiGreen, ansiReset, ansiBold, repo, ansiReset, workspace, branch)
	return nil
}

func cmdStart(_ context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	repos := cmd.Args().Tail()
	if name == "" || len(repos) == 0 {
		return fmt.Errorf("usage: wp start <name> <repo> [repo...]")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := createWorkspace(cfg, name); err != nil {
		return err
	}
	for _, repo := range repos {
		if err := addRepo(cfg, name, repo); err != nil {
			return err
		}
	}
	return nil
}

func cmdRemove(_ context.Context, cmd *cli.Command) error {
	workspace := cmd.Args().Get(0)
	repo := cmd.Args().Get(1)
	if workspace == "" || repo == "" {
		return fmt.Errorf("usage: wp remove <workspace> <repo>")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	src, err := repoSource(cfg, repo)
	if err != nil {
		return err
	}

	wtPath := worktreePath(cfg, workspace, repo)
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("repo %q not found in workspace %q", repo, workspace)
	}

	if err := runGit("-C", src, "worktree", "remove", wtPath); err != nil {
		return fmt.Errorf("failed to remove worktree (has uncommitted changes?): %w", err)
	}

	fmt.Printf("%s✓%s Removed %s from workspace %s\n", ansiGreen, ansiReset, repo, workspace)
	return nil
}

func cmdCleanup(_ context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	if name == "" {
		return fmt.Errorf("usage: wp cleanup <name>")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	wsPath := workspacePath(cfg, name)
	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		return fmt.Errorf("workspace %q does not exist", name)
	}

	repos, err := activeWorktrees(cfg, name)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		src, srcErr := repoSource(cfg, repo)
		if srcErr != nil {
			fmt.Fprintf(os.Stderr, "%swarn:%s skipping %s: %v\n", ansiYellow, ansiReset, repo, srcErr)
			continue
		}
		wtPath := worktreePath(cfg, name, repo)
		if err := runGit("-C", src, "worktree", "remove", "--force", wtPath); err != nil {
			fmt.Fprintf(os.Stderr, "%swarn:%s could not remove worktree for %s: %v\n", ansiYellow, ansiReset, repo, err)
		} else {
			fmt.Printf("  Removed worktree %s\n", repo)
		}
	}

	if err := os.RemoveAll(wsPath); err != nil {
		return fmt.Errorf("could not remove workspace dir: %w", err)
	}

	fmt.Printf("%s✓%s Cleaned up workspace %s%s%s\n", ansiGreen, ansiReset, ansiBold, name, ansiReset)
	return nil
}

func cmdList(_ context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(cfg.WorkspaceRoot)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No workspaces yet.")
			return nil
		}
		return err
	}

	found := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		repos, _ := activeWorktrees(cfg, name)
		repoStr := "(empty)"
		if len(repos) > 0 {
			repoStr = strings.Join(repos, ", ")
		}
		fmt.Printf("  %s%s%s  [%s]\n", ansiBold, name, ansiReset, repoStr)
		found = true
	}

	if !found {
		fmt.Println("No workspaces yet.")
	}
	return nil
}

func cmdStatus(_ context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	if name == "" {
		return fmt.Errorf("usage: wp status <name>")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	wsPath := workspacePath(cfg, name)
	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		return fmt.Errorf("workspace %q does not exist", name)
	}

	repos, err := activeWorktrees(cfg, name)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Printf("%s%s%s: (empty)\n", ansiBold, name, ansiReset)
		return nil
	}

	fmt.Printf("%s%s%s:\n", ansiBold, name, ansiReset)
	for _, repo := range repos {
		wtPath := worktreePath(cfg, name, repo)

		branch, _ := gitOutput(wtPath, "rev-parse", "--abbrev-ref", "HEAD")
		if branch == "" {
			branch = "???"
		}

		statusOut, _ := gitOutput(wtPath, "status", "--porcelain")
		state := ansiGreen + "clean" + ansiReset
		if statusOut != "" {
			state = ansiYellow + "dirty" + ansiReset
		}

		fmt.Printf("  %-20s %s%-12s%s %s\n", repo, ansiCyan, branch, ansiReset, state)
	}
	return nil
}
