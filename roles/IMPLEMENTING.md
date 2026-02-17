# Implementing Role

Pick up an issue, write code and tests, submit PR.

## Context Loading

`CLAUDE.md` is automatically loaded by Claude Code at session start.

When implementing an issue, read ONLY these additional documents:

1. This file (`IMPLEMENTING.md`)
2. The GitHub issue you are working on
3. `spec/<module>_impl.md` - implementation details (data structures, algorithms)

The issue provides requirements and acceptance criteria. The impl file provides code patterns and technical details.

**DO NOT READ:**
- Any other files from `roles/`
- `spec/*_reqs.md`, `spec/*_arch.md`, `spec/*_plan.md` - already distilled into the issue during planning

## Starting an Issue

1. Read the issue fully
2. Understand the acceptance criteria before writing code

## Writing Code

- Code must trace to requirements listed in the issue
- Follow patterns in existing codebase
- No unrelated changes

## Writing Tests

Tests verify requirements, not just code coverage.

### Test Design

Every test answers: "Which requirement does this verify?"

1. Setup: Define initial state
2. Execute: Run the operation
3. Assert: Result satisfies requirement

### Test Naming

Reference requirement IDs in test names (e.g., `TestFR1_LoadConfig`, `TestNFR1_ValidationErrors`).

## Creating PR

Use template at `.github/pull_request_template.md`

- Link the issue (Closes #N)
- List requirements fulfilled
- Confirm tests verify requirements
- No unrelated changes
