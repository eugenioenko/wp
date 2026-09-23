package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
)

// The generated shell scripts filter the current word themselves. The callback
// receives only the completed positional arguments before that word.
func completionArgs(cmd *cli.Command) ([]string, bool) {
	args := cmd.Args().Slice()
	if len(args) > 0 && strings.HasPrefix(args[len(args)-1], "-") {
		completeFlags(cmd, args[len(args)-1])
		return nil, true
	}
	return args, false
}

func completeFlags(cmd *cli.Command, prefix string) {
	for _, flag := range cmd.Flags {
		if visible, ok := flag.(cli.VisibleFlag); ok && !visible.IsVisible() {
			continue
		}
		for _, name := range flag.Names() {
			dashes := "--"
			if len(name) == 1 {
				dashes = "-"
			}
			candidate := dashes + name
			if candidate != prefix && strings.HasPrefix(candidate, prefix) {
				fmt.Fprintln(cmd.Root().Writer, candidate)
			}
		}
	}
}

func printCompletions(cmd *cli.Command, names []string) {
	for _, name := range names {
		fmt.Fprintln(cmd.Root().Writer, name)
	}
}

func workspaceNames(cfg *Config) []string {
	entries, err := os.ReadDir(cfg.WorkspaceRoot)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

func repoNames(cfg *Config, excluded map[string]bool) []string {
	var names []string
	for name := range cfg.Repos {
		if !excluded[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func completeWorkspace(ctx context.Context, cmd *cli.Command) {
	args, handled := completionArgs(cmd)
	if handled || len(args) != 0 {
		return
	}
	if cfg, err := loadConfig(); err == nil {
		printCompletions(cmd, workspaceNames(cfg))
	}
}

func completeStart(ctx context.Context, cmd *cli.Command) {
	args, handled := completionArgs(cmd)
	if handled || len(args) == 0 {
		return
	}
	if cfg, err := loadConfig(); err == nil {
		excluded := make(map[string]bool, len(args)-1)
		for _, name := range args[1:] {
			excluded[name] = true
		}
		printCompletions(cmd, repoNames(cfg, excluded))
	}
}

func completeAdd(ctx context.Context, cmd *cli.Command) {
	args, handled := completionArgs(cmd)
	if handled {
		return
	}
	if len(args) == 0 {
		completeWorkspace(ctx, cmd)
		return
	}
	if len(args) != 1 {
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		return
	}
	active, err := activeWorktrees(cfg, args[0])
	if err != nil {
		return
	}
	excluded := make(map[string]bool, len(active))
	for _, name := range active {
		excluded[name] = true
	}
	printCompletions(cmd, repoNames(cfg, excluded))
}

func completeRemove(ctx context.Context, cmd *cli.Command) {
	args, handled := completionArgs(cmd)
	if handled {
		return
	}
	if len(args) == 0 {
		completeWorkspace(ctx, cmd)
		return
	}
	if len(args) != 1 {
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		return
	}
	active, err := activeWorktrees(cfg, args[0])
	if err != nil {
		return
	}
	var names []string
	for _, name := range active {
		if _, configured := cfg.Repos[name]; configured {
			names = append(names, name)
		}
	}
	printCompletions(cmd, names)
}
