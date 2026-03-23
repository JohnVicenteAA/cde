# cde

A CLI tool that creates and manages preconfigured tmux coding environments.

## Requirements

The following dependencies must be installed before using `cde`:

| Dependency | Required for | Link |
|------------|-------------|------|
| [Go](https://go.dev/) | Building/installing `cde` | https://go.dev/dl/ |
| [tmux](https://github.com/tmux/tmux) | All modes | https://github.com/tmux/tmux/wiki/Installing |
| [Neovim](https://neovim.io/) | `ide` mode | https://neovim.io/ |
| [Claude Code](https://docs.anthropic.com/en/docs/claude-code) | `wtree`, `mrepo` modes | https://docs.anthropic.com/en/docs/claude-code |
| [lazygit](https://github.com/jesseduffield/lazygit) | `wtree`, `mrepo` modes | https://github.com/jesseduffield/lazygit#installation |

## Install

```sh
go install github.com/JohnVicenteAA/cde@latest
```

## Commands

### `cde create`

Create a new tmux coding environment.

```sh
cde create [name] [flags]
```

If no name is given, the current directory name is used.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--mode` | `-m` | `ide` | Session mode (`ide`, `wtree`, `mrepo`) |
| `--num` | `-n` | `2` | Number of columns in wtree mode |

### `cde attach`

Attach to an existing cde tmux session.

```sh
cde attach --name <session_name>
```

### `cde session`

List and manage cde sessions.

```sh
cde session --list              # list all cde sessions
cde session delete --name <n>   # delete a specific session
cde session delete --all        # delete all cde sessions
```

## Modes

### ide

Default mode. Opens a tmux session with nvim and two shell panes.

```
+----------+----------+
|          |          |
|   nvim   |  shell   |
|          +----------+
|          |  shell   |
+----------+----------+
```

```sh
cde create              # uses current dir name
cde create myproject    # named session: myproject_ide
```

### wtree

Opens a tmux session with N paired columns, each containing a [Claude Code](https://claude.ai/claude-code) worktree on top and a lazygit instance watching that same worktree on the bottom. Requires a git repository.

```
+--------------------+--------------------+
|  claude --worktree |  claude --worktree |
|    <name>-0        |    <name>-1        |
|       (60%)        |       (60%)        |
+--------------------+--------------------+
|  lazygit -p        |  lazygit -p        |
|  .claude/worktrees/|  .claude/worktrees/|
|    <name>-0 (40%)  |    <name>-1 (40%)  |
+--------------------+--------------------+
```

```sh
cde create -m wtree         # 2 columns (default)
cde create -m wtree -n 3    # 3 columns
```

### mrepo

Multi-repo mode. Run from a parent directory that contains multiple git repos. Prompts you to select which repos to open, then creates a tmux session with one paired Claude Code + lazygit column per repo. Must be run from **outside** a git repo.

```
+--------------------+--------------------+
|  claude --worktree |  claude --worktree |
|    label-repo1     |    label-repo2     |
|       (60%)        |       (60%)        |
+--------------------+--------------------+
|  lazygit -p        |  lazygit -p        |
|  repo1/.claude/    |  repo2/.claude/    |
|  worktrees/        |  worktrees/        |
|    label-repo1     |    label-repo2     |
|       (40%)        |       (40%)        |
+--------------------+--------------------+
```

```sh
cde create -m mrepo    # prompts for repo selection and session label
```

## Session naming

Sessions are named `{name}_{mode}`, with dots replaced by underscores. This allows running multiple modes side by side:

```sh
cde create -m ide      # session: dirname_ide
cde create -m wtree    # session: dirname_wtree
cde create -m mrepo    # session: mrepo_label_repo1_repo2
```

If a session already exists, you'll be prompted to reattach or replace it.

## Suggested tmux configuration

The following `.tmux.conf` is optional but recommended for a better experience with `cde`:

```tmux
# List of plugins
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-sensible'
set -g @plugin 'christoomey/vim-tmux-navigator'
set -g @plugin 'tmux-plugins/tmux-resurrect'

set -g mouse on
set -g set-titles on
set -g set-titles-string "#{window_name}"
# Initialize TMUX plugin manager (keep this line at the very bottom of tmux.conf)
run '~/.tmux/plugins/tpm/tpm'
```
