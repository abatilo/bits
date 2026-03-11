# bits

Minimal file-based task tracker for AI agents. Go CLI using cobra.

## Architecture

- `cmd/bits/` — CLI entry point and all commands
- `internal/storage/` — File-based task storage, project root detection, markdown serialization
- `internal/session/` — Session ownership and drain state (session.json)
- `internal/task/` — Task model and ID generation
- `internal/deps/` — Dependency graph, cycle detection, readiness sorting
- `internal/output/` — Human and JSON output formatters
- `plugins/bits/` — Claude Code plugin (hooks, skills, slash commands)

## Storage

Tasks are markdown files with YAML frontmatter stored at `~/.bits/<sanitized-project-path>/`.

Project scoping uses `FindProjectRoot()` in `internal/storage/project.go`, which walks up from cwd looking for `.git` as either a directory (normal repo) or a regular file (git worktree). Each worktree gets isolated storage since worktree paths are unique.

## Testing

- All tests use stdlib `testing` only (no testify)
- White-box testing (same package, `//nolint:testpackage` directive)
- `t.TempDir()` for temp directories, `t.Chdir()` for path-dependent tests
- `//nolint:govet // Intentional shadow in subtest` on `:=` declarations in subtests
