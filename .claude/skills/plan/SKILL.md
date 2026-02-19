---
name: plan
description: Decompose a milestone into GitHub issues
disable-model-invocation: true
argument-hint: <module-name>
---

Plan implementation of the $ARGUMENTS module.

## Context Loading

Read ONLY these documents:

1. `spec/$ARGUMENTS_plan.md` - steps, dependencies, milestones
2. `spec/$ARGUMENTS_reqs.md` - requirements for issue traceability
3. `spec/$ARGUMENTS_arch.md` - architecture for understanding scope
4. `.github/ISSUE_TEMPLATE.md` - issue format

**DO NOT READ** `spec/*_impl.md` — implementation details, needed only when coding.

## Decomposing a Milestone

1. Read milestone steps in the plan
2. For each step, estimate implementation scope
3. If >500 LOC, split into multiple issues
4. Each implementation issue includes its own unit tests — unit tests are not separate issues
5. Create separate acceptance test issues that verify requirements (FR/NFR) end-to-end, organized by requirement not by implementation step
6. Benchmarks and cross-platform builds remain separate issues
7. Present the full proposed issue list to the user for review and approval **before** creating any GitHub issues
8. After approval, create issues using `gh issue create` (never the REST API). Use `.github/ISSUE_TEMPLATE.md` for issue body format.

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
