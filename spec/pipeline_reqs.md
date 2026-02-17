# Pipeline Module — Requirements

## 1. Purpose

The pipeline module constructs a DAG from the step definitions in the config,
validates its structure (no cycles, all dependencies exist), executes steps in
topological order with maximum parallelism, evaluates CEL conditions to decide
whether to run conditional steps, and aggregates results from all steps into a
final pipeline report.

## 2. Functional Requirements

**FR1 — DAG Construction**
Build a directed acyclic graph from the `pipeline` section of the config. Each
step is a node. `depends_on` entries are directed edges. Steps with no
dependencies are roots (ready to execute immediately).

**FR2 — Cycle Detection**
Detect cycles in the DAG before execution begins. If a cycle is found, report
the cycle path (e.g., `A → B → C → A`) and return an error. Do not execute
any steps.

**FR3 — Topological Sort**
Compute a topological ordering of steps. Steps whose dependencies have all
completed are eligible for execution. The ordering must be deterministic for
the same input (sort by step name to break ties).

**FR4 — Parallel Execution**
Execute independent steps concurrently using goroutines. A step becomes
eligible when all its `depends_on` steps have completed successfully. A
semaphore or worker pool limits the maximum number of concurrent agents
(configurable, default: number of pipeline steps).

**FR5 — CEL Condition Evaluation**
Before executing a step, evaluate its `condition` field as a CEL expression.
The expression receives a `steps` variable containing the results of all
completed steps. If the condition evaluates to `false`, skip the step
(mark as `skipped`, not `failure`). An empty condition means always execute.

**FR6 — Step Result Aggregation**
After all steps complete (or skip), aggregate results into a `PipelineResult`
containing: list of step results (with status, output, log path), overall
status (`success` if all non-skipped steps succeeded, `failure` otherwise),
and wall-clock duration.

**FR7 — Failure Propagation**
If a step fails, all steps that depend on it (transitively) are marked as
`skipped` with a reason indicating the failed dependency. Independent
branches continue executing.

**FR8 — Pipeline Summary**
Print a human-readable summary at the end: each step's name, status,
duration, and key output fields. Print the final PR URL if available.

## 3. Non-Functional Requirements

**NFR1 — Deterministic Output**
For the same config and step results, the summary output must be identical
across runs. Step execution order may vary due to parallelism, but the
summary is always printed in config-defined order.

**NFR2 — Graceful Shutdown**
On receiving SIGINT or SIGTERM, cancel the context, wait for running agents
to finish (or timeout), and print partial results.

**NFR3 — No Shared Mutable State**
Step results are communicated through channels or a synchronized results map.
No global variables.

## 4. Interfaces

**Depends on:** config (`StepDef`, `Config`), agent (`RunAgent`, `StepResult`),
infra (`CloneResult`, `EnvFile`, image name).

**Consumed by:** CLI `run` command (top-level entry point).

**External dependencies:** `github.com/google/cel-go` (CEL evaluation),
standard library (`sync`, `context`, `time`).
