# cde

A CLI tool that creates preconfigured tmux coding environments.

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
| `--mode` | `-m` | `ide` | Session mode (`ide`, `agent`) |
| `--num` | `-n` | `2` | Number of `claude --worktree` panes in agent mode |

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

### agent

Opens a tmux session with N [Claude Code](https://claude.ai/claude-code) worktree panes, lazygit, and lazydocker. Requires a git repository.

```
+--------------------+--------------------+
| claude --worktree  | claude --worktree  |
+--------------------+--------------------+
|      lazygit       |    lazydocker      |
+--------------------+--------------------+
```

```sh
cde -m agent         # 2 claude panes (default)
cde -m agent -n 4    # 4 claude panes
```

## Session naming

Sessions are named `{name}_{mode}`, with dots replaced by underscores. This allows running both modes side by side:

```sh
cde -m ide      # session: dirname_ide
cde -m agent    # session: dirname_agent
```

If a session already exists, you'll be prompted to reattach or replace it.

## Dependencies

- [tmux](https://github.com/tmux/tmux)
- [neonvim](https://neonvim.io/) (ide mode)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (agent mode)
- [lazygit](https://github.com/jesseduffield/lazygit) (agent mode)
- [lazydocker](https://github.com/jesseduffield/lazydocker) (agent mode)
