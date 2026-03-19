# cde

A CLI tool that creates preconfigured tmux coding environments.

## Requirements

The following dependencies must be installed before using `cde`:

| Dependency | Required for | Link |
|------------|-------------|------|
| [Go](https://go.dev/) | Building/installing `cde` | https://go.dev/dl/ |
| [tmux](https://github.com/tmux/tmux) | All modes | https://github.com/tmux/tmux/wiki/Installing |
| [Neovim](https://neovim.io/) | `ide` mode | https://neovim.io/ |
| [Claude Code](https://docs.anthropic.com/en/docs/claude-code) | `wtree` mode | https://docs.anthropic.com/en/docs/claude-code |
| [lazygit](https://github.com/jesseduffield/lazygit) | `wtree` mode | https://github.com/jesseduffield/lazygit#installation |

## Install

```sh
go install github.com/JohnVicenteAA/cde@latest
```

## Usage

```sh
cde [name] [flags]
```

If no name is given, the current directory name is used.

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--mode` | `-m` | `ide` | Session mode (`ide`, `wtree`) |
| `--num` | `-n` | `2` | Number of columns in wtree mode |

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
cde              # uses current dir name
cde myproject    # named session: myproject_ide
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
cde -m wtree         # 2 columns (default)
cde -m wtree -n 3    # 3 columns
```

## Session naming

Sessions are named `{name}_{mode}`, with dots replaced by underscores. This allows running both modes side by side:

```sh
cde -m ide      # session: dirname_ide
cde -m wtree    # session: dirname_wtree
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
