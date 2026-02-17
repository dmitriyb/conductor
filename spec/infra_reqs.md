# Infra Module — Requirements

## 1. Purpose

The infra module provides infrastructure services that agents depend on but do
not control: credential retrieval, git repository cloning with authentication,
and Docker image building. It abstracts over multiple secret backends and
ensures credentials never leak to disk or logs.

## 2. Functional Requirements

**FR1 — Credential Store Interface**
Define a `CredentialStore` interface with a single method:
`Get(ctx context.Context, name string) (string, error)`.
All secret lookups go through this interface.

**FR2 — rbw Backend**
Implement `CredentialStore` using `rbw get <name>`. Shell out to the `rbw` CLI.
Return an error if `rbw` is not installed or the vault is locked.

**FR3 — Environment Variable Backend**
Implement `CredentialStore` that reads secrets from environment variables.
The `name` parameter maps directly to the env var name.

**FR4 — File Backend**
Implement `CredentialStore` that reads secrets from files. The `name` parameter
is a file path. The file contents (trimmed of trailing newline) are the secret.

**FR5 — Backend Selection**
Select the backend based on `credentials.backend` in the config (`rbw`, `env`,
`file`). Return an error for unknown backends.

**FR6 — Git Repository Cloning**
Clone a git repository to a temporary directory. Inject the GitHub PAT into the
HTTPS URL for authentication (`https://<token>@github.com/...`). Support both
`https://` and `git@` URL formats by converting `git@` to HTTPS.

**FR7 — Token Cleanup**
After cloning, reconfigure the repo remote to strip the embedded token, replacing
it with the plain HTTPS URL. The token must never persist in `.git/config`.

**FR8 — Docker Image Building**
Build a Docker image from either the config-specified Dockerfile or a generated
one. Pass `build_args` from config as `--build-arg` flags. Tag the image with a
deterministic name derived from the project name.

**FR9 — Dockerfile Generation**
If no custom Dockerfile is specified, generate one from the base image. The
generated Dockerfile installs Claude Code, creates the agent user, sets up
onboarding state, and configures git defaults (matching the current Dockerfile
pattern).

## 3. Non-Functional Requirements

**NFR1 — Secret Hygiene**
Secrets must never appear in log output, error messages, or on-disk files
(except ephemerally in `/dev/shm`). The credential store writes env files
to RAM-backed tmpfs.

**NFR2 — Cleanup on Failure**
If cloning or building fails, all temporary directories and files must be
removed before the error is returned. Use `defer` for cleanup.

**NFR3 — Timeout**
Git clone and Docker build operations must accept a `context.Context` and
respect its deadline. Default timeout: 5 minutes for clone, 10 minutes for
build.

## 4. Interfaces

**Depends on:** config (`Credentials`, `Docker`, `Project` types).

**Consumed by:** agent (needs cloned repo path, env file path, image name),
pipeline (calls `BuildImage` during `conductor build`).

**External dependencies:** `os/exec` (rbw, git, docker CLI), `os` (file I/O),
standard library.
