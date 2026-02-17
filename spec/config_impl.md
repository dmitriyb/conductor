# Config Module — Implementation

## 1. Struct Definitions

```go
// internal/config/types.go

type Config struct {
    Project     Project              `yaml:"project"`
    Credentials Credentials          `yaml:"credentials"`
    Docker      Docker               `yaml:"docker"`
    Agents      map[string]AgentDef  `yaml:"agents"`
    Pipeline    []StepDef            `yaml:"pipeline"`
}

type Project struct {
    Name       string `yaml:"name"`
    Repository string `yaml:"repository"`
}

type Credentials struct {
    Backend string               `yaml:"backend"` // rbw | env | file
    Secrets map[string]SecretRef `yaml:"secrets"`
}

type SecretRef struct {
    Name string `yaml:"name"` // backend-specific key
    Env  string `yaml:"env"`  // env var name to expose in container
}

type Docker struct {
    BaseImage  string            `yaml:"base_image"`
    Dockerfile string            `yaml:"dockerfile"`
    BuildArgs  map[string]string `yaml:"build_args"`
}

type AgentDef struct {
    Prompt       PromptDef      `yaml:"prompt"`
    Workspace    string         `yaml:"workspace"` // rw | ro
    OutputSchema map[string]any `yaml:"output_schema"`
    Tools        []string       `yaml:"tools"`
}

type PromptDef struct {
    System string `yaml:"system"` // file path or inline text
    Task   string `yaml:"task"`   // Go text/template
}

type StepDef struct {
    Name      string   `yaml:"name"`
    Agent     string   `yaml:"agent"`
    DependsOn []string `yaml:"depends_on"`
    Condition string   `yaml:"condition"` // CEL expression, empty = always
}
```

## 2. Loading

```go
// internal/config/load.go

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("config: read %s: %w", path, err)
    }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("config: parse %s: %w", path, err)
    }
    return &cfg, nil
}
```

Default path: `"orchestrator.yaml"` relative to cwd when `--config` is empty.

## 3. Validation

Collects all errors into a slice with field paths. Uses `errors.Join`.

```go
// internal/config/validate.go

func Validate(cfg *Config) error {
    var errs []error
    check := func(cond bool, path, msg string) {
        if !cond {
            errs = append(errs, fmt.Errorf("%s: %s", path, msg))
        }
    }

    check(cfg.Project.Name != "", "project.name", "required")
    check(cfg.Project.Repository != "", "project.repository", "required")

    validBackends := map[string]bool{"rbw": true, "env": true, "file": true}
    check(validBackends[cfg.Credentials.Backend], "credentials.backend",
        fmt.Sprintf("must be one of: rbw, env, file (got %q)", cfg.Credentials.Backend))
    check(cfg.Docker.BaseImage != "", "docker.base_image", "required")
    check(len(cfg.Agents) > 0, "agents", "at least one agent must be defined")

    for name, agent := range cfg.Agents {
        p := "agents." + name
        check(agent.Prompt.System != "", p+".prompt.system", "required")
        check(agent.Prompt.Task != "", p+".prompt.task", "required")
        check(agent.Workspace == "rw" || agent.Workspace == "ro",
            p+".workspace", fmt.Sprintf("must be rw or ro (got %q)", agent.Workspace))
    }

    check(len(cfg.Pipeline) > 0, "pipeline", "at least one step required")
    stepNames := map[string]bool{}
    for i, step := range cfg.Pipeline {
        p := fmt.Sprintf("pipeline[%d]", i)
        check(step.Name != "", p+".name", "required")
        if step.Agent != "" {
            _, ok := cfg.Agents[step.Agent]
            check(ok, p+".agent", fmt.Sprintf("references undefined agent %q", step.Agent))
        }
        for _, dep := range step.DependsOn {
            check(stepNames[dep], p+".depends_on", fmt.Sprintf("unknown step %q", dep))
        }
        stepNames[step.Name] = true
    }
    return errors.Join(errs...)
}
```

## 4. Logging

```go
// internal/config/logging.go

func InitLogging(level string) *slog.Logger {
    var lvl slog.Level
    switch strings.ToLower(level) {
    case "debug":  lvl = slog.LevelDebug
    case "warn":   lvl = slog.LevelWarn
    case "error":  lvl = slog.LevelError
    default:       lvl = slog.LevelInfo
    }
    opts := &slog.HandlerOptions{Level: lvl}
    var h slog.Handler
    if term.IsTerminal(int(os.Stderr.Fd())) {
        h = slog.NewTextHandler(os.Stderr, opts)
    } else {
        h = slog.NewJSONHandler(os.Stderr, opts)
    }
    return slog.New(h)
}
```

## 5. CLI Dispatch

```go
// main.go (sketch)

func main() {
    cfgPath := flag.String("config", "orchestrator.yaml", "config file path")
    logLevel := flag.String("log-level", "info", "log level")
    flag.Parse()
    logger := config.InitLogging(*logLevel)

    cfg, err := config.Load(*cfgPath)
    // handle err...

    switch flag.Args()[0] {
    case "validate":  err = config.Validate(cfg)
    case "build":     err = infra.BuildImage(cfg, logger)
    case "run":       err = pipeline.Execute(cfg, logger)
    }
}
```

## 6. Example orchestrator.yaml

```yaml
project:
  name: differentia
  repository: https://github.com/dmitriyb/differentia.git

credentials:
  backend: rbw
  secrets:
    claude_token: { name: claude-oauth-token, env: CLAUDE_CODE_OAUTH_TOKEN }
    github_pat:   { name: differentia-pat,    env: AGENT_GH_TOKEN }

docker:
  base_image: debian:bookworm-slim
  build_args: { ZIG_VERSION: "0.14.1" }

agents:
  implementer:
    prompt:
      system: roles/IMPLEMENTING.md
      task: "Implement GitHub issue #{{.IssueNumber}} from {{.RepoURL}}."
    workspace: rw
    output_schema: { pr_number: int, branch: string, status: string }
  reviewer:
    prompt:
      system: roles/REVIEWING.md
      task: "Review PR #{{.PRNumber}} in {{.RepoOwner}}/{{.RepoName}}."
    workspace: ro
    output_schema: { status: string, comment_count: int, summary: string }

pipeline:
  - { name: implement, agent: implementer }
  - { name: review, agent: reviewer, depends_on: [implement] }
  - { name: fix, agent: fixer, depends_on: [review],
      condition: "steps.review.output.status == 'changes_requested'" }
```
