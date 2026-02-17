# Infra Module — Implementation

## 1. Credential Store Interface

```go
// internal/infra/creds.go

type CredentialStore interface {
    Get(ctx context.Context, name string) (string, error)
}

func NewCredentialStore(backend string) (CredentialStore, error) {
    switch backend {
    case "rbw":  return &rbwStore{}, nil
    case "env":  return &envStore{}, nil
    case "file": return &fileStore{}, nil
    default:     return nil, fmt.Errorf("unknown credential backend: %q", backend)
    }
}
```

## 2. Backends

```go
// internal/infra/creds_rbw.go
type rbwStore struct{}
func (s *rbwStore) Get(ctx context.Context, name string) (string, error) {
    out, err := exec.CommandContext(ctx, "rbw", "get", name).Output()
    if err != nil { return "", fmt.Errorf("rbw get %q: %w", name, err) }
    return strings.TrimRight(string(out), "\n"), nil
}

// internal/infra/creds_env.go
type envStore struct{}
func (s *envStore) Get(_ context.Context, name string) (string, error) {
    val, ok := os.LookupEnv(name)
    if !ok { return "", fmt.Errorf("env var %q not set", name) }
    return val, nil
}

// internal/infra/creds_file.go
type fileStore struct{}
func (s *fileStore) Get(_ context.Context, name string) (string, error) {
    data, err := os.ReadFile(name)
    if err != nil { return "", fmt.Errorf("read secret file %q: %w", name, err) }
    return strings.TrimRight(string(data), "\n"), nil
}
```

## 3. Env File Writer

Writes all secrets to a `chmod 600` file in `/dev/shm` (RAM-backed tmpfs).

```go
type EnvFile struct{ Path string }

func WriteEnvFile(ctx context.Context, store CredentialStore,
    secrets map[string]config.SecretRef) (*EnvFile, error) {
    f, err := os.CreateTemp("/dev/shm", ".conductor-env-")
    if err != nil { return nil, fmt.Errorf("create env file: %w", err) }
    defer f.Close()
    if err := f.Chmod(0600); err != nil { os.Remove(f.Name()); return nil, err }

    for _, ref := range secrets {
        secret, err := store.Get(ctx, ref.Name)
        if err != nil { os.Remove(f.Name()); return nil, err }
        fmt.Fprintf(f, "%s=%s\n", ref.Env, secret)
    }
    return &EnvFile{Path: f.Name()}, nil
}

func (e *EnvFile) Remove() { if e != nil { os.Remove(e.Path) } }
```

## 4. Git Cloner

```go
// internal/infra/git.go

type CloneResult struct {
    Dir      string // tmpdir root
    RepoPath string // Dir + "/repo"
}

func Clone(ctx context.Context, repoURL string,
    store CredentialStore, patSecretName string) (*CloneResult, error) {
    httpsURL := toHTTPS(repoURL)

    token, err := store.Get(ctx, patSecretName)
    if err != nil { return nil, fmt.Errorf("clone: get PAT: %w", err) }

    authedURL := strings.Replace(httpsURL, "https://", "https://"+token+"@", 1)
    dir, err := os.MkdirTemp("", "conductor-")
    if err != nil { return nil, err }
    repoPath := filepath.Join(dir, "repo")

    if out, err := exec.CommandContext(ctx, "git", "clone", "--quiet",
        authedURL, repoPath).CombinedOutput(); err != nil {
        os.RemoveAll(dir)
        return nil, fmt.Errorf("git clone: %w\n%s", err, out)
    }

    // Strip token from remote
    exec.CommandContext(ctx, "git", "-C", repoPath,
        "remote", "set-url", "origin", httpsURL).Run()

    return &CloneResult{Dir: dir, RepoPath: repoPath}, nil
}

func toHTTPS(url string) string {
    if strings.HasPrefix(url, "git@github.com:") {
        return strings.Replace(url, "git@github.com:", "https://github.com/", 1)
    }
    return url
}
```

## 5. Docker Image Builder

```go
// internal/infra/docker.go

func BuildImage(ctx context.Context, cfg *config.Config,
    logger *slog.Logger) (string, error) {
    tag := "conductor-" + cfg.Project.Name
    dockerfile := cfg.Docker.Dockerfile
    if dockerfile == "" {
        var err error
        dockerfile, err = generateDockerfile(cfg.Docker.BaseImage)
        if err != nil { return "", err }
        defer os.Remove(dockerfile)
    }
    args := []string{"build", "--tag", tag, "-f", dockerfile}
    for k, v := range cfg.Docker.BuildArgs {
        args = append(args, "--build-arg", k+"="+v)
    }
    args = append(args, ".")
    cmd := exec.CommandContext(ctx, "docker", args...)
    cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
    if err := cmd.Run(); err != nil { return "", fmt.Errorf("docker build: %w", err) }
    return tag, nil
}
```

## 6. Generated Dockerfile

`generateDockerfile` writes a temp file from the base image. The template
follows the existing Dockerfile pattern: install curl/git/jq/gh, install
Claude Code, create `agent` user, bake onboarding state, set git defaults.

```
FROM <base_image>
RUN apt-get update && apt-get install -y curl git bash jq openssh-client ...
RUN curl -fsSL https://cli.github.com/packages/... && apt-get install -y gh
RUN curl -fsSL https://claude.ai/install.sh | bash
RUN useradd -m -s /bin/bash agent
RUN mkdir -p /home/agent/.claude && echo '{"hasCompletedOnboarding":true}' > ...
USER agent
WORKDIR /workspace
ENTRYPOINT ["/bin/bash", "-c"]
```

Build args injected as `ARG`+`ENV` pairs before the relevant `RUN` instruction.
