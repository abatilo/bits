# bits

> **Minimal task tracking for AI agents — one active task, zero distractions**

bits is a file-based task tracker designed for AI coding agents. It enforces a
single active task at a time, preventing agents from context-switching and
losing focus.

## Philosophy

- **One task at a time**: Only one task can be active. This constraint keeps
  agents focused and prevents half-finished work.
- **File-based storage**: Tasks are Markdown files with YAML frontmatter. Human
  readable, no database required.
- **Git repository scoping**: Tasks are scoped to the git repository you're
  working in. Different projects have different task lists.
- **Dependencies with cycle detection**: Tasks can depend on other tasks. bits
  prevents circular dependencies.

## Installation

### From source

```bash
go install github.com/abatilo/bits/cmd/bits@latest
```

### Build locally

```bash
git clone https://github.com/abatilo/bits.git
cd bits
go build -o bits ./cmd/bits
```

## Quick Start

```bash
# Initialize bits for this repository
bits init

# Add a task
bits add "Fix the login bug" -d "Users can't log in with email addresses containing a plus sign"

# See what's ready to work on
bits ready

# Start working on a task
bits claim abc123

# When done, close it with a reason
bits close abc123 "Fixed in commit 1a2b3c4"
```

## Command Reference

All commands support `--json` for machine-readable output.

### init

Initialize bits for the current git repository.

```bash
bits init
bits init --force  # Reinitialize even if already exists
```

### add

Create a new task.

```bash
bits add "Task title"
bits add "Task title" -d "Detailed description"
bits add "Urgent fix" -p critical  # Priority: critical, high, medium, low
```

Output:
```
[abc123] Task title
  Status:   open
  Priority: medium
  Created:  2025-01-19 10:30
```

### list

List tasks with optional status filters.

```bash
bits list              # All tasks
bits list --open       # Only open tasks
bits list --active     # Only active tasks
bits list --closed     # Only closed tasks
```

Output:
```
[*] P1 [def456] Implement caching
[ ] P2 [abc123] Fix the login bug
[X] P3 [ghi789] Update readme
```

Status icons: `[ ]` open, `[*]` active, `[X]` closed
Priority marks: `P0` critical, `P1` high, `P2` medium, `P3` low

### show

Display full details of a task.

```bash
bits show abc123
```

Output:
```
[abc123] Fix the login bug
  Status:   open
  Priority: medium
  Created:  2025-01-19 10:30
  Depends:  xyz789

Users can't log in with email addresses containing a plus sign.
```

### ready

List tasks that are ready to be worked on (open, with all dependencies closed).

```bash
bits ready
```

### claim

Start working on a task. The task must be open and all its dependencies must be
closed. Only one task can be active at a time.

```bash
bits claim abc123
```

Errors:
- If another task is already active
- If the task has unclosed dependencies
- If the task is not in `open` status

### release

Stop working on a task without completing it. Returns it to `open` status.

```bash
bits release abc123
```

### close

Complete a task. Requires a reason explaining what was done.

```bash
bits close abc123 "Fixed in commit 1a2b3c4"
```

The task must be in `active` status to be closed.

### dep

Add a dependency. The first task will depend on the second task.

```bash
bits dep abc123 xyz789  # abc123 depends on xyz789
```

bits prevents circular dependencies. If adding the dependency would create a
cycle, the command fails.

### undep

Remove a dependency.

```bash
bits undep abc123 xyz789
```

### rm

Remove a task and clean up any references to it in other tasks' dependencies.

```bash
bits rm abc123
```

### prune

Remove all closed tasks.

```bash
bits prune
```

### session

Session management commands for Claude Code integration. These commands support
multi-instance scenarios where multiple Claude Code sessions may be running.

#### session claim

Claim primary session ownership for this project. Reads session info from stdin.

```bash
echo '{"session_id": "abc", "source": "claude-code"}' | bits session claim
```

Output:
```json
{"claimed": true}
```

If another session already owns this project:
```json
{"claimed": false, "owner": "existing-session-id"}
```

#### session release

Release session ownership. Only the owner can release.

```bash
echo '{"session_id": "abc", "source": "claude-code"}' | bits session release
```

#### session prune

Manually remove a stale session file.

```bash
bits session prune
```

#### session hook

Stop hook with session ownership check. Only blocks if:
1. This session is the owner (session_id matches)
2. Drain mode is active
3. Tasks remain (active or open)

```bash
echo '{"session_id": "abc", "source": "claude-code"}' | bits session hook
```

### drain

Drain mode commands for working through all tasks before exiting.

#### drain claim

Activate drain mode. Uses the session owner from the session file.

```bash
bits drain claim
```

Output:
```json
{"success": true, "drain_active": true, "message": "Drain mode activated"}
```

Drain mode is automatically deactivated when the stop hook detects all tasks are complete.

#### drain release

Deactivate drain mode manually.

```bash
bits drain release
```

## Storage Format

Tasks are stored in `~/.bits/<sanitized-project-path>/`.

For example, if your project is at `/Users/alice/projects/myapp`, tasks are
stored in `~/.bits/Users-alice-projects-myapp/`.

Each task is a Markdown file with YAML frontmatter:

```markdown
---
id: abc123
title: Fix the login bug
status: open
priority: medium
created_at: 2025-01-19T10:30:00Z
depends_on:
  - xyz789
---

Users can't log in with email addresses containing a plus sign.
The `+` character is being URL-encoded incorrectly.
```

### Task Fields

| Field | Description |
|-------|-------------|
| `id` | 3-8 character identifier (auto-generated, grows to avoid collisions) |
| `title` | Short task title |
| `status` | `open`, `active`, or `closed` |
| `priority` | `critical`, `high`, `medium`, or `low` |
| `created_at` | RFC3339 timestamp |
| `closed_at` | RFC3339 timestamp (when closed) |
| `close_reason` | Why the task was closed |
| `depends_on` | List of task IDs this task depends on |

## Task Lifecycle

```
open ──claim──> active ──close──> closed
  ^               │
  └───release─────┘
```

- **open**: Task exists but no one is working on it
- **active**: Task is being worked on (only one allowed)
- **closed**: Task is complete

## Claude Code Integration

Configure bits hooks for session management and drain mode:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "bits session claim"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "bits session release"
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "bits session hook"
          }
        ]
      }
    ]
  }
}
```

The stop hook only blocks exit when:
1. This session is the primary owner
2. Drain mode is active (`bits drain claim` was called)
3. Tasks remain to be completed

When all tasks are complete, drain mode is automatically deactivated.

## Multi-Instance Support

bits supports multiple Claude Code instances working on the same project:

| Scenario | Behavior |
|----------|----------|
| First Claude starts | Claims session, becomes primary |
| Second Claude starts | Sees existing session, does nothing |
| Primary runs drain mode | Exit blocked until tasks complete |
| Secondary tries to exit | Always allowed (not primary) |
| Primary exits normally | Session released, file deleted |
| Stale session | Use `bits session prune` to clean up |

This enables workflows like:
- Planning instances that create tasks and exit freely
- Work instances that drain all tasks before exiting
- Multiple parallel read-only instances

## JSON Output

All commands support `--json` for machine-readable output:

```bash
bits list --json
```

```json
[
  {
    "id": "abc123",
    "title": "Fix the login bug",
    "status": "open",
    "priority": "medium",
    "created_at": "2025-01-19T10:30:00Z",
    "depends_on": ["xyz789"],
    "description": "Users can't log in with email addresses containing a plus sign."
  }
]
```

## License

MIT License. See [LICENSE](LICENSE) for details.
