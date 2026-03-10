---
name: bits-plan
description: Analyze conversation history and create bits tasks with dependencies. Use when extracting work from a discussion, turning decisions into tracked tasks, or breaking down what was discussed into actionable items. Triggers on "create tasks from this", "turn this into bits", "extract tasks", "what should we track", "break this down into tasks", "plan from this conversation". Uses collaborative Codex debate to refine scope, discovers verification commands, creates self-contained tasks with context and acceptance criteria.
argument-hint: "[optional focus area]"
---

# Bits Planning Skill

Review the conversation history above to identify work that needs planning. Extract requirements, decisions, and context discussed—these inform the bits tasks you create. If the user provided additional instructions below, incorporate those as well.

## When to Use This Skill

| Scenario | Approach |
|----------|----------|
| Complex multi-step implementation | Use this skill for structured planning |
| Breaking down large features into tasks | Use this skill with debate refinement |
| Simple single-task work | Use bits skill directly |
| Quick research or exploration | Use Explore agent |

This is a three-phase process: discovery, planning with collaborative debate, then task creation with a mandatory goal contract.

## Phase 1: Discovery

Gather context from the conversation history and find verification commands.

### Step 1: Verification Commands
Run a focused Explore query to discover development commands. This may return nothing, especially for new projects—in which case, simply create tasks without verification sections.
```
Find the ACTUAL commands used in this project for verification. Search in order:
1. mise.toml / .mise.toml (mise task runner - https://github.com/jdx/mise)
2. package.json scripts / pyproject.toml / Makefile / Justfile
3. .github/workflows (CI jobs are authoritative)
4. docs/CONTRIBUTING.md or README.md

For each category, report the EXACT command string:
- Linting/formatting (e.g., `mise run lint`, `go fmt ./...`)
- Static analysis / type checking (e.g., `mise run check`, `staticcheck ./...`, `golangci-lint run`)
- Unit tests (e.g., `mise run test`, `go test ./...`)
- Scoped E2E tests - run specific tests (e.g., `mise run test:e2e -- -run TestAuth`, `go test ./e2e/... -run TestAuth`)
- Full E2E tests - run entire suite (e.g., `mise run test:e2e`, `go test ./e2e/...`)

Output format: "CATEGORY: [exact command]"
Stop searching a category once you find an authoritative source.
```

### Step 2: Discovery Synthesis
Consolidate findings from conversation history into planning input:
- **Architecture overview**: Patterns, conventions, and constraints discussed
- **Testing setup**: Where tests live, how to run them, what coverage exists
- **Verification commands**: From Step 1
- **Known risks**: Edge cases and caveats identified

This synthesis becomes the input for Phase 2.

## Phase 2: Planning with Collaborative Debate

Use multi-round refinement for thorough planning.

### Guiding Principles: Speed-of-Light Implementation

**Treat planning as a minimization problem.** The goal is not to design a comprehensive solution—it's to find the smallest, fastest path to the desired outcome.

- **Minimize changes**: What is the absolute minimum number of lines, files, and touch points needed? Every additional change is a potential bug, a review burden, and merge conflict risk.
- **Minimize complexity**: Prefer boring, obvious solutions over clever ones. If two approaches work, choose the one a junior developer could understand in 5 minutes.
- **Minimize scope**: Ruthlessly cut anything that isn't strictly required. "Nice to have" belongs in a separate future task, not this plan.
- **Minimize risk**: Favor incremental changes over big-bang rewrites. Ship something small that works over something ambitious that might not.

**Ask at every decision point**: "Is there a simpler way?" If the answer is yes, take it.

### Step 1: Initial Plan
Use the Plan subagent with **model: "opus"** to design the minimum viable implementation based on discovery synthesis. The plan should answer: "What is the smallest change that achieves the goal?"

### Step 2: Interview Codex

