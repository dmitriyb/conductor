# Planning Role

Decompose milestones into GitHub issues and handle specification changes.

## Context Loading

`CLAUDE.md` is automatically loaded by Claude Code at session start.

When planning a module, read ONLY these additional documents:

1. This file (`PLANNING.md`)
2. `spec/<module>_plan.md` - steps, dependencies, milestones
3. `spec/<module>_reqs.md` - requirements for issue traceability
4. `spec/<module>_arch.md` - architecture for understanding scope
5. `.github/ISSUE_TEMPLATE.md` - issue format

**DO NOT READ:**
- Any other files from `roles/`
- `spec/*_impl.md` - implementation details, needed only when coding
- Other module specs unless there's a cross-module dependency

## Decomposing a Milestone

1. Read milestone steps in `spec/*_plan.md`
2. For each step, estimate implementation scope
3. If >500 LOC, split into multiple issues
4. Each implementation issue includes its own unit tests — unit tests are not separate issues. They test the code being written (functions, structs, internal logic).
5. Create separate acceptance test issues that verify requirements (FR/NFR) are satisfied end-to-end. Acceptance tests are organized by requirement, not by implementation step.
6. Benchmarks and cross-platform builds remain separate issues (they are infrastructure, not feature implementation).
7. Present the full proposed issue list to the user for review and approval **before** creating any GitHub issues. Do not create issues until the user explicitly approves.
8. After approval, create issues using `gh issue create` (never the REST API). Use the template at `.github/ISSUE_TEMPLATE.md` for issue body format.

### Issue Creation Checklist

- [ ] Title references plan step (e.g., "[Step 1.3] Implement config validation")
- [ ] Spec references filled (requirement IDs, plan step)
- [ ] Acceptance criteria include unit tests for the code being implemented
- [ ] Scope estimate provided

### Acceptance Test Issues

- [ ] Title references requirement (e.g., "[FR4] Acceptance test: CLI subcommands")
- [ ] Tests verify the requirement is fulfilled, not individual functions
- [ ] Scope covers end-to-end behavior described in the requirement

## Handling Specification Changes

When requirements, architecture, or implementation details change:

1. Update the relevant `spec/*.md` file
2. Assess impact on open issues:
   - **No code written yet**: Close issue with comment explaining change
   - **PR in progress**: Close PR, close issue, create new issue if needed
   - **Completed work affected**: Create new issue to address the change
3. Update `*_plan.md` if step scope/dependencies changed

Keep it lightweight. Not every spec edit needs issue triage.
