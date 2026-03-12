# CLAUDE.md

## Project

`cde` is a Go CLI tool that creates preconfigured tmux coding environments. It uses Cobra for CLI parsing.

## Structure

- `main.go` — entrypoint
- `cmd/root.go` — CLI flags, session naming, shared helpers
- `cmd/tmux.go` — `TmuxRunner` interface and real implementation
- `cmd/ide.go` — ide mode layout
- `cmd/wtree.go` — wtree mode layout
- `cmd/mock_tmux_test.go` — mock `TmuxRunner` for tests

## Commands

```sh
go build -o cde .       # build
go install .             # install to $GOPATH/bin
go test ./... -v         # run tests
```

## Conventions

- Tmux interactions go through the `TmuxRunner` interface — never call `exec.Command("tmux", ...)` directly in mode files
- Tests use `mockTmux` to verify the exact sequence of tmux commands without a real tmux session
- `isGitRepo` is a `var func()` so tests can override it