Use threaded conversations to probe and refine the plan. Claude interviews Codex, asking probing questions to surface gaps, risks, and simplification opportunities.

**Start the thread:**
Use `mcp__codex__codex` to share the initial plan and begin the interview. Inform Codex of your capabilities so it can request specific research:
```
prompt: "I'm planning this implementation: [plan].

I can run multiple parallel sub-agents to gather information:
- Explore agents to search the codebase and answer specific questions
- Plan agents for deeper architectural analysis

Help me think through this critically. What concerns do you have? What might I be missing? Are there specific questions I should investigate with an Explore agent before proceeding?"
```

**Probe deeper:**
Continue with `mcp__codex__codex-reply` using the returned `threadId`. Ask about literally anything: technical implementation, concerns, tradeoffs, edge cases, assumptions, risks, dependencies.

Questions should not be obvious—probe deeper into things that might not have been considered:
- "What's the hardest part of this?"
- "Where could this break in production?"
- "What assumptions am I making that might be wrong?"
- "Is there a simpler way to achieve this?"
- "What would you cut if you had to ship this in half the time?"

Challenge assumptions. Ask about the hard parts. Push back when answers feel incomplete.

**Continue until the plan is fully fleshed out:**
There's no fixed number of rounds. Keep interviewing until:
- Major concerns have been addressed or explicitly deferred
- The plan feels minimal and well-understood
- You're confident it represents the smallest viable implementation

**Synthesize insights:**
After the interview, integrate Codex's feedback into the final plan. Document what was deferred and why.

### Step 3: Draft the Goal Contract

Before creating any tasks, translate the user's request into a **Goal Contract**: 3-7 observable truths that must ALL be proven true for the work to be considered done.

**If the user's request is ambiguous** and you cannot write concrete, verifiable contract lines, STOP and ask the user for clarification. Do not create a Replan task — ambiguity requires user input.

The goal contract must be:
- **Observable**: Each line can be verified by running a command, inspecting a file, or testing behavior
- **Concrete**: No vague qualifiers like "properly handles" or "works correctly" — state what specifically must be true
- **Complete**: If all lines are true, the original request is satisfied

Example for "add auth to the API":
1. `POST /login` returns a JWT for valid credentials and 401 for invalid
2. Protected endpoints return 403 when no token is provided
3. Protected endpoints succeed with a valid token
4. Token expiration is enforced (expired tokens return 401)
5. Tests exist covering login success, login failure, protected access, and token expiration
6. `mise run lint` and `mise run test` pass

### Quality Gate
Before creating tasks, confirm:
- Goal contract lines are specific and verifiable
- All discovered edge cases addressed or explicitly deferred with rationale
- Error paths defined (what happens when X fails?)
- Testing strategy covers new code
- Trade-offs documented with reasoning

## Phase 3: Create Tasks

### Step 1: Create the Root Verification Task FIRST

This is mandatory. Create it before any implementation tasks so its ID is available for dependency wiring.

```bash
bits add "Verify goal: <concise goal summary>" -p high -d "<goal contract description>" --json
```

The description must contain four sections:

```markdown
# Original Request
<The user's original request, verbatim or faithfully paraphrased>

# Goal Contract
Each line must be proven true to close this task:
1. <observable truth 1>
2. <observable truth 2>
...

# Allowed Evidence
For each contract line, the verifier must cite one of:
- A closed child verification task that proves the line
- Command output demonstrating the behavior
- Specific file content or test results

Vibe-checking is not evidence. Each line needs explicit proof.

# Failure Handling
- If a contract line is false: create a new blocker bit, add it as a dependency, release this task
- If a contract line is ambiguous: create a Replan task and stop drain
- Track rounds via [goal:<this_task_id>][round:N] prefix on spawned blockers
- Escalate after 3 failed rounds without progress
```

Note the returned task ID — this is the **root verifier ID** used throughout.

### Step 2: Create Implementation Tasks

