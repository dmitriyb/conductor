# Pipeline Module — Architecture

## 1. Component Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                      pipeline package                        │
│                                                              │
│  ┌──────────────┐   ┌──────────────┐   ┌─────────────────┐  │
│  │  DAG Builder  │   │  Executor    │   │  CEL Evaluator  │  │
│  │              │   │              │   │                 │  │
│  │  steps →     │   │  goroutines  │   │  condition +    │  │
│  │  adjacency   │   │  per step,   │   │  step results   │  │
│  │  list,       │   │  channel-    │   │  → bool         │  │
│  │  topo sort,  │   │  based       │   │                 │  │
│  │  cycle check │   │  coordination│   │                 │  │
│  └──────┬───────┘   └──────┬───────┘   └──────┬──────────┘  │
│         │                  │                  │              │
│         └──────────┬───────┘                  │              │
│                    │                          │              │
│                    ▼                          │              │
│         ┌──────────────────┐                  │              │
│         │  Execute(ctx,    │◄─────────────────┘              │
│         │    cfg, logger)  │                                 │
│         │                  │──► agent.RunAgent (per step)    │
│         └──────────┬───────┘                                 │
│                    │                                         │
│                    ▼                                         │
│            PipelineResult                                    │
│         ┌──────────────────┐                                 │
│         │  Results []      │                                 │
│         │  Status  string  │                                 │
│         │  Duration time   │                                 │
│         └──────────────────┘                                 │
└──────────────────────────────────────────────────────────────┘
```

## 2. Package Layout

```
internal/
└── pipeline/
    ├── dag.go         DAG construction, topo sort, cycle detection
    ├── executor.go    Parallel step execution, coordination
    ├── cel.go         CEL condition evaluation
    └── result.go      PipelineResult, summary printer
```

## 3. DAG Representation

```
Graph
├── Nodes    map[string]*Node
└── Order    []string            (topological order)

Node
├── Name      string
├── Step      config.StepDef
├── InDegree  int
├── Deps      []*Node            (edges: this depends on)
└── Dependents []*Node           (edges: these depend on this)
```

The graph is built from `config.Pipeline`. Each step name is unique
(enforced by config validation). `InDegree` tracks the number of
unresolved dependencies for scheduling.

## 4. Topological Sort Algorithm

Kahn's algorithm (BFS-based):

```
1. Initialize queue with all nodes where InDegree == 0
2. Sort queue alphabetically (determinism)
3. While queue is not empty:
   a. Pop node from queue
   b. Append to result order
   c. For each dependent:
      - Decrement InDegree
      - If InDegree == 0, add to queue (maintain sort)
4. If result length != node count → cycle detected
```

If a cycle is detected, a DFS finds and reports the exact cycle path.

## 5. Execution Architecture

```
Execute(ctx, cfg)
    │
    ├── Build DAG
    ├── Topo sort (+ cycle check)
    ├── Setup infra (clone repo, write env file, build image)
    │
    ├── Launch coordinator goroutine
    │     │
    │     ├── ready channel ◄── nodes with InDegree == 0
    │     │
    │     └── For each ready node:
    │           │
    │           └── spawn goroutine:
    │                 ├── evaluate CEL condition
    │                 │   ├── false → mark skipped
    │                 │   └── true  → agent.RunAgent(...)
    │                 │
    │                 ├── store StepResult
    │                 │
    │                 └── for each dependent:
    │                       decrement InDegree
    │                       if InDegree == 0 → send to ready channel
    │
    ├── Wait for all goroutines (sync.WaitGroup)
    │
    └── Aggregate results → PipelineResult
```

## 6. CEL Environment

Single variable `steps` of type `map<string, object>`. Each object has
`status: string` and `output: map[string]dyn`. Example expression:
`steps.review.output.status == "changes_requested"`. Expressions are
parsed at DAG build time and evaluated at step execution time.

## 7. Failure Propagation

BFS from failed node through `Dependents` edges. Each transitive dependent
is marked `skipped` with reason `"dependency X failed"`. Independent
branches continue executing.

## 8. Design Decisions

**D1 — Kahn's algorithm over DFS-based topo sort**
Kahn's algorithm naturally integrates with the BFS-based scheduler. Nodes
become ready when their InDegree reaches 0, which maps directly to the
scheduling condition.

**D2 — Goroutine per step, not worker pool**
Each step runs in its own goroutine. Since the number of pipeline steps is
small (typically < 10), the overhead is negligible. A semaphore can limit
concurrency if needed.

**D3 — CEL over custom expression language**
CEL (Common Expression Language) is a well-specified, sandboxed expression
language used by Kubernetes and other infrastructure tools. It supports
the exact operations needed (field access, comparisons, boolean logic)
without the risks of eval or template-based conditions.

**D4 — Channel-based coordination**
Steps signal readiness via a buffered channel. This avoids polling and
integrates naturally with Go's concurrency model.

**D5 — Results in a sync.Map**
Step results are stored in a `sync.Map` keyed by step name. This allows
concurrent reads (for CEL evaluation of downstream conditions) and writes
(when steps complete) without explicit locking.
