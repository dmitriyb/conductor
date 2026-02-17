---
name: review
description: Review a pull request following the reviewing role
disable-model-invocation: true
argument-hint: <pr-number>
---

Using the reviewing role (@roles/REVIEWING.md) and @~/.claude/skills/go-expert/SKILL.md, review PR #$ARGUMENTS. Each comment should be short, explicit, and aligned with the code (with a possible example). Summary comment should be short and not duplicate other comments.
