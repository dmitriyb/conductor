# Agent Module — Implementation Plan

## 1. Steps

| Step | Description | Requirement | Architecture | Implementation | Done |
|------|-------------|-------------|--------------|----------------|------|
| 3.1 | Define `StepResult` type | — | §1 Component Diagram | §1 StepResult Type | |
| 3.2 | Implement prompt template rendering (system + task) | FR1, NFR2 | §5 Template Data Model | §2 Template Rendering | |
| 3.3 | Implement init script generation | FR4 | §4 step 2, §7 D2 | §3 Init Script | |
| 3.4 | Implement container runner (`RunAgent`) | FR2, FR3, FR8, NFR1 | §4 Execution Flow | §4 Container Runner | |
| 3.5 | Implement output marker parser and schema validator | FR5, FR6 | §6 Output Protocol, §7 D3 | §5 Output Parser | |
| 3.6 | Implement log file writing | FR7 | §7 D5 | §4 (integrated in RunAgent) | |
| 3.7 | Write unit tests for template, parser, and runner | NFR3 | — | — | |

## 2. Dependency DAG

```
3.1 (StepResult)
 │
 ├──► 3.2 (template rendering)
 │       │
 │       └──► 3.4 (container runner) ──► 3.7 (tests)
 │              ▲           │              ▲
 │              │           │              │
 │       3.3 (init script)  └── 3.6 (log)  │
 │                                         │
 └──► 3.5 (output parser) ────────────────►┘
```

Steps 3.2, 3.3, and 3.5 can proceed in parallel after 3.1. Step 3.4
integrates all of them. Step 3.6 is part of 3.4. Step 3.7 runs last.

## 3. Milestones

**M1 — Template Rendering (steps 3.1–3.2)**
Prompt templates render correctly with `TemplateData`. System prompts load
from file paths. Task templates interpolate step outputs. The output contract
section is appended to system prompts.

**M2 — Container Runner (steps 3.3–3.4, 3.6)**
`RunAgent` starts a Docker container with the correct security flags,
mounts, and environment. Full output is captured and logged. Context
cancellation kills the container.

**M3 — Output Parsing (step 3.5)**
The parser finds the `###PIPELINE_OUTPUT###` marker, extracts JSON, validates
against the output schema, and returns a `StepResult`. Missing fields and
type mismatches produce clear errors.

## 4. Verification Criteria

- Rendering a task template with `{{.PRNumber}}` and `TemplateData{PRNumber: "42"}`
  produces `"42"` in the output.
- Rendering a template with `{{.Steps.implement.Output.pr_number}}` accesses
  prior step output correctly.
- Loading a system prompt from a file path reads the file contents.
- `ParseOutput` with a valid marker line returns a populated `StepResult`.
- `ParseOutput` with no marker returns an error containing "no marker found".
- `ParseOutput` with invalid JSON after the marker returns an error.
- Schema validation catches missing required fields.
- Schema validation catches type mismatches (string field with int value).
- `buildDockerArgs` includes `--cap-drop=ALL` and `-u 1000:1000`.
- Read-only workspace produces a mount with `:ro` suffix.
- Context cancellation during `RunAgent` returns an error (not a hang).

## 5. LOC Estimate

| Step | Estimated LOC |
|------|---------------|
| 3.1 | 15 |
| 3.2 | 80 |
| 3.3 | 30 |
| 3.4 | 120 |
| 3.5 | 80 |
| 3.6 | 15 |
| 3.7 | 210 (tests) |
| **Total** | **~550** |
