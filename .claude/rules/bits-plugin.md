---
paths:
  - "plugins/bits/**/*"
---

# Bits Plugin Development

## Version Bumping - MANDATORY

**STOP. Before committing ANY change to this plugin, you MUST bump the version.**

This applies when modifying: `commands/`, `skills/`, `hooks/`, `.claude-plugin/`, or any plugin configuration.

### Step 1: Determine version increment

| Directory | Change | Bump |
|:----------|:-------|:-----|
| `skills/` | Add skill | MINOR |
| `skills/` | Modify skill | PATCH |
| `skills/` | Remove skill | PATCH (MAJOR if breaking) |
| `commands/` | Add command | MINOR |
| `commands/` | Modify command | PATCH |
| `hooks/` | Add hook | MINOR |
| `hooks/` | Modify hook behavior | PATCH |
| Any | Breaking change, major restructure | MAJOR |

### Step 2: Update version in plugin.json

```
.claude-plugin/plugin.json → "version": "X.Y.Z"
```

This is the single source of truth for the plugin version.

---

**Why this matters**: Plugin consumers need version numbers to track updates. Forgetting to bump creates confusion about what version contains which changes.

---

## Hooks Architecture (Session-Based)

The hooks use bits session management to track primary Claude instance ownership.
Only the primary instance (first to start) can be blocked during drain mode.

### Hook Events

| Event | Matcher | Command | Purpose |
|-------|---------|---------|---------|
| SessionStart | (any) | `bits session claim` | Claim primary session ownership |
| SessionStart | compact | `bits session compact` | Re-inject drain context after compaction |
| SessionEnd | (any) | `bits session release` | Release session ownership |
| Stop | (any) | `bits session hook` | Check drain mode and block if needed |

### Session Flow

1. First Claude instance starts → claims session via `bits session claim`
2. Secondary instances → see existing session, do nothing
3. Primary runs `/bits-drain` → `bits drain claim` sets drain_active=true
4. Primary tries to exit → blocked if drain_active AND tasks remain
5. Drain exceeds 12 hours → Stop hook force-releases drain and allows exit
6. Secondary tries to exit → allowed (not session owner)
7. Primary exits normally → `bits session release` deletes session file
8. Context compaction → `bits session compact` re-injects drain state

### Drain Release

- `bits drain release` — fails if active/open tasks remain
- `bits drain release --force` — suspends drain even with remaining tasks (for replan/escalation)

### Session File

Location: `~/.bits/<project>/session.json`

```json
{
  "session_id": "abc123",
  "started_at": "2025-01-21T10:00:00Z",
  "source": "claude-code",
  "drain_active": false,
  "drain_started_at": null
}
```
