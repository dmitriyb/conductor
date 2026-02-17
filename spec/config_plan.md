# Config Module — Implementation Plan

## 1. Steps

| Step | Description | Requirement | Architecture | Implementation | Done |
|------|-------------|-------------|--------------|----------------|------|
| 1.1 | Define Config struct tree in `types.go` | FR1 | §3 Data Model | §1 Struct Definitions | |
| 1.2 | Implement `Load()` — read YAML, unmarshal to `*Config` | FR2 | §4 Data Flow | §2 Loading | |
| 1.3 | Implement `Validate()` — field checks, cross-references, multi-error | FR3, NFR1 | §5 D2, D6 | §3 Validation | |
| 1.4 | Implement `InitLogging()` — slog setup with TTY detection | FR5 | §5 D3 | §4 Logging Initialization | |
| 1.5 | Wire CLI entry point — flag parsing, subcommand dispatch | FR4, FR6 | §1 Component Diagram | §5 CLI Dispatch | |
| 1.6 | Write unit tests for Load, Validate, and InitLogging | NFR2, NFR3 | — | — | |

## 2. Dependency DAG

```
1.1 (types)
 │
 ├──► 1.2 (load)
 │       │
 │       ├──► 1.3 (validate) ──► 1.6 (tests)
 │       │                         ▲
 │       └─────────────────────────┘
 │
 └──► 1.4 (logging)
         │
         └──► 1.5 (CLI) ──► 1.6 (tests)
```

Steps 1.2 and 1.4 can proceed in parallel after 1.1. Step 1.5 needs both
1.3 and 1.4. Step 1.6 runs last.

## 3. Milestones

**M1 — Types + Loading (steps 1.1–1.2)**
A `Config` value can be populated from a YAML file. Tests confirm round-trip
fidelity with the example orchestrator.yaml.

**M2 — Validation (step 1.3)**
`conductor validate` prints all errors for a malformed config and exits 0
for a valid config.

**M3 — CLI wired (steps 1.4–1.5)**
`conductor validate orchestrator.yaml` works end-to-end. Logging output
switches between text (TTY) and JSON (pipe).

## 4. Verification Criteria

- `go vet ./...` and `go test ./...` pass.
- Loading the example YAML from config_impl.md §6 returns a `*Config` with
  all fields populated.
- Validating a config with a missing `project.name` returns an error
  containing the string `project.name: required`.
- Validating a config where a pipeline step references a non-existent agent
  returns an error containing `references undefined agent`.
- `conductor validate --config /dev/null` exits non-zero.
- Logging to a non-TTY produces JSON lines; logging to a TTY produces
  key=value text.

## 5. LOC Estimate

| Step | Estimated LOC |
|------|---------------|
| 1.1 | 80 |
| 1.2 | 30 |
| 1.3 | 100 |
| 1.4 | 30 |
| 1.5 | 50 |
| 1.6 | 110 (tests) |
| **Total** | **~400** |
