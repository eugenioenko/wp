# wp

Git worktree workspace manager. Group worktrees from multiple repos into named workspaces, one per task or ticket.

## Install

```bash
go install github.com/eugenioenko/wp@latest
```

## Setup

```bash
wp init
```

Creates `~/.config/wp/config.toml`:

```toml
workspace_root = "~/Workspaces"

[repos]
myapp = "~/Documents/myapp"
backend = "~/Documents/backend"
```

## Usage

```bash
wp create issue-1              # Create a workspace
wp add issue-1 myapp           # Add repo worktree (creates branch issue-1)
wp add issue-1 backend         # Add another repo
wp status issue-1              # Show branches and dirty state
wp list                        # List all workspaces
wp remove issue-1 backend      # Remove a repo worktree
wp cleanup issue-1             # Remove all worktrees and workspace dir
```

Open the workspace folder in your terminal multiplexer of choice (herdr, zellij, tmux, screen) and work from there.

## License

MIT
