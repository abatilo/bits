---
name: bits-drain
description: Start working on the next ready task
allowed-tools:
  - Bash(bits:*)
  - Bash(jq:*)
  - Read
---

# Bits-Drain: Checkpointed Autonomous Loop

An autonomous scheduler that claims one task at a time, implements it with validation, and tail-calls itself to continue. Uses bits as the single source of truth.

## Startup

Activate drain mode to block exit until work is complete:
```bash
bits drain claim
```

## Resume or Claim

Check for an already active task:
```bash
bits list --active --json | jq -r '.[0].id // empty'
```

**If a task is already active:** Resume it (skip to Execute).

**If no task is active:** Find the next ready task:
```bash
bits ready --json | jq -r '.[0].id // empty'
```

**If no ready tasks exist:** Check if any open (blocked) tasks remain:
```bash
bits list --open --json | jq -r 'length'
```
- If open tasks exist: they are blocked on unclosed dependencies. Inform the user and stop.
- If no tasks at all: `bits drain release` and stop.

**If a ready task exists:** Claim it:
```bash
bits claim <task_id>
```

## Execute

Read the task details:
```bash
bits show <task_id>
```

Implement the task fully. Use /commit for atomic commits.

## Validate

After implementation, validate the task before closing:

1. **Task-local validation**: Run any verification commands mentioned in the task description (lint, test, etc.). Check acceptance criteria from the description.
2. **Evidence required**: Do not close a task without concrete evidence it is satisfied — command output, file inspection, or test results.

## Terminal States

Every task must reach exactly one of these states:

### Close (task satisfied)
All acceptance criteria met with evidence:
```bash
bits close <task_id> "reason with evidence summary"
```

### Release with New Blockers (validation found gaps)
Implementation revealed required follow-up work:
1. Create new bits for each gap discovered
2. Add dependencies from the current task to those new bits (current task depends on new work)
3. Release the current task so new blockers can be claimed:
```bash
bits add "Fix discovered gap" -d "Description" --json
# Note the new task ID
bits dep <current_task_id> <new_task_id>
bits release <current_task_id>
```

If this is a root verification task (title starts with "Verify goal:"), track the round:
- Check existing tasks for `[goal:<this_task_id>]` prefix to find the current round number
- Prefix new blocker titles: `[goal:<this_task_id>][round:N] <description>`
- Round N = max existing round + 1 (start at 1)

### Escalate (autonomy should stop)
Create a Replan task and suspend drain when:
- The same root verification task has failed 3+ rounds without progress (same gaps keep appearing)
- The user request is ambiguous and you cannot determine what "done" means
- You are stuck in a loop or unsure how to proceed

```bash
bits add "Replan: <original goal summary>" -p critical -d "Description of what went wrong and why replanning is needed" --json
bits drain release --force
```

After escalation, do NOT invoke /bits-drain again. Inform the user.

## Continue the Loop

After any **Close** or **Release with New Blockers**, immediately invoke `/bits-drain` to process the next task.

This tail-call is mandatory. The only reasons NOT to re-invoke:
1. No open or ready tasks remain (all work complete)
2. A Replan task was created (escalation)
3. User input is genuinely required (ambiguous requirement)

```
/bits-drain
```

$ARGUMENTS
