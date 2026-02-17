# Config Module — Requirements

## 1. Purpose

The config module defines the schema for `orchestrator.yaml`, loads and validates
configuration files, provides CLI entry points (`run`, `validate`, `build`), and
initializes structured logging via `slog`. It is the foundation all other modules
depend on — no module may import anything outside `config` without first receiving
a validated `Config` value.

## 2. Functional Requirements

**FR1 — YAML Schema Definition**
Define Go structs that map 1:1 to orchestrator.yaml sections: `project`, `credentials`,
`docker`, `agents` (map of agent definitions), and `pipeline` (list of steps with
`depends_on` and `condition` fields). All fields use `yaml` struct tags.

**FR2 — Configuration Loading**
Read orchestrator.yaml from a file path (defaulting to `./orchestrator.yaml`).
Unmarshal YAML into the struct tree. Return a typed `*Config` or a wrapped error
with file path and line context.

**FR3 — Configuration Validation**
After loading, validate the config: required fields present, credential backend
is one of `rbw | env | file`, each pipeline step references a defined agent name,
`depends_on` references exist, Docker base image is non-empty. Collect all errors
and return them as a single multi-error.

**FR4 — CLI Commands**
Expose three subcommands via a thin CLI layer (cobra or bare `os.Args`):
- `conductor run [--config path]` — load, validate, execute pipeline.
- `conductor validate [--config path]` — load, validate, print result, exit.
- `conductor build [--config path]` — load, build Docker image, exit.

**FR5 — Structured Logging**
Initialize `slog.Logger` with JSON handler for non-TTY and text handler for TTY.
Set level via `--log-level` flag (default `info`). Attach `component` attribute
to each module's logger (e.g., `component=config`).

**FR6 — Config Access Pattern**
Provide the validated `*Config` as a plain value — no global state, no singleton.
Callers receive it from `Load()` and pass it explicitly.

## 3. Non-Functional Requirements

**NFR1 — Error Quality**
Every validation error must include the field path (e.g., `pipeline[2].agent`)
and a human-readable reason. Errors are joinable via `errors.Join`.

**NFR2 — No External State**
Loading and validation are pure functions of the file contents. No network calls,
no environment variable reads (except for credential backends at runtime).

**NFR3 — Parse Performance**
Config files under 1 000 lines must parse and validate in < 50 ms.

## 4. Interfaces

**Consumed by:** infra (credential + docker config), agent (agent definitions),
pipeline (step definitions, conditions).

**Produces:** `*Config` struct tree, `*slog.Logger`.

**External dependencies:** `gopkg.in/yaml.v3`, standard library (`slog`, `os`, `errors`).
