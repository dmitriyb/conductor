# Pipeline Module — Implementation Plan

## 1. Steps

| Step | Description | Requirement | Architecture | Implementation | Done |
|------|-------------|-------------|--------------|----------------|------|
| 4.1 | Define DAG types (`Graph`, `Node`) and build graph from config | FR1 | §3 DAG Representation | §1, §2 DAG Types + Construction | |
| 4.2 | Implement topological sort (Kahn's algorithm) with cycle detection | FR2, FR3 | §4 Topo Sort Algorithm | §3 Topo Sort + Cycle Detection | |
| 4.3 | Define `PipelineResult` type and summary printer | FR6, FR8, NFR1 | §1 Component Diagram | §4 Pipeline Result | |
| 4.4 | Implement CEL condition evaluator | FR5 | §6 CEL Environment | §5 CEL Evaluator | |
| 4.5 | Implement parallel executor with channel-based coordination | FR4, FR7, NFR2, NFR3 | §5 Execution Architecture | §6, §7 Executor + Failure Propagation | |
| 4.6 | Wire `Execute` into CLI `run` command | — | — | — | |
| 4.7 | Write unit tests for DAG, topo sort, CEL, and executor | — | — | — | |

## 2. Dependency DAG

```
4.1 (DAG types) ──► 4.2 (topo sort) ──► 4.5 (executor) ──► 4.6 (CLI wire)
                                           ▲                    │
4.3 (result type) ─────────────────────────┘                    │
                                           ▲                    ▼
4.4 (CEL evaluator) ──────────────────────►┘               4.7 (tests)
```

Steps 4.1, 4.3, and 4.4 can proceed in parallel. Step 4.2 depends on 4.1.
Step 4.5 depends on 4.2, 4.3, and 4.4. Steps 4.6 and 4.7 run last.

## 3. Milestones

**M1 — DAG + Topo Sort (steps 4.1–4.2)**
A graph can be built from config steps. Topological sort produces a
deterministic ordering. Cycles are detected and reported with the cycle
path. Unit tests cover linear chains, fan-out, fan-in, and cyclic graphs.

**M2 — CEL Evaluator (step 4.4)**
CEL expressions evaluate against a steps result map. `true`/`false`
conditions work. Invalid expressions produce clear errors. Tests cover
field access (`steps.review.output.status`), comparisons, and boolean
logic.

**M3 — Parallel Executor (step 4.5)**
The executor runs steps in parallel. Independent branches execute
concurrently. Failed steps propagate to dependents. Skipped steps
(condition = false) unblock dependents correctly. The end-to-end flow
works with the existing orchestrator's implement → review → fix pattern.

**M4 — Integration (steps 4.6–4.7)**
`conductor run` executes a full pipeline from config. The summary
prints correctly. SIGINT triggers graceful shutdown.

## 4. Verification Criteria

- `BuildGraph` with steps A→B→C produces a graph with 3 nodes and correct edges.
- `topoSort` of A→B→C returns `["A", "B", "C"]`.
- `topoSort` of {A→B, A→C} returns `["A", "B", "C"]` (alphabetical tie-break).
- `topoSort` of A→B→A returns an error containing "cycle" and the path "A → B → A".
- `EvalCondition("steps.review.output.status == 'approved'", ...)` returns `true`
  when the review step has `status: "approved"` in its output.
- `EvalCondition("", ...)` returns `true` (empty condition = always run).
- `EvalCondition` with an invalid expression returns an error.
- When step A fails, its dependents B and C are marked "skipped".
- When step A fails, independent step D (no dependency on A) still executes.
- Pipeline with 3 independent steps executes all 3 concurrently (observable
  via overlapping timestamps in logs).
- `PipelineResult.Print` output is identical for the same inputs across runs.
- SIGINT during execution cancels running agents and prints partial results.

## 5. LOC Estimate

| Step | Estimated LOC |
|------|---------------|
| 4.1 | 50 |
| 4.2 | 80 |
| 4.3 | 40 |
| 4.4 | 60 |
| 4.5 | 150 |
| 4.6 | 30 |
| 4.7 | 190 (tests) |
| **Total** | **~600** |

Note: slightly over original 500 estimate due to CEL integration and
failure propagation logic. All individual steps remain under 500 LOC.