Create bits tasks using the bits skill. **Tasks must be self-contained** with sufficient context for immediate implementation without repeating discovery work. Be verbose and repeat context; prioritize completeness over brevity.

#### Task Description Structure

```markdown
# Context
Explain why this task exists and what it solves:
- The specific problem being addressed
- Why this solution was chosen over alternatives
- Relevant constraints or assumptions

# References
List all files and resources to consult during implementation:
- `path/to/relevant/file.ts` - reason it's relevant
- `path/to/example/pattern.ts:42-58` - specific pattern to follow
- [Link or doc reference] - what to learn

# Code Snippets
Include code from discovery only if it shows patterns or templates to replicate:

\`\`\`language
// Existing code to modify or pattern to follow
\`\`\`

# File Changes
Specify every file that will be touched:

Files to **edit**:
- `path/to/file.ts` - specific changes required

Files to **create**:
- `path/to/new/file.ts` - purpose and responsibility

Files to **delete**:
- `path/to/obsolete/file.ts` - reason for removal

# Acceptance Criteria
Concrete criteria that must be true when this task is done:
- Criterion 1
- Criterion 2

# Verification Commands
Commands to run after implementation:
- `[lint command]`
- `[test command]`
```

**Section inclusion rules:**
- Always include: Context, References, File Changes, Acceptance Criteria
- Include Code Snippets only if discovery surfaced code to replicate or templates to follow
- Include Verification Commands only if the project has test or lint commands
- Include Files to delete only if the task involves deletion or refactoring
- Replace all bracketed examples with concrete values

#### Task Requirements
1. Scope each task to complete in one focused session
2. Use specific language; avoid vague descriptions or qualifiers
3. When implementation reveals new tasks or scope changes, create separate bits tasks for each discovery instead of expanding this task

### Step 3: Create Verification Tasks

For each verification command discovered in Phase 1, and for groups of related acceptance criteria, create verification tasks:

```bash
bits add "Verify: <what is being verified>" -d "<description with exact commands to run>" --json
```

These act as intermediate verification between implementation and the root verifier.

### Step 4: Wire Dependencies

The dependency structure should be layered:

```
Implementation tasks → Verification tasks → Root verifier
```

1. Verification tasks depend on the implementation tasks they validate
2. The root verifier depends on ALL verification tasks
3. If a full E2E/integration test exists, create a verification task for it — it's one more blocker of the root verifier, not the root verifier itself

```bash
# Verification tasks depend on implementation
bits dep <verify_task_id> <impl_task_1_id>
bits dep <verify_task_id> <impl_task_2_id>

# Root verifier depends on verification tasks
bits dep <root_verifier_id> <verify_task_1_id>
bits dep <root_verifier_id> <verify_task_2_id>
bits dep <root_verifier_id> <e2e_verify_task_id>
```

### Step 5: CLAUDE.md Update Task

Create a documentation maintenance task that depends on the root verifier:

1. **Create the task**:
   - Title: "Update CLAUDE.md documentation"
   - Description: Review git commits from the last 4 days. Update CLAUDE.md files:
     (1) Add documentation for new patterns
     (2) Fix stale references
     (3) Create CLAUDE.md in directories lacking documentation
     Delete redundant or low-signal sections. Use the Explore subagent for thorough discovery. Use /commit for atomic commits.

2. **Set up dependencies**:
   ```bash
   bits add "Update CLAUDE.md documentation"
   bits dep <claude_md_task_id> <root_verifier_id>
   ```

## Handling Failures

When discovery or planning reveals blocking issues:
1. Create a P0 meta task titled: "Create plan for [blocker-topic]"
2. Description must include:
   - What was blocking and why it matters
   - Instruction to use Explore subagent for discovery
   - Instruction to use Plan subagent to design fix
   - Instruction to create implementation bits tasks via bits skill
3. Any implementation tasks spawned from meta tasks are also P0

$ARGUMENTS
