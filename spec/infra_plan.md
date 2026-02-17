# Infra Module — Implementation Plan

## 1. Steps

| Step | Description | Requirement | Architecture | Implementation | Done |
|------|-------------|-------------|--------------|----------------|------|
| 2.1 | Define `CredentialStore` interface and factory | FR1, FR5 | §3 CredentialStore Design | §1 Interface, §2 Backends | |
| 2.2 | Implement rbw backend | FR2 | §3 | §2 rbw | |
| 2.3 | Implement env and file backends | FR3, FR4 | §3 | §2 env, file | |
| 2.4 | Implement env file writer (secrets to `/dev/shm`) | NFR1 | §7 D3 | §3 Env File Writer | |
| 2.5 | Implement git cloner with token injection and cleanup | FR6, FR7, NFR2 | §4 Git Clone Flow | §4 Git Cloner | |
| 2.6 | Implement Docker image builder and Dockerfile generation | FR8, FR9 | §5, §6 Build Flow | §5, §6 Docker | |
| 2.7 | Write unit tests for credential backends, cloner, builder | NFR3 | — | — | |

## 2. Dependency DAG

```
2.1 (interface)
 │
 ├──► 2.2 (rbw)
 │
 ├──► 2.3 (env, file)
 │
 └──► 2.4 (env file writer)
         │
         └──► 2.5 (git clone) ──► 2.7 (tests)
                                    ▲
2.6 (docker) ───────────────────────┘
```

Steps 2.2, 2.3, 2.4, and 2.6 can all proceed in parallel after 2.1.
Step 2.5 depends on 2.4 (needs env file for token). Step 2.7 runs last.

## 3. Milestones

**M1 — Credential Backends (steps 2.1–2.4)**
All three credential backends work. The env file writer creates a file in
`/dev/shm` with correct permissions. Unit tests cover each backend with
mocks (rbw tested via fake exec).

**M2 — Git Clone (step 2.5)**
`Clone()` clones a repo with token injection and strips the token from the
remote afterward. Integration test with a local git repo.

**M3 — Docker Build (step 2.6)**
`BuildImage()` generates a Dockerfile from base image config and builds it.
The generated Dockerfile matches the structure from the existing orchestrator.

## 4. Verification Criteria

- `envStore.Get` returns the value of a set env var and errors on unset.
- `fileStore.Get` reads a file and trims trailing newline.
- `rbwStore.Get` shells out to `rbw get` and returns trimmed output.
- `WriteEnvFile` creates a file in `/dev/shm` with `0600` permissions.
- `Clone` with a `git@` URL converts it to HTTPS before cloning.
- After `Clone`, `git remote get-url origin` does NOT contain a token.
- `BuildImage` with no Dockerfile generates one from the base image.
- `BuildImage` passes all `build_args` from config as `--build-arg` flags.
- All operations respect `context.Context` cancellation.

## 5. LOC Estimate

| Step | Estimated LOC |
|------|---------------|
| 2.1 | 25 |
| 2.2 | 20 |
| 2.3 | 30 |
| 2.4 | 40 |
| 2.5 | 70 |
| 2.6 | 90 |
| 2.7 | 125 (tests) |
| **Total** | **~400** |
