# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Conductor is a configuration-driven CLI tool that orchestrates Claude Code agents in Docker containers. Given an `orchestrator.yaml`, it provisions infrastructure, renders prompts, runs agents, and coordinates them as a DAG pipeline with conditions and parallelism.

## Module Hierarchy

| Module | Purpose | Depends On |
|--------|---------|------------|
| config | YAML schema, loading, validation, CLI, structured logging | — |
| infra | Credential backends (rbw, env, file), git clone, Docker build | config |
| agent | Container lifecycle, prompt rendering, output parsing | config, infra |
| pipeline | DAG construction, topological sort, CEL conditions, parallel execution | config, agent |

## Technical Constraints

- **Go standard library first**: Only external deps are `yaml.v3`, `x/term`, and `cel-go`
- **CLI shelling**: Git, Docker, and rbw operations use `os/exec`, not SDK bindings
- **No global state**: Config passed explicitly, loggers created per component
- **Secret hygiene**: Credentials never written to disk, only to `/dev/shm` (RAM-backed tmpfs)

## Build & Test

- Build and test: `go test ./...`
- Vet: `go vet ./...`

## Git Conventions

- Default branch is `main` (never `master`)
- Always `git fetch origin` before creating a new branch
- Always branch from `origin/main`, not from the current branch

## Organizational Constraints

- **Module dependency order**: config → infra → agent → pipeline, implemented in that order
- **Spec traceability**: All code must trace back to requirements in spec/ documents
- **Requirement-driven testing**: Tests verify requirements are fulfilled, not just for coverage

## Where to Find Details

- **Skills**: `.claude/skills/` — `/implement`, `/review`, `/fix`, `/plan` commands
- **Requirements**: `spec/*_reqs.md` — functional and non-functional requirements per module
- **Architecture**: `spec/*_arch.md` — design decisions, data flow, component interactions
- **Implementation details**: `spec/*_impl.md` — data structures, algorithms, code examples
- **Implementation plans**: `spec/*_plan.md` — numbered steps, dependencies, milestones
