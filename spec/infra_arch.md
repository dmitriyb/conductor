# Infra Module — Architecture

## 1. Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     infra package                       │
│                                                         │
│  ┌─────────────────────┐   ┌────────────────────────┐   │
│  │   CredentialStore    │   │       GitCloner        │   │
│  │   (interface)        │   │                        │   │
│  │  Get(ctx, name)      │   │  Clone(ctx, url, dst)  │   │
│  │      string, error   │   │  injects token from    │   │
│  └──┬──────┬──────┬─────┘   │  CredentialStore       │   │
│     │      │      │         └────────────┬───────────┘   │
│     ▼      ▼      ▼                      │               │
│  ┌─────┐┌─────┐┌─────┐                  │               │
│  │ rbw ││ env ││file │                  │               │
│  └─────┘└─────┘└─────┘                  │               │
│                                          │               │
│  ┌───────────────────────────────────────┘               │
│  │                                                       │
│  │  ┌──────────────────────────┐                         │
│  │  │     ImageBuilder         │                         │
│  │  │                          │                         │
│  │  │  Build(ctx, cfg) error   │                         │
│  │  │  generates Dockerfile    │                         │
│  │  │  if none specified,      │                         │
│  │  │  runs docker build       │                         │
│  │  └──────────────────────────┘                         │
│  │                                                       │
└──┼───────────────────────────────────────────────────────┘
   │
   ▼
 config.*Config
 (read-only input)
```

## 2. Package Layout

```
internal/
└── infra/
    ├── creds.go        CredentialStore interface + factory
    ├── creds_rbw.go    rbw backend
    ├── creds_env.go    env backend
    ├── creds_file.go   file backend
    ├── git.go          GitCloner
    └── docker.go       ImageBuilder
```

## 3. CredentialStore Design

The interface is intentionally minimal — a single `Get` method. Backend
selection happens once at startup via a factory function.

```
     NewCredentialStore(backend string)
                │
    ┌───────────┼───────────┐
    ▼           ▼           ▼
  rbwStore   envStore   fileStore

Each implements:
  Get(ctx, name) → (secret string, err error)
```

Backends are selected by the `credentials.backend` config field. The factory
returns an error for unknown values. This is validated at config time (FR3 in
config_reqs.md) but the factory also checks as defense-in-depth.

## 4. Git Clone Flow

```
config.Project.Repository
        │
        ▼
  parse URL format
  (git@ → https://)
        │
        ▼
  CredentialStore.Get("github_pat")
        │
        ▼
  inject token into URL
  https://<token>@github.com/owner/repo.git
        │
        ▼
  exec: git clone --quiet <url> <tmpdir>/repo
        │
        ▼
  strip token from remote:
  git remote set-url origin https://github.com/owner/repo.git
        │
        ▼
  return clone path
```

The temporary directory is created with `os.MkdirTemp`. On failure, the
directory is removed via `defer`.

## 5. Docker Build Flow

```
config.Docker
    │
    ├── Dockerfile specified?
    │     YES → use as-is
    │     NO  → generate from BaseImage (§6)
    │
    ▼
  exec: docker build
    --tag conductor-<project.name>
    --build-arg K=V (for each build_arg)
    -f <dockerfile>
    .
    │
    ▼
  return image name or error
```

## 6. Design Decisions

**D1 — CLI shelling vs SDK**
Git and Docker operations shell out to CLI tools (`git`, `docker`) rather
than using Go SDKs. Rationale: the Docker SDK adds significant dependency
weight; git operations are simple clones. Claude Code is installed via
shell anyway.

**D2 — Token in URL (not credential helper)**
Embedding the token in the clone URL is simpler than configuring a git
credential helper inside the container. The token is stripped from the
remote immediately after clone (FR7).

**D3 — Env file in /dev/shm**
Secrets are written to an env file in `/dev/shm` (RAM-backed tmpfs), never
to disk. The file is `chmod 600` and cleaned up on exit.

**D4 — No Docker layer caching logic**
The builder does not manage layer caching beyond Docker's built-in mechanism.
Users who need custom caching can provide their own Dockerfile.
