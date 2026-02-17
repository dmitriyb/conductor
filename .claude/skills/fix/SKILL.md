---
name: fix
description: Fix review comments on a pull request
disable-model-invocation: true
argument-hint: <pr-number>
---

Read PR #$ARGUMENTS comments, fix each comment and provide a concise short response like "Fixed" or "Addressed", or answer in more detail if it is a question. Reply to EACH comment individually on GitHub (using the GitHub API to reply to each review comment thread). Do NOT post a single bulk comment summarizing all changes.
