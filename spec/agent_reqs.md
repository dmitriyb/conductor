# Agent Module — Requirements

## 1. Purpose

The agent module manages the lifecycle of a single Claude Code agent running
inside a Docker container. It renders prompt templates, starts the container
with the correct mounts and environment, captures output, parses structured
results via output markers, and validates the JSON payload against the agent's
declared output schema.

## 2. Functional Requirements

**FR1 — Prompt Template Rendering**
Render agent prompts using Go's `text/template`. The system prompt is loaded
from a file path (relative to the repo root) or used inline. The task prompt
is a Go template that receives a data context with fields like `IssueNumber`,
`RepoURL`, `PRNumber`, `RepoOwner`, `RepoName`, and outputs from prior steps
(accessed as `Steps.<name>.Output.<field>`).

**FR2 — Container Lifecycle**
Start a Docker container for each agent execution using the Docker CLI:
- Set the image to the one built by infra.
- Drop all capabilities (`--cap-drop=ALL`).
- Set `--security-opt=no-new-privileges`.
- Run as UID 1000 (`-u 1000:1000`).
- Pass the env file via `--env-file`.
- Mount workspace, prompts, and optional skills directory.
- Capture combined stdout+stderr.
- Return exit code and full output.

**FR3 — Mount Layout**
Configure container mounts per the agent protocol:
- `/workspace` — the cloned repo (read-write or read-only per agent config).
- `/tmp/orchestrator-prompts` — rendered prompt files (read-only).
- `/home/agent/.claude/skills` — host skills directory (read-only, optional).
- SSH agent socket forwarded if available.

**FR4 — Init Script Generation**
Generate a bash init script that runs inside the container before Claude:
- Configure git identity (name, email from agent config or defaults).
- Authenticate `gh` CLI using the scoped PAT from the env file.
- Configure SSH commit signing if an SSH agent socket is available.
- `cd /workspace` and `exec claude -p --dangerously-skip-permissions ...`
  reading system and task prompts from the mounted prompt files.

**FR5 — Output Marker Parsing**
Parse the agent's stdout for the pipeline output marker (`###PIPELINE_OUTPUT###`).
Extract the JSON payload from the last line matching the marker. Return an error
if no marker is found.

**FR6 — JSON Output Validation**
Validate the extracted JSON against the agent's `output_schema` from config.
Check that all required keys are present and have the expected types (`string`,
`int`). Return a structured `StepResult` containing the parsed fields.

**FR7 — Logging and Log Files**
Write the full agent output to a log file at `<log_dir>/conductor-<role>-<timestamp>.log`.
Log agent start, completion, and exit code via `slog`.

**FR8 — Timeout and Cancellation**
Accept a `context.Context`. If the context is cancelled or its deadline is
exceeded, kill the container (`docker kill`) and return an error.

## 3. Non-Functional Requirements

**NFR1 — No Direct Docker SDK**
Use `os/exec` to call the `docker` CLI. This avoids a large SDK dependency
and matches the existing approach.

**NFR2 — Template Safety**
Template rendering must not panic. Use `template.Must` only at init time.
Runtime rendering must return errors, not panic.

**NFR3 — Deterministic Prompt Files**
Prompt files are written to a temp directory, mounted read-only into the
container, and cleaned up after the container exits. They must not persist
across agent runs.

## 4. Interfaces

**Depends on:** config (`AgentDef`, `PromptDef`), infra (`CloneResult`,
`EnvFile`, image name).

**Consumed by:** pipeline (calls `RunAgent` for each step).

**Produces:** `StepResult` containing status, parsed output fields, log path.
