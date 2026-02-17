# Reviewing Role

Review a pull request for correctness, spec traceability, and test quality.

## Context Loading

`CLAUDE.md` is automatically loaded by Claude Code at session start.

When reviewing a PR, read ONLY these additional documents:

1. This file (`REVIEWING.md`)
2. The PR diff and description
3. The linked issue (from the PR body)
4. `spec/<module>_reqs.md` - to verify spec traceability
5. `spec/<module>_impl.md` - to verify patterns are followed

**DO NOT READ:**
- Any other files from `roles/`
- `spec/*_arch.md`, `spec/*_plan.md` - not needed for code review

## Review Process

1. Read the PR description and linked issue
2. Read the full diff
3. Check spec traceability: does the code implement what the issue requires?
4. Check test quality: do tests verify requirements, not just coverage?
5. Check code quality: correctness, error handling, existing patterns followed
6. Post inline comments for each issue found
7. Post a summary review

## Posting Comments

When submitting GitHub PR reviews on the user's own PRs, always use event `COMMENT` (not `APPROVE` or `REQUEST_CHANGES`), as GitHub disallows approving or requesting changes on your own PRs.

Write a JSON file and pass it via `--input`. Do NOT use `-f` flags for reviews with inline comments — the nested `comments` array cannot be constructed with `-f`.

```bash
cat > /tmp/review.json << 'EOF'
{
  "event": "COMMENT",
  "body": "Brief summary of review findings.",
  "comments": [
    {
      "path": "internal/config/validate.go",
      "line": 42,
      "body": "Short, explicit comment with code example if needed."
    }
  ]
}
EOF
gh api repos/{owner}/{repo}/pulls/{number}/reviews --method POST --input /tmp/review.json
```

- Each comment should be short, explicit, and aligned with the code
- Include a code example if the fix isn't obvious
- The summary comment should be brief and not duplicate inline comments

## What to Check

### Spec Traceability
- Code changes map to requirements listed in the issue
- No unrelated changes snuck in
- Test names reference requirement IDs

### Correctness
- Error paths handled (not just happy path)
- Edge cases considered
- No resource leaks (unclosed files, goroutine leaks)
- Context cancellation respected where applicable

### Patterns
- Follows conventions in existing codebase
- No global state introduced
- Errors wrapped with context (`fmt.Errorf("...: %w", err)`)

### Tests
- Tests verify requirements, not implementation details
- Failure cases tested, not just success
- No flaky patterns (timing-dependent, order-dependent)

## After Review

If changes requested: the fixer role will address each comment individually.
If approved: the PR is ready for merge.
