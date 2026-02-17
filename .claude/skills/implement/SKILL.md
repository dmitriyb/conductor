---
name: implement
description: Implement a GitHub issue following the implementing role
disable-model-invocation: true
argument-hint: <issue-number>
---

First run `git checkout main && git pull --rebase` to ensure you are on the latest main. Then, using the implementing role (@roles/IMPLEMENTING.md) and @~/.claude/skills/go-expert/SKILL.md, implement issue #$ARGUMENTS and create a PR.
