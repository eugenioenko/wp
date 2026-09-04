# wp

Git worktree workspace manager. Group worktrees from multiple repos into named workspaces, one per task or ticket.

## Install

```bash
curl -sSfL https://raw.githubusercontent.com/eugenioenko/wp/main/install.sh | sh
```

Or with Go:

```bash
go install github.com/eugenioenko/wp@latest
```

## Setup

```bash
wp init
```

Edit `~/.config/wp/config.toml` to register your repos:

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

## Workflow

```bash
wp create issue-123
wp add issue-123 myapp
wp add issue-123 backend

# cd into the workspace
cd ~/Workspaces/issue-123

# open your multiplexer (herdr, zellij, tmux, screen)
herdr

# when done
wp cleanup issue-123
```

Tip: add a shell alias for quick navigation:

```bash
wpcd() { cd ~/Workspaces/"$1"; }
```

## Commands

| Command | Description |
|---|---|
| `wp init` | Create default config file |
| `wp create <name>` | Create a new workspace |
| `wp add <name> <repo>` | Add a repo worktree to a workspace |
| `wp remove <name> <repo>` | Remove a repo worktree from a workspace |
| `wp status <name>` | Show branches and dirty state |
| `wp list` | List all workspaces |
| `wp cleanup <name>` | Remove all worktrees and delete workspace |
| `wp help` | Show help |
| `wp --version` | Print version |

## License

MIT
