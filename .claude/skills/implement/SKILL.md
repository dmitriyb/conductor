---
name: implement
description: Implement a GitHub issue — write code, tests, and create a PR
disable-model-invocation: true
argument-hint: <issue-number>
---

First run `git checkout main && git pull --rebase` to ensure you are on the latest main.

Implement issue #$ARGUMENTS. Use @~/.claude/skills/go-expert/SKILL.md for Go-specific guidance.

## Context Loading

Read ONLY these documents:

1. The GitHub issue you are working on
2. `spec/<module>_impl.md` - implementation details for the relevant module

**DO NOT READ** `spec/*_reqs.md`, `spec/*_arch.md`, `spec/*_plan.md` — already distilled into the issue.

## Workflow

1. Read the issue fully. Understand acceptance criteria before writing code.
2. Read the relevant `spec/*_impl.md` for technical details and patterns.
3. Create a feature branch.
4. Write code that traces to requirements listed in the issue.
5. Follow patterns in existing codebase. No unrelated changes.
6. Write tests that verify requirements, not just coverage.
7. Name tests with requirement IDs (e.g., `TestFR1_LoadConfig`, `TestNFR1_ValidationErrors`).
8. Run `go test ./...` to confirm all tests pass.
9. Commit and push.
10. Create a PR using `.github/pull_request_template.md`. Link the issue with "Closes #N".
